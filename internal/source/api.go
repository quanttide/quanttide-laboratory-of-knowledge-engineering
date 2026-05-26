package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APISource struct {
	name    string
	url     string
	client  *http.Client
	path    string
}

func NewAPISource(name, url string) *APISource {
	return &APISource{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		path: "data",
	}
}

func (s *APISource) Name() string { return s.name }

func (s *APISource) Fetch() ([]Record, error) {
	resp, err := s.client.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("api source %q: %w", s.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api source %q: HTTP %d", s.name, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api source %q read: %w", s.name, err)
	}

	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("api source %q unmarshal: %w", s.name, err)
	}

	items := extractItems(raw, s.path)

	var records []Record
	for i, item := range items {
		text, ok := item.(string)
		if !ok {
			data, _ := json.Marshal(item)
			text = string(data)
		}
		records = append(records, Record{
			ID:      fmt.Sprintf("%s-%d", s.name, i),
			Source:  s.name,
			Content: text,
			Metadata: map[string]string{
				"api_url": s.url,
			},
			Timestamp: time.Now(),
		})
	}

	if len(records) == 0 {
		records = append(records, Record{
			ID:      s.name + "-0",
			Source:  s.name,
			Content: string(body),
			Metadata: map[string]string{
				"api_url": s.url,
			},
			Timestamp: time.Now(),
		})
	}

	return records, nil
}

func extractItems(raw any, path string) []any {
	switch v := raw.(type) {
	case []any:
		return v
	case map[string]any:
		if val, ok := v[path]; ok {
			if items, ok := val.([]any); ok {
				return items
			}
		}
		for _, val := range v {
			if items, ok := val.([]any); ok {
				return items
			}
		}
	}
	return nil
}
