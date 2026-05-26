package document_test

import (
	"fmt"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/document"
)

func ExampleNew() {
	doc := document.New("doc.md", map[string]any{"title": "测试"}, "# Hello\n世界")
	fmt.Println(doc.Path)
	fmt.Println(doc.Frontmatter["title"])
	fmt.Print(doc.Body)
	// Output:
	// doc.md
	// 测试
	// # Hello
	// 世界
}

func ExampleDocument_Summarize() {
	doc := document.New("readme.md", map[string]any{"version": 2}, "你好世界")
	fmt.Println(doc.Summarize())
	// Output: readme.md | 正文 4 字, 1 个块, frontmatter 1 字段
}

func ExampleDocument_Blocks() {
	doc := document.New("example.md", nil, "# 标题\n\n正文段落")
	fmt.Println(len(doc.Blocks))
	// Output: 2
}
