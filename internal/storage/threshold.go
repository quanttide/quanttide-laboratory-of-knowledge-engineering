package storage

import "github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"

const (
	DefaultHighThreshold   = 0.7
	DefaultMediumThreshold = 0.4
)

type ThresholdPolicy struct {
	High   float64
	Medium float64
}

func DefaultThresholdPolicy() ThresholdPolicy {
	return ThresholdPolicy{
		High:   DefaultHighThreshold,
		Medium: DefaultMediumThreshold,
	}
}

type ThresholdResult struct {
	AutoInsert []*triple.Triple
	Pending    []*triple.Triple
	Discarded  []*triple.Triple
}

func (p ThresholdPolicy) Classify(triples []*triple.Triple) ThresholdResult {
	var result ThresholdResult
	for _, t := range triples {
		switch {
		case t.Confidence >= p.High:
			result.AutoInsert = append(result.AutoInsert, t)
		case t.Confidence >= p.Medium:
			result.Pending = append(result.Pending, t)
		default:
			result.Discarded = append(result.Discarded, t)
		}
	}
	return result
}
