package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractorBasic(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	out := filepath.Join(dir, "out")
	os.MkdirAll(in, 0755)
	os.WriteFile(filepath.Join(in, "doc.md"), []byte("---\ntitle: Test\n---\n\n# Hello"), 0644)
	os.WriteFile(filepath.Join(in, "plain.txt"), []byte("plain text"), 0644)
	os.WriteFile(filepath.Join(in, "skip.json"), []byte("{}"), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", in, "-output", out}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(out)
	if len(entries) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(entries))
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			t.Errorf("expected .json files, got %s", e.Name())
		}
	}
}

func TestExtractorSingleFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	in := filepath.Join(dir, "single.md")
	os.WriteFile(in, []byte("# Just a doc"), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", in, "-output", out}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(out)
	if len(entries) != 1 {
		t.Fatalf("expected 1 output file, got %d", len(entries))
	}
}

func TestExtractorWithFrontmatter(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	in := filepath.Join(dir, "fm.md")
	os.WriteFile(in, []byte("---\nkey: val\n---\n\nbody"), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", in, "-output", out}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(out)
	if len(entries) != 1 {
		t.Fatalf("expected 1 output, got %d", len(entries))
	}
}

func TestExtractorEmptyDir(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", dir, "-output", out}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(out)
	if len(entries) != 0 {
		t.Errorf("expected 0 outputs for empty dir, got %d", len(entries))
	}
}

func TestExtractorNoInput(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{}); err == nil {
		t.Error("expected error for missing -input")
	}
}

func TestExtractorBadInput(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-input", "/nonexistent", "-output", out}); err == nil {
		t.Error("expected error for bad input")
	}
}
