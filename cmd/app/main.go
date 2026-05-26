package main

import (
	"fmt"
	"os"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/pkg/knowledge"
)

func main() {
	text := "Hello, 知识工程!"
	if len(os.Args) > 1 {
		text = os.Args[1]
	}

	doc := knowledge.NewDocument(text)
	fmt.Println(doc.Summarize())
}
