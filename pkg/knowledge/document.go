package knowledge

import "fmt"

// Document 表示一个待处理的知识文档。
type Document struct {
	Content string
}

// NewDocument 创建新文档。
func NewDocument(content string) *Document {
	return &Document{Content: content}
}

// Summarize 返回文档摘要信息。
func (d *Document) Summarize() string {
	return fmt.Sprintf("文档长度: %d 字", len([]rune(d.Content)))
}
