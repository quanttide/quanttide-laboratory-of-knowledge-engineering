package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestFederateFileSource(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "fused.json")
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "a.md"), []byte("# Doc A"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.md"), []byte("# Doc B"), 0644)

	sourceDef := "files=file:" + srcDir
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-sources", sourceDef, "-output", out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected fused.json: %v", err)
	}
}

func TestFederateMultipleSources(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "fused.json")
	src1 := filepath.Join(dir, "src1")
	src2 := filepath.Join(dir, "src2")
	os.MkdirAll(src1, 0755)
	os.MkdirAll(src2, 0755)
	os.WriteFile(filepath.Join(src1, "x.md"), []byte("X"), 0644)
	os.WriteFile(filepath.Join(src2, "y.md"), []byte("Y"), 0644)

	sourceDef := "a=file:" + src1 + ",b=file:" + src2
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-sources", sourceDef, "-output", out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected fused.json: %v", err)
	}
}

func TestFederateNoSources(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.json")
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-sources", "", "-output", out}); err == nil {
		t.Error("expected error for empty sources")
	}
}

func TestFederateUnknownSourceType(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.json")
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-sources", "x=unknown:/path", "-output", out}); err == nil {
		t.Error("expected error for unknown source type")
	}
}

func TestFederateWithStrategy(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.json")
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "d.md"), []byte("# Data"), 0644)

	sourceDef := "data=file:" + srcDir
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-sources", sourceDef, "-strategy", "weighted_average", "-output", out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Errorf("expected output: %v", err)
	}
}
