package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/storage"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/triple"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(fs, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	input := fs.String("input", "./triples.jsonl", "input triples file")
	threshold := fs.Float64("threshold", 0.7, "confidence threshold for auto-insert")
	fs.Parse(args)

	policy := storage.ThresholdPolicy{
		High:   *threshold,
		Medium: *threshold - 0.3,
	}
	if policy.Medium < 0 {
		policy.Medium = 0
	}

	dir := filepath.Dir(*input)
	if dir == "." {
		dir = "./data"
	}
	store := storage.NewJSONStore(dir)
	defer store.Close()

	triples, err := readTriples(*input)
	if err != nil {
		return err
	}

	if len(triples) == 0 {
		fmt.Println("! 输入文件无三元组")
		return nil
	}

	result := policy.Classify(triples)

	fmt.Printf("输入 %d 条三元组\n", len(triples))
	fmt.Printf("  高置信度 (≥%.1f): %d 条 — 自动入库\n", policy.High, len(result.AutoInsert))
	fmt.Printf("  中置信度 (≥%.1f): %d 条 — 暂存待确认\n", policy.Medium, len(result.Pending))
	fmt.Printf("  低置信度 (<%.1f): %d 条 — 保留不入库\n", policy.Medium, len(result.Discarded))

	if len(result.AutoInsert) == 0 {
		fmt.Println("! 没有达到阈值的三元组，无需入库")
		return nil
	}

	inserted, err := store.InsertTriples(result.AutoInsert)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✓ 成功入库 %d 条三元组\n", inserted)

	if len(result.Pending) > 0 {
		pendingPath := filepath.Join(filepath.Dir(*input), "pending.jsonl")
		f, err := os.Create(pendingPath)
		if err == nil {
			defer f.Close()
			enc := json.NewEncoder(f)
			for _, t := range result.Pending {
				enc.Encode(t)
			}
			fmt.Printf("  %d 条暂存至 %s\n", len(result.Pending), pendingPath)
		}
	}
	return nil
}

func readTriples(path string) ([]*triple.Triple, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open triples file: %w", err)
	}
	defer f.Close()

	var triples []*triple.Triple
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var t triple.Triple
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping malformed line: %v\n", err)
			continue
		}
		triples = append(triples, &t)
	}
	return triples, scanner.Err()
}
