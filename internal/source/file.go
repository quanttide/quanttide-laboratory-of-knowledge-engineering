package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileSource struct {
	name string
	dir  string
	exts []string
}

func NewFileSource(name, dir string, exts []string) *FileSource {
	if len(exts) == 0 {
		exts = []string{".md", ".txt"}
	}
	return &FileSource{
		name: name,
		dir:  dir,
		exts: exts,
	}
}

func (s *FileSource) Name() string { return s.name }

func (s *FileSource) Fetch() ([]Record, error) {
	info, err := os.Stat(s.dir)
	if err != nil {
		return nil, fmt.Errorf("file source: %w", err)
	}

	var paths []string
	if info.IsDir() {
		if err := filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			for _, allowed := range s.exts {
				if ext == allowed {
					paths = append(paths, path)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("walk dir: %w", err)
		}
	} else {
		paths = append(paths, s.dir)
	}

	var records []Record
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		records = append(records, Record{
			ID:      path,
			Source:  s.name,
			Content: string(data),
			Metadata: map[string]string{
				"path": path,
				"size": fmt.Sprintf("%d", len(data)),
			},
			Timestamp: time.Now(),
		})
	}

	return records, nil
}
