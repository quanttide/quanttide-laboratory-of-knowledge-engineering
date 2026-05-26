package chunker_test

import (
	"fmt"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/chunker"
)

func ExampleNew() {
	c := chunker.New(5)
	chunks := c.Chunk("一二三四五六七八九十")
	fmt.Println(len(chunks))
	fmt.Println(chunks[0].Text)
	fmt.Println(chunks[0].Truncated)
	fmt.Println(chunks[1].Truncated)
	// Output:
	// 3
	// 一二三四五
	// true
	// true
}

func ExampleChunker_singleChunk() {
	c := chunker.New(100)
	chunks := c.Chunk("短文本")
	fmt.Println(len(chunks))
	fmt.Println(chunks[0].Truncated)
	// Output:
	// 1
	// false
}

func ExampleChunker_empty() {
	c := chunker.New(10)
	chunks := c.Chunk("")
	fmt.Println(len(chunks))
	// Output: 0
}
