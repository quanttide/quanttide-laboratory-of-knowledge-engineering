package source

import (
	"fmt"
	"sort"
	"sync"
)

type FusionStrategy interface {
	Name() string
	Fuse(sources map[string][]Record) []FusedTriple
}

type FusedTriple struct {
	Subject       string
	Predicate     string
	Object        string
	Confidence    float64
	SourceWeights map[string]float64
	Sources       []string
}

type MaxConfidenceFusion struct{}

func (m *MaxConfidenceFusion) Name() string { return "max_confidence" }

func (m *MaxConfidenceFusion) Fuse(sources map[string][]Record) []FusedTriple {
	keyMap := make(map[string]*FusedTriple)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for srcName, records := range sources {
		wg.Add(1)
		go func(name string, recs []Record) {
			defer wg.Done()
			for _, rec := range recs {
				triples := extractCandidateTriples(rec)
				mu.Lock()
				for _, t := range triples {
					key := t.Subject + "|" + t.Predicate + "|" + t.Object
					existing, ok := keyMap[key]
					if !ok {
						keyMap[key] = &FusedTriple{
							Subject:       t.Subject,
							Predicate:     t.Predicate,
							Object:        t.Object,
							Confidence:    t.Confidence,
							SourceWeights: map[string]float64{name: t.Confidence},
							Sources:       []string{name},
						}
					} else {
						if t.Confidence > existing.Confidence {
							existing.Confidence = t.Confidence
						}
						existing.SourceWeights[name] = t.Confidence
						if !contains(existing.Sources, name) {
							existing.Sources = append(existing.Sources, name)
						}
					}
				}
				mu.Unlock()
			}
		}(srcName, records)
	}
	wg.Wait()

	result := make([]FusedTriple, 0, len(keyMap))
	for _, t := range keyMap {
		result = append(result, *t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})
	return result
}

type WeightedAverageFusion struct {
	SourceWeights map[string]float64
}

func (w *WeightedAverageFusion) Name() string { return "weighted_average" }

func (w *WeightedAverageFusion) Fuse(sources map[string][]Record) []FusedTriple {
	base := &MaxConfidenceFusion{}
	triples := base.Fuse(sources)

	for i, t := range triples {
		var totalWeight float64
		var weightedSum float64
		for src, conf := range t.SourceWeights {
			weight := w.SourceWeights[src]
			if weight <= 0 {
				weight = 1.0
			}
			weightedSum += conf * weight
			totalWeight += weight
		}
		if totalWeight > 0 {
			triples[i].Confidence = weightedSum / totalWeight
		}
	}

	return triples
}

type candidateTriple struct {
	Subject    string
	Predicate  string
	Object     string
	Confidence float64
}

func extractCandidateTriples(rec Record) []candidateTriple {
	return []candidateTriple{
		{
			Subject:    rec.Source,
			Predicate:  "source",
			Object:     rec.ID,
			Confidence: 1.0,
		},
	}
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func NewFusionStrategy(name string, weights map[string]float64) (FusionStrategy, error) {
	switch name {
	case "max_confidence":
		return &MaxConfidenceFusion{}, nil
	case "weighted_average":
		return &WeightedAverageFusion{SourceWeights: weights}, nil
	default:
		return nil, fmt.Errorf("unknown fusion strategy %q", name)
	}
}
