package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/document"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/markdown"
)

type ErrorRecord struct {
	Path      string `json:"path"`
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(fs, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	input := fs.String("input", "", "input file or directory")
	output := fs.String("output", "./parsed", "output directory")
	fs.Parse(args)

	if *input == "" {
		return fmt.Errorf("error: -input is required")
	}

	info, err := os.Stat(*input)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(*output, 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}

	var files []string
	if info.IsDir() {
		if err := filepath.WalkDir(*input, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".txt" {
				files = append(files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("walking directory: %w", err)
		}
	} else {
		files = append(files, *input)
	}

	success := 0
	var errors []ErrorRecord

	for _, path := range files {
		err := processFile(path, *output)
		if err != nil {
			errors = append(errors, ErrorRecord{
				Path:      path,
				Error:     err.Error(),
				Timestamp: "now",
			})
			fmt.Fprintf(os.Stderr, "warn: %s: %v\n", path, err)
			continue
		}
		success++
	}

	if len(errors) > 0 {
		errPath := filepath.Join(*output, "errors.jsonl")
		f, err := os.Create(errPath)
		if err == nil {
			defer f.Close()
			enc := json.NewEncoder(f)
			for _, e := range errors {
				enc.Encode(e)
			}
		}
	}

	fmt.Printf("✓ 处理 %d 个文件，成功 %d 个，失败 %d 个", len(files), success, len(errors))
	if len(errors) > 0 {
		fmt.Printf("（详见 %s/errors.jsonl）", *output)
	}
	fmt.Println()
	return nil
}

func processFile(path, outputDir string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	decoded, _, err := markdown.DetectAndDecode(raw)
	if err != nil {
		return err
	}

	result, err := markdown.Parse(decoded, path)
	if err != nil {
		return err
	}

	doc := document.New(path, result.Frontmatter, result.Body)

	outPath := filepath.Join(outputDir, filepath.Base(path)+".json")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	return nil
}
