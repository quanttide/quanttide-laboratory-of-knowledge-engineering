package storage

import (
	"os"
	"testing"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

func TestThresholdClassify(t *testing.T) {
	policy := ThresholdPolicy{High: 0.7, Medium: 0.4}

	triples := []*triple.Triple{
		triple.New("t1", "src", 0.9, "A", "B", "C"),
		triple.New("t2", "src", 0.5, "D", "E", "F"),
		triple.New("t3", "src", 0.2, "G", "H", "I"),
	}

	result := policy.Classify(triples)
	if len(result.AutoInsert) != 1 {
		t.Errorf("expected 1 auto-insert, got %d", len(result.AutoInsert))
	}
	if len(result.Pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(result.Pending))
	}
	if len(result.Discarded) != 1 {
		t.Errorf("expected 1 discarded, got %d", len(result.Discarded))
	}
}

func TestThresholdCustom(t *testing.T) {
	policy := ThresholdPolicy{High: 0.5, Medium: 0.2}

	triples := []*triple.Triple{
		triple.New("t1", "src", 0.6, "A", "B", "C"),
		triple.New("t2", "src", 0.3, "D", "E", "F"),
	}

	result := policy.Classify(triples)
	if len(result.AutoInsert) != 1 {
		t.Errorf("expected 1 auto-insert, got %d", len(result.AutoInsert))
	}
	if len(result.Pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(result.Pending))
	}
}

func TestJSONStoreInsertAndQuery(t *testing.T) {
	dir, err := os.MkdirTemp("", "jsonstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)

	triples := []*triple.Triple{
		triple.New("t1", "doc1.md", 0.9, "用户", "下单", "订单"),
		triple.New("t2", "doc2.md", 0.8, "订单", "包含", "商品"),
	}

	inserted, err := s.InsertTriples(triples)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if inserted != 2 {
		t.Errorf("expected 2 inserted, got %d", inserted)
	}

	results, err := s.Query(QueryPattern{Subject: "用户"})
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Predicate != "下单" {
		t.Errorf("expected predicate '下单', got %s", results[0].Predicate)
	}
}

func TestJSONStoreMergeConflict(t *testing.T) {
	dir, err := os.MkdirTemp("", "jsonstore-merge")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)

	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src1", 0.7, "A", "B", "C"),
	})

	s.InsertTriples([]*triple.Triple{
		triple.New("t2", "src2", 0.9, "A", "B", "C"),
	})

	results, err := s.Query(QueryPattern{})
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (merged), got %d", len(results))
	}
	if results[0].Confidence != 0.9 {
		t.Errorf("expected confidence 0.9 (max merged), got %f", results[0].Confidence)
	}
}

func TestJSONStorePersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "jsonstore-persist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s1 := NewJSONStore(dir)
	s1.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 0.95, "X", "Y", "Z"),
	})
	s1.Close()

	s2 := NewJSONStore(dir)
	results, err := s2.Query(QueryPattern{})
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 persisted result, got %d", len(results))
	}
}

func TestJSONStoreEmptyQuery(t *testing.T) {
	dir, err := os.MkdirTemp("", "jsonstore-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	results, err := s.Query(QueryPattern{})
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDefaultThresholdPolicy(t *testing.T) {
	p := DefaultThresholdPolicy()
	if p.High != 0.7 {
		t.Errorf("expected High=0.7, got %f", p.High)
	}
	if p.Medium != 0.4 {
		t.Errorf("expected Medium=0.4, got %f", p.Medium)
	}
}

func TestQueryNeighbors(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-neighbor")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 0.9, "A", "knows", "B"),
		triple.New("t2", "src", 0.8, "A", "likes", "C"),
		triple.New("t3", "src", 0.7, "B", "knows", "A"),
	})

	neighbors, err := s.QueryNeighbors("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(neighbors) != 3 {
		t.Fatalf("expected 3 neighbors for A, got %d", len(neighbors))
	}

	hasOut := false
	hasIn := false
	for _, n := range neighbors {
		if n.Direction == "out" && n.Node == "B" && n.Predicate == "knows" {
			hasOut = true
		}
		if n.Direction == "in" && n.Node == "B" && n.Predicate == "knows" {
			hasIn = true
		}
	}
	if !hasOut {
		t.Error("expected A -knows-> B as outgoing")
	}
	if !hasIn {
		t.Error("expected B -knows-> A as incoming")
	}
}

func TestQueryByPredicate(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-pred")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 0.9, "A", "knows", "B"),
		triple.New("t2", "src", 0.8, "A", "likes", "C"),
		triple.New("t3", "src", 0.7, "B", "knows", "C"),
	})

	results, err := s.QueryByPredicate("knows")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 'knows' edges, got %d", len(results))
	}
}

func TestQueryPath(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-path")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 1.0, "A", "knows", "B"),
		triple.New("t2", "src", 1.0, "B", "knows", "C"),
		triple.New("t3", "src", 1.0, "C", "knows", "D"),
	})

	paths, err := s.QueryPath("A", "D", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("expected at least 1 path from A to D")
	}
	if len(paths[0]) != 3 {
		t.Errorf("expected path length 3 (A→B→C→D), got %d", len(paths[0]))
	}
}

func TestQueryPathNoPath(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-nopath")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 1.0, "A", "knows", "B"),
		triple.New("t2", "src", 1.0, "C", "knows", "D"),
	})

	paths, err := s.QueryPath("A", "D", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestQueryPathSelf(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-self")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 1.0, "A", "knows", "B"),
	})

	paths, err := s.QueryPath("A", "A", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for self-query, got %d", len(paths))
	}
}

func TestQueryNeighborsUnknownNode(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-unknown")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	s.InsertTriples([]*triple.Triple{
		triple.New("t1", "src", 0.9, "A", "knows", "B"),
	})

	neighbors, err := s.QueryNeighbors("Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(neighbors) != 0 {
		t.Errorf("expected 0 neighbors for unknown node, got %d", len(neighbors))
	}
}

func TestQueryByPredicateEmpty(t *testing.T) {
	dir, _ := os.MkdirTemp("", "graph-pred-empty")
	defer os.RemoveAll(dir)

	s := NewJSONStore(dir)
	results, err := s.QueryByPredicate("anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty store, got %d", len(results))
	}
}
