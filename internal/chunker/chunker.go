package chunker

import (
	"strings"
	"unicode/utf8"
)

const DefaultOverlapRatio = 0.10

type Chunk struct {
	Index     int
	Text      string
	Truncated bool
}

type Chunker struct {
	ChunkSize  int
	Overlap    int
}

func New(chunkSize int) *Chunker {
	overlap := int(float64(chunkSize) * DefaultOverlapRatio)
	if overlap < 1 {
		overlap = 1
	}
	return &Chunker{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

func (c *Chunker) Chunk(text string) []Chunk {
	runes := []rune(text)
	total := len(runes)
	if total == 0 {
		return nil
	}

	var chunks []Chunk
	start := 0
	index := 0

	for start < total {
		end := start + c.ChunkSize
		truncated := false
		if end >= total {
			end = total
		} else {
			truncated = true
		}

		chunkText := string(runes[start:end])
		chunks = append(chunks, Chunk{
			Index:     index,
			Text:      chunkText,
			Truncated: truncated,
		})
		index++

		if end == total {
			break
		}

		start = end - c.Overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

func CountTokens(text string) int {
	words := strings.Fields(text)
	count := len(words)
	count += utf8.RuneCountInString(text) / 2
	return count
}
