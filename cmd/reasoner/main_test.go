package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestReasonerWithRules(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "test.mgl")
	os.WriteFile(ruleFile, []byte(`contains_tc(A, B) :- contains(A, B)`), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-rule", ruleFile, "-query", "contains_tc(?X, ?Y)"}); err != nil {
		t.Fatal(err)
	}
}

func TestReasonerEmptyDir(t *testing.T) {
	dir := t.TempDir()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-rule", dir, "-query", "contains_tc(?X, ?Y)"}); err != nil {
		t.Fatal(err)
	}
}

func TestReasonerTransitiveClosure(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "tc.mgl")
	content := `contains_tc(A, C) :- contains(A, B), contains_tc(B, C)
contains_tc(A, B) :- contains(A, B)`
	os.WriteFile(ruleFile, []byte(content), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-rule", ruleFile, "-query", "contains_tc(订单, ?Y)"}); err != nil {
		t.Fatal(err)
	}
}

func TestReasonerNoResults(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "empty.mgl")
	os.WriteFile(ruleFile, []byte(`foo(A) :- bar(A)`), 0644)

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-rule", ruleFile, "-query", "foo(?X)"}); err != nil {
		t.Fatal(err)
	}
}

func TestReasonerNoQuery(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	if err := run(fs, []string{"-rule", "./rules"}); err == nil {
		t.Error("expected error for missing -query")
	}
}
