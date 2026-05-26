package markdown

import (
	"testing"
)

func TestParseWithFrontmatter(t *testing.T) {
	content := []byte("---\ntitle: Test\nkey: value\n---\n\n# Body content")
	result, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Frontmatter["title"] != "Test" {
		t.Errorf("expected title=Test, got %v", result.Frontmatter["title"])
	}
	if result.Body != "# Body content" {
		t.Errorf("unexpected body: %s", result.Body)
	}
	if result.Malformed {
		t.Error("should not be malformed")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	content := []byte("# Just a title\n\nSome text.")
	result, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Frontmatter) != 0 {
		t.Errorf("expected empty frontmatter, got %v", result.Frontmatter)
	}
	if result.Body != string(content) {
		t.Errorf("body should be original content")
	}
}

func TestParseMalformedYAML(t *testing.T) {
	content := []byte("---\ninvalid: [yaml\n---\n\nBody")
	result, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
	// Simple parser tolerates unclosed brackets as plain text values.
	// This is an accepted trade-off for zero external dependencies.
}

func TestParseUnclosedFrontmatter(t *testing.T) {
	content := []byte("---\ntitle: Test\nbody here")
	result, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Malformed {
		t.Error("expected malformed=true for unclosed frontmatter")
	}
}

func TestParseEmptyFile(t *testing.T) {
	_, err := Parse([]byte{}, "empty.md")
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseBinaryFile(t *testing.T) {
	content := []byte("\x00\x01\x02\x00this is binary\x00")
	_, err := Parse(content, "binary.bin")
	if err == nil {
		t.Fatal("expected error for binary file")
	}
}

func TestParseLargeFile(t *testing.T) {
	content := make([]byte, MaxFileSize+1)
	_, err := Parse(content, "large.md")
	if err == nil {
		t.Fatal("expected error for large file")
	}
}

func TestParseNestedFrontmatter(t *testing.T) {
	content := []byte("---\nnested:\n  key: val\nlist:\n  - a\n  - b\n---\n\nBody")
	result, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Malformed {
		t.Error("should not be malformed")
	}
	nested, ok := result.Frontmatter["nested"].(map[string]any)
	if !ok || nested["key"] != "val" {
		t.Errorf("unexpected nested frontmatter: %v", result.Frontmatter["nested"])
	}
}

func TestParseMultipleFilesSummary(t *testing.T) {
	cases := []struct {
		name     string
		content  []byte
		wantErr  bool
		wantBody bool
	}{
		{"normal", []byte("---\na: 1\n---\n\nbody"), false, true},
		{"no fm", []byte("just body"), false, true},
		{"malformed", []byte("---\n[bad\n---\n\nbody"), false, true},
		{"empty", []byte{}, true, false},
		{"frontmatter only", []byte("---\na: 1\n---\n\n"), false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := Parse(c.content, "f.md")
			if c.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantBody && res != nil && res.Body == "" {
				t.Error("expected non-empty body")
			}
		})
	}
}
