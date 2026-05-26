package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderEmptyInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(input, []byte(""), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", input, "-threshold", "0.7"}); err != nil {
		t.Fatal(err)
	}
}

func TestLoaderWithTriples(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "triples.jsonl")
	content := `{"id":"t1","source":"doc.md","confidence":0.9,"subject":"用户","predicate":"下单","object":"订单","verification":"unverified"}
{"id":"t2","source":"doc.md","confidence":0.5,"subject":"订单","predicate":"包含","object":"商品","verification":"unverified"}
{"id":"t3","source":"doc.md","confidence":0.2,"subject":"商品","predicate":"有","object":"库存","verification":"unverified"}`
	os.WriteFile(input, []byte(content), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", input, "-threshold", "0.7"}); err != nil {
		t.Fatal(err)
	}

	storePath := filepath.Join(dir, "store.jsonl")
	if _, err := os.Stat(storePath); err == nil {
		data, _ := os.ReadFile(storePath)
		if len(data) == 0 {
			t.Error("expected non-empty store.jsonl")
		}
	}
}

func TestLoaderCustomThreshold(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "triples.jsonl")
	content := `{"id":"t1","source":"doc.md","confidence":0.6,"subject":"A","predicate":"B","object":"C","verification":"unverified"}`
	os.WriteFile(input, []byte(content), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", input, "-threshold", "0.5"}); err != nil {
		t.Fatal(err)
	}
}

func TestLoaderInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "bad.jsonl")
	os.WriteFile(input, []byte("not json\n"), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", input, "-threshold", "0.7"}); err != nil {
		t.Fatal(err)
	}
}
