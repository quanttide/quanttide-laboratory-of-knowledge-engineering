package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratorCode(t *testing.T) {
	dir := t.TempDir()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-target", "code", "-query", "contains(?X, ?Y)", "-output", dir}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Error("expected generated files")
	}
}

func TestGeneratorAPI(t *testing.T) {
	dir := t.TempDir()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-target", "api", "-query", "contains(?X, ?Y)", "-output", dir}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Error("expected generated files")
	}
}

func TestGeneratorFAQ(t *testing.T) {
	dir := t.TempDir()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-target", "faq", "-query", "contains(?X, ?Y)", "-output", dir}); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Error("expected generated files")
	}
}

func TestGeneratorUnknownTarget(t *testing.T) {
	dir := t.TempDir()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-target", "unknown", "-query", "contains(?X, ?Y)", "-output", dir}); err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestGeneratorOutputToFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.go")
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-target", "code", "-query", "contains(?X, ?Y)", "-output", out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected output file: %v", err)
	}
}
