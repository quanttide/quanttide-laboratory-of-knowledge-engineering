package document

import (
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := New("test.md", map[string]any{"title": "Test"}, "# Hello\nWorld")
	if doc.Path != "test.md" {
		t.Errorf("expected path test.md, got %s", doc.Path)
	}
	if doc.Body != "# Hello\nWorld" {
		t.Errorf("unexpected body: %s", doc.Body)
	}
	if doc.Frontmatter["title"] != "Test" {
		t.Errorf("unexpected frontmatter title: %v", doc.Frontmatter["title"])
	}
}

func TestSummarize(t *testing.T) {
	doc := New("doc.md", map[string]any{"a": "b"}, "body text")
	summary := doc.Summarize()
	if summary == "" {
		t.Error("summary should not be empty")
	}
}

func TestSplitBlocks(t *testing.T) {
	body := "# Title\n\nSome paragraph.\n\n```\ncode block\n```\n\nMore text."
	doc := New("test.md", nil, body)
	if len(doc.Blocks) < 2 {
		t.Errorf("expected at least 2 blocks, got %d", len(doc.Blocks))
	}
}

func TestEmptyBody(t *testing.T) {
	doc := New("empty.md", nil, "")
	if len(doc.Blocks) != 0 {
		t.Errorf("expected 0 blocks for empty body, got %d", len(doc.Blocks))
	}
}

func TestChineseText(t *testing.T) {
	body := "你好世界"
	doc := New("zh.md", nil, body)
	if doc.Body != body {
		t.Errorf("body mismatch")
	}
}
