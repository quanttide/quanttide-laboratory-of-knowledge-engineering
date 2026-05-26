package source

import "time"

type Record struct {
	ID        string
	Source    string
	Content   string
	Metadata  map[string]string
	Timestamp time.Time
}

type Source interface {
	Name() string
	Fetch() ([]Record, error)
}
