package markdown_test

import (
	"fmt"
	"log"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/markdown"
)

func ExampleParse() {
	result, err := markdown.Parse([]byte("---\ntitle: 示例\n---\n# Hello"), "test.md")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Frontmatter["title"])
	fmt.Print(result.Body)
	// Output:
	// 示例
	// # Hello
}

func ExampleParse_noFrontmatter() {
	result, err := markdown.Parse([]byte("纯文本内容"), "plain.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(result.Frontmatter))
	fmt.Print(result.Body)
	// Output:
	// 0
	// 纯文本内容
}

func ExampleParse_unclosedFrontmatter() {
	result, err := markdown.Parse([]byte("---\ntitle: 测试\n未闭合"), "bad.md")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Malformed)
	fmt.Println(result.Body)
	// Output:
	// true
	// ---
	// title: 测试
	// 未闭合
}

func ExampleDetectAndDecode() {
	decoded, enc, err := markdown.DetectAndDecode([]byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(enc)
	fmt.Print(string(decoded))
	// Output:
	// utf-8
	// hello
}
