package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/chunker"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/llm"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

type ParsedDoc struct {
	Path        string         `json:"Path"`
	Frontmatter map[string]any `json:"Frontmatter"`
	Body        string         `json:"Body"`
}

type AssessorError struct {
	Source    string `json:"source"`
	ChunkIdx  int    `json:"chunk_idx"`
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

type LLMExtraction struct {
	Triples []struct {
		Subject         string  `json:"subject"`
		Predicate       string  `json:"predicate"`
		Object          string  `json:"object"`
		Novelty         float64 `json:"novelty"`
		DomainSpecificity float64 `json:"domain_specificity"`
		CounterIntuitive float64 `json:"counter_intuitive"`
		Confidence      float64 `json:"confidence"`
		Sentence        string  `json:"sentence"`
	} `json:"triples"`
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(fs, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	input := fs.String("input", "./parsed", "input directory with parsed JSON files")
	output := fs.String("output", "./triples.jsonl", "output triples file")
	model := fs.String("model", "gpt-4o-mini", "LLM model name")
	apiURL := fs.String("api-url", "https://api.openai.com/v1", "OpenAI-compatible API URL")
	maxFails := fs.Int("max-fails", 3, "max consecutive failures per source before removal")
	fs.Parse(args)

	llmClient := llm.New(*apiURL, os.Getenv("OPENAI_API_KEY"), *model)

	entries, err := os.ReadDir(*input)
	if err != nil {
		return fmt.Errorf("reading input dir: %w", err)
	}

	outFile, err := os.Create(*output)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer outFile.Close()

	ch := chunker.New(2000)
	totalTriples := 0
	totalDocs := 0
	var errors []AssessorError

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(*input, entry.Name())
		doc, err := readParsedDoc(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: read error: %v\n", path, err)
			continue
		}
		totalDocs++

		chunks := ch.Chunk(doc.Body)
		consecFails := 0

		for _, chunk := range chunks {
			triples, err := assessChunk(llmClient, doc.Path, chunk)
			if err != nil {
				consecFails++
				errors = append(errors, AssessorError{
					Source:    doc.Path,
					ChunkIdx:  chunk.Index,
					Error:     err.Error(),
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				})
				if consecFails >= *maxFails {
					fmt.Fprintf(os.Stderr, "warn: %s: %d consecutive failures, removing from queue\n", doc.Path, *maxFails)
					break
				}
				continue
			}
			consecFails = 0

			for _, t := range triples {
				enc := json.NewEncoder(outFile)
				enc.Encode(t)
				totalTriples++
			}
		}
	}

	fmt.Printf("✓ 评估 %d 个文档，产生 %d 条三元组", totalDocs, totalTriples)
	if len(errors) > 0 {
		fmt.Printf("，%d 个错误", len(errors))
	}
	fmt.Println()
	return nil
}

func readParsedDoc(path string) (*ParsedDoc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc ParsedDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func assessChunk(client *llm.Client, source string, chunk chunker.Chunk) ([]*triple.Triple, error) {
	messages := []llm.Message{
		{Role: "system", Content: llm.AssessmentPrompt},
		{Role: "user", Content: chunk.Text},
	}

	resp, err := client.Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("llm error: %w", err)
	}

	content := resp.Choices[0].Message.Content

	var extraction LLMExtraction
	if err := json.Unmarshal([]byte(content), &extraction); err != nil {
		return nil, fmt.Errorf("malformed JSON response: %w", err)
	}

	var triples []*triple.Triple
	for i, t := range extraction.Triples {
		if t.Subject == "" || t.Predicate == "" || t.Object == "" {
			continue
		}

		tri := triple.New(
			fmt.Sprintf("%s-chunk%d-%d", source, chunk.Index, i),
			source,
			t.Confidence,
			t.Subject, t.Predicate, t.Object,
		)
		tri.SetContext("sentence", t.Sentence)
		if chunk.Truncated {
			tri.SetContext("truncated", "true")
		}

		if t.Confidence <= 0 || t.Confidence > 1 {
			tri.MarkMalformed()
		}

		triples = append(triples, tri)
	}

	return triples, nil
}
