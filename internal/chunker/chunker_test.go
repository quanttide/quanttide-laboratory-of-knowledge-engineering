package chunker

import (
	"strings"
	"testing"
)

func TestChunkBasic(t *testing.T) {
	c := New(10)
	text := "一二三四五六七八九十"
	chunks := c.Chunk(text)
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].Text != "一二三四五六七八九十" {
		t.Errorf("unexpected first chunk: %s", chunks[0].Text)
	}
	if chunks[0].Truncated {
		t.Error("single chunk should not be truncated")
	}
}

func TestChunkMultiple(t *testing.T) {
	c := New(5)
	text := "一二三四五六七八九十"
	chunks := c.Chunk(text)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	if !chunks[0].Truncated {
		t.Error("first chunk should be truncated when there are more chunks")
	}
}

func TestChunkOverlap(t *testing.T) {
	c := New(10)
	text := "零一二三四五六七八九零一二三四五六七八九"
	chunks := c.Chunk(text)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	overlap := c.ChunkSize - c.Overlap
	expectedStart := overlap
	if expectedStart < 0 {
		expectedStart = 0
	}
	runes := []rune(text)
	expectedSecond := string(runes[expectedStart : expectedStart+c.ChunkSize])
	if chunks[1].Text != expectedSecond {
		t.Errorf("expected second chunk to start at position %d:\n  got:      %s\n  expected: %s", expectedStart, chunks[1].Text, expectedSecond)
	}
}

func TestChunkEmpty(t *testing.T) {
	c := New(10)
	chunks := c.Chunk("")
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkSmallText(t *testing.T) {
	c := New(100)
	text := "short"
	chunks := c.Chunk(text)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Truncated {
		t.Error("small text should not be truncated")
	}
}

func TestChunkIndexSequence(t *testing.T) {
	c := New(3)
	text := "abcdefghij"
	chunks := c.Chunk(text)
	for i, ch := range chunks {
		if ch.Index != i {
			t.Errorf("chunk %d: expected index %d, got %d", i, i, ch.Index)
		}
	}
}

func TestCountTokens(t *testing.T) {
	text := "hello world 你好"
	count := CountTokens(text)
	if count <= 0 {
		t.Error("expected positive token count")
	}
}

func TestChunkNoLoss(t *testing.T) {
	c := New(5)
	text := "abcdefghijklmnopqrstuvwxyz"
	chunks := c.Chunk(text)

	var reconstructed strings.Builder
	for i, ch := range chunks {
		reconstructed.WriteString(ch.Text)
		if i < len(chunks)-1 {
			start := i*c.ChunkSize - i*c.Overlap
			if start < 0 {
				start = 0
			}
			end := start + len([]rune(ch.Text))
			nextStart := (i+1)*c.ChunkSize - i*c.Overlap
			overlap := end - nextStart
			if overlap > 0 {
				reconstructed.WriteString(string([]rune(ch.Text)[len([]rune(ch.Text))-overlap:]))
			}
		}
	}
}
