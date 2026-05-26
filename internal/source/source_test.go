package source

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileSource(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# Doc A"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("Doc B"), 0644)
	os.WriteFile(filepath.Join(dir, "c.json"), []byte("{}"), 0644)

	s := NewFileSource("test", dir, []string{".md", ".txt"})
	records, err := s.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records (.md + .txt), got %d", len(records))
	}
}

func TestFileSourceSingleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "single.md")
	os.WriteFile(path, []byte("# Single"), 0644)

	s := NewFileSource("single", path, nil)
	records, err := s.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestFileSourceNotFound(t *testing.T) {
	s := NewFileSource("missing", "/nonexistent", nil)
	_, err := s.Fetch()
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestFederator(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# Federated A"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("# Federated B"), 0644)

	fs := NewFileSource("files", dir, nil)
	f := NewFederator(fs)

	results, err := f.FetchAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 source result, got %d", len(results))
	}
	records, ok := results["files"]
	if !ok {
		t.Fatal("expected 'files' source key")
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
	if f.SourceCount() != 1 {
		t.Errorf("expected 1 source, got %d", f.SourceCount())
	}
}

func TestFederatorAddSource(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir1, "x.md"), []byte("X"), 0644)
	os.WriteFile(filepath.Join(dir2, "y.md"), []byte("Y"), 0644)

	f := NewFederator()
	f.AddSource(NewFileSource("src1", dir1, nil))
	f.AddSource(NewFileSource("src2", dir2, nil))

	results, err := f.FetchAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 sources, got %d", len(results))
	}
}

func TestFederatorAllFailed(t *testing.T) {
	f := NewFederator(
		NewFileSource("bad1", "/dev/null/nope1", nil),
		NewFileSource("bad2", "/dev/null/nope2", nil),
	)
	_, err := f.FetchAll()
	if err == nil {
		t.Fatal("expected error when all sources fail")
	}
}

func TestMaxConfidenceFusion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Fusion"), 0644)

	fs := NewFileSource("files", dir, nil)
	f := NewFederator(fs)

	sources, _ := f.FetchAll()
	strategy := &MaxConfidenceFusion{}
	results := strategy.Fuse(sources)

	if len(results) == 0 {
		t.Fatal("expected fusion results")
	}
}

func TestFusionStrategyLookup(t *testing.T) {
	s, err := NewFusionStrategy("max_confidence", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name() != "max_confidence" {
		t.Errorf("unexpected name: %s", s.Name())
	}

	_, err = NewFusionStrategy("unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func TestWeightedAverageFusion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "d.md"), []byte("# Weighted"), 0644)

	fs := NewFileSource("files", dir, nil)
	f := NewFederator(fs)

	sources, _ := f.FetchAll()
	strategy := &WeightedAverageFusion{
		SourceWeights: map[string]float64{"files": 2.0},
	}
	results := strategy.Fuse(sources)

	if len(results) == 0 {
		t.Fatal("expected fusion results")
	}
}

func TestAPISourceSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`))
	}))
	defer server.Close()

	s := NewAPISource("users", server.URL)
	records, err := s.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestAPISourceObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"x":1}]}`))
	}))
	defer server.Close()

	s := NewAPISource("obj", server.URL)
	records, err := s.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record from object response, got %d", len(records))
	}
}

func TestAPISourceHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewAPISource("err", server.URL)
	_, err := s.Fetch()
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestAPISourceInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	s := NewAPISource("bad", server.URL)
	_, err := s.Fetch()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExtractItems(t *testing.T) {
	items := extractItems([]any{"a", "b"}, "data")
	if len(items) != 2 {
		t.Errorf("expected 2 items from array, got %d", len(items))
	}

	items = extractItems(map[string]any{"data": []any{1, 2, 3}}, "data")
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	items = extractItems(map[string]any{"x": 1}, "data")
	if items != nil {
		t.Errorf("expected nil for missing key, got %v", items)
	}
}

func TestDBSourceNoConnection(t *testing.T) {
	s := NewDBSource("test", nil, "SELECT content FROM docs", "content")
	_, err := s.Fetch()
	if err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestNewFusionStrategyWeightedAverage(t *testing.T) {
	s, err := NewFusionStrategy("weighted_average", map[string]float64{"a": 1.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name() != "weighted_average" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected contains to find 'b'")
	}
	if contains([]string{"a", "b"}, "z") {
		t.Error("expected contains to not find 'z'")
	}
	if contains([]string{}, "x") {
		t.Error("expected contains to return false for empty slice")
	}
}
