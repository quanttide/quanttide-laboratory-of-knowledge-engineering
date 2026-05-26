package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

type JSONStore struct {
	dir    string
	mu     sync.RWMutex
	rows   []TripleRecord
}

func NewJSONStore(dir string) *JSONStore {
	s := &JSONStore{
		dir:  dir,
		rows: make([]TripleRecord, 0),
	}
	s.load()
	return s
}

func (s *JSONStore) InsertTriples(triples []*triple.Triple) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inserted := 0
	for _, t := range triples {
		existing := s.find(t.Subject, t.Predicate, t.Object)
		if existing != nil {
			if t.Confidence > existing.Confidence {
				existing.Confidence = t.Confidence
				existing.Source = t.Source
			}
		} else {
			s.rows = append(s.rows, TripleRecord{
				Subject:    t.Subject,
				Predicate:  t.Predicate,
				Object:     t.Object,
				Confidence: t.Confidence,
				Source:     t.Source,
			})
			inserted++
		}
	}

	if err := s.flush(); err != nil {
		return inserted, fmt.Errorf("flush error: %w", err)
	}
	return inserted, nil
}

func (s *JSONStore) Query(pattern QueryPattern) ([]TripleRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []TripleRecord
	for _, r := range s.rows {
		if pattern.Subject != "" && r.Subject != pattern.Subject {
			continue
		}
		if pattern.Predicate != "" && r.Predicate != pattern.Predicate {
			continue
		}
		if pattern.Object != "" && r.Object != pattern.Object {
			continue
		}
		results = append(results, r)
		if pattern.Limit > 0 && len(results) >= pattern.Limit {
			break
		}
	}
	return results, nil
}

func (s *JSONStore) QueryNeighbors(node string) ([]Neighbor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var neighbors []Neighbor
	seen := map[string]bool{}
	for _, r := range s.rows {
		if r.Subject == node {
			key := "out:" + r.Predicate + ":" + r.Object
			if !seen[key] {
				seen[key] = true
				neighbors = append(neighbors, Neighbor{
					Node:       r.Object,
					Predicate:  r.Predicate,
					Direction:  "out",
					Confidence: r.Confidence,
				})
			}
		}
		if r.Object == node {
			key := "in:" + r.Predicate + ":" + r.Subject
			if !seen[key] {
				seen[key] = true
				neighbors = append(neighbors, Neighbor{
					Node:       r.Subject,
					Predicate:  r.Predicate,
					Direction:  "in",
					Confidence: r.Confidence,
				})
			}
		}
	}
	return neighbors, nil
}

func (s *JSONStore) QueryByPredicate(predicate string) ([]TripleRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []TripleRecord
	for _, r := range s.rows {
		if r.Predicate == predicate {
			results = append(results, r)
		}
	}
	return results, nil
}

func (s *JSONStore) QueryPath(from, to string, maxDepth int) ([][]TripleRecord, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var paths [][]TripleRecord
	visited := map[string]bool{}

	var dfs func(current string, path []TripleRecord)
	dfs = func(current string, path []TripleRecord) {
		if len(path) > maxDepth {
			return
		}
		if current == to && len(path) > 0 {
			p := make([]TripleRecord, len(path))
			copy(p, path)
			paths = append(paths, p)
			return
		}
		if visited[current] {
			return
		}
		visited[current] = true
		for _, r := range s.rows {
			if r.Subject == current {
				dfs(r.Object, append(path, r))
			}
		}
		visited[current] = false
	}

	if from != to {
		dfs(from, nil)
	}

	return paths, nil
}

func (s *JSONStore) Close() error {
	return s.flush()
}

func (s *JSONStore) find(subject, predicate, object string) *TripleRecord {
	for i := range s.rows {
		r := &s.rows[i]
		if r.Subject == subject && r.Predicate == predicate && r.Object == object {
			return r
		}
	}
	return nil
}

func (s *JSONStore) load() {
	path := filepath.Join(s.dir, "store.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	for _, line := range lines {
		var r TripleRecord
		if err := json.Unmarshal([]byte(line), &r); err == nil {
			s.rows = append(s.rows, r)
		}
	}
}

func (s *JSONStore) flush() error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(s.dir, "store.jsonl")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range s.rows {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
