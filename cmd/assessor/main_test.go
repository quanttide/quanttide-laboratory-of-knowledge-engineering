package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAssessorBasic(t *testing.T) {
	dir := t.TempDir()
	parsedDir := filepath.Join(dir, "parsed")
	os.MkdirAll(parsedDir, 0755)
	docJSON := `{"Path":"doc.md","Frontmatter":{"title":"Test"},"Body":"Hello world knowledge engineering"}`
	os.WriteFile(filepath.Join(parsedDir, "doc.md.json"), []byte(docJSON), 0644)

	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"triples\":[]}"}}]}`))
	}))
	defer mockLLM.Close()

	output := filepath.Join(dir, "triples.jsonl")
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer os.Unsetenv("OPENAI_API_KEY")

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", parsedDir, "-output", output, "-api-url", mockLLM.URL, "-model", "mock", "-max-fails", "3"}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(output); err != nil {
		t.Errorf("expected triples.jsonl: %v", err)
	}
}

func TestAssessorWithTriples(t *testing.T) {
	dir := t.TempDir()
	parsedDir := filepath.Join(dir, "parsed")
	os.MkdirAll(parsedDir, 0755)
	docJSON := `{"Path":"doc.md","Frontmatter":{},"Body":"订单包含商品"}`
	os.WriteFile(filepath.Join(parsedDir, "order.md.json"), []byte(docJSON), 0644)

	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"triples\":[{\"subject\":\"订单\",\"predicate\":\"包含\",\"object\":\"商品\",\"confidence\":0.9,\"sentence\":\"订单包含商品\"}]}"}}]}`))
	}))
	defer mockLLM.Close()

	output := filepath.Join(dir, "triples.jsonl")
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer os.Unsetenv("OPENAI_API_KEY")

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", parsedDir, "-output", output, "-api-url", mockLLM.URL, "-model", "mock"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("expected output: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected triples content")
	}
}

func TestAssessorEmptyDir(t *testing.T) {
	dir := t.TempDir()
	parsedDir := filepath.Join(dir, "parsed")
	os.MkdirAll(parsedDir, 0755)
	output := filepath.Join(dir, "triples.jsonl")

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", parsedDir, "-output", output, "-api-url", "http://localhost:1", "-model", "mock"}); err != nil {
		t.Fatal(err)
	}
}

func TestAssessorNoAPIKey(t *testing.T) {
	dir := t.TempDir()
	parsedDir := filepath.Join(dir, "parsed")
	os.MkdirAll(parsedDir, 0755)
	output := filepath.Join(dir, "triples.jsonl")

	os.Unsetenv("OPENAI_API_KEY")
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", parsedDir, "-output", output, "-api-url", "http://localhost:1", "-model", "mock"}); err != nil {
		t.Fatal(err)
	}
}
