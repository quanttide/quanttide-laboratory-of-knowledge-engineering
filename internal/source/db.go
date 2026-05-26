package source

import (
	"database/sql"
	"fmt"
	"time"
)

type DBSource struct {
	name   string
	db     *sql.DB
	query  string
	column string
}

func NewDBSource(name string, db *sql.DB, query, column string) *DBSource {
	return &DBSource{
		name:   name,
		db:     db,
		query:  query,
		column: column,
	}
}

func (s *DBSource) Name() string { return s.name }

func (s *DBSource) Fetch() ([]Record, error) {
	if s.db == nil {
		return nil, fmt.Errorf("db source %q: database not connected", s.name)
	}

	rows, err := s.db.Query(s.query)
	if err != nil {
		return nil, fmt.Errorf("db source %q query: %w", s.name, err)
	}
	defer rows.Close()

	var records []Record
	index := 0
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			continue
		}
		records = append(records, Record{
			ID:      fmt.Sprintf("%s-%d", s.name, index),
			Source:  s.name,
			Content: content,
			Metadata: map[string]string{
				"db_column": s.column,
			},
			Timestamp: time.Now(),
		})
		index++
	}

	return records, rows.Err()
}
