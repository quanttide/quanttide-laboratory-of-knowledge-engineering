package source

import "fmt"

type Federator struct {
	sources []Source
}

func NewFederator(sources ...Source) *Federator {
	return &Federator{sources: sources}
}

func (f *Federator) AddSource(s Source) {
	f.sources = append(f.sources, s)
}

func (f *Federator) FetchAll() (map[string][]Record, error) {
	results := make(map[string][]Record)
	var errs []error

	for _, s := range f.sources {
		records, err := s.Fetch()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", s.Name(), err))
			continue
		}
		results[s.Name()] = records
	}

	if len(errs) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all sources failed: %v", errs)
	}

	return results, nil
}

func (f *Federator) SourceCount() int {
	return len(f.sources)
}
