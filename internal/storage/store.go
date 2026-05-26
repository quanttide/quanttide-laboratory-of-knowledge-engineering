package storage

import "github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"

type TripleRecord struct {
	Subject    string
	Predicate  string
	Object     string
	Confidence float64
	Source     string
}

type Neighbor struct {
	Node       string
	Predicate  string
	Direction  string
	Confidence float64
}

type Store interface {
	InsertTriples(triples []*triple.Triple) (int, error)
	Query(pattern QueryPattern) ([]TripleRecord, error)
	QueryNeighbors(node string) ([]Neighbor, error)
	QueryPath(from, to string, maxDepth int) ([][]TripleRecord, error)
	QueryByPredicate(predicate string) ([]TripleRecord, error)
	Close() error
}

type QueryPattern struct {
	Subject   string
	Predicate string
	Object    string
	Limit     int
}
