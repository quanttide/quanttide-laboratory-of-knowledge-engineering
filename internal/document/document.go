package document

import (
	"fmt"
	"strings"
)

type Block struct {
	Type    string
	Content string
}

type Document struct {
	Path     string
	Frontmatter map[string]any
	Body     string
	Blocks   []Block
}

func New(path string, fm map[string]any, body string) *Document {
	doc := &Document{
		Path:        path,
		Frontmatter: fm,
		Body:        body,
		Blocks:      splitBlocks(body),
	}
	return doc
}

func (d *Document) Summarize() string {
	return fmt.Sprintf("%s | 正文 %d 字, %d 个块, frontmatter %d 字段",
		d.Path, len([]rune(d.Body)), len(d.Blocks), len(d.Frontmatter))
}

func splitBlocks(body string) []Block {
	lines := strings.Split(body, "\n")
	var blocks []Block
	var buf []string
	blockType := "paragraph"

	flush := func() {
		if len(buf) > 0 {
			blocks = append(blocks, Block{Type: blockType, Content: strings.Join(buf, "\n")})
			buf = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			flush()
			blockType = "code"
			buf = append(buf, line)
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			flush()
			blockType = "heading"
			buf = append(buf, line)
			flush()
			blockType = "paragraph"
			continue
		}
		if trimmed == "" {
			flush()
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return blocks
}
