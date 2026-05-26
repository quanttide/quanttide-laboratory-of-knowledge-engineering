package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/generator"
	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/rule"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(fs, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	target := fs.String("target", "code", "generator target (code / api / faq)")
	query := fs.String("query", "", "query to find relevant facts")
	output := fs.String("output", "./generated", "output directory or file")
	rulePath := fs.String("rule", "./rules", "rule file or directory")
	fs.Parse(args)

	gen, err := generator.GetGenerator(*target)
	if err != nil {
		return err
	}

	facts, err := loadFacts(*query, *rulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: could not load facts from rules: %v\n", err)
		fmt.Fprintf(os.Stderr, "using fallback sample facts\n")
		facts = sampleFacts()
	}

	result, err := gen.Generate(facts, *query)
	if err != nil {
		return err
	}

	if err := writeOutput(result, *output); err != nil {
		fmt.Fprintf(os.Stderr, "warn: cannot write to output path: %v\n", err)
		fmt.Println("--- stdout ---")
		fmt.Println(result.Content)
		return nil
	}

	if result.Format == "go" {
		if err := gofmt(result.Path); err != nil {
			fmt.Fprintf(os.Stderr, "warn: go fmt failed (output preserved): %v\n", err)
		}
	}

	fmt.Printf("✓ 已生成 %s 到 %s\n", result.Type, result.Path)
	return nil
}

func loadFacts(query, rulePath string) ([]generator.Fact, error) {
	if query == "" {
		return nil, fmt.Errorf("no query specified")
	}

	rs, err := rule.Load(rulePath)
	if err != nil {
		return nil, err
	}

	queryAtom, err := rule.ParseAtom(query)
	if err != nil {
		return nil, fmt.Errorf("parse query: %w", err)
	}

	sampleFacts := loadSampleRuleFacts()
	results, err := rs.EvaluateQuery(*queryAtom, sampleFacts, 0)
	if err != nil {
		return nil, err
	}

	var facts []generator.Fact
	for _, r := range results {
		switch len(r.Args) {
		case 2:
			facts = append(facts, generator.Fact{
				Subject:   r.Args[0],
				Predicate: r.Predicate,
				Object:    r.Args[1],
			})
		case 3:
			facts = append(facts, generator.Fact{
				Subject:   r.Args[0],
				Predicate: r.Args[1],
				Object:    r.Args[2],
			})
		}
	}

	return facts, nil
}

func loadSampleRuleFacts() []rule.Fact {
	return []rule.Fact{
		{Predicate: "contains", Args: []string{"订单", "商品"}},
		{Predicate: "contains", Args: []string{"订单", "用户"}},
		{Predicate: "contains", Args: []string{"商品", "库存"}},
		{Predicate: "contains", Args: []string{"用户", "地址"}},
		{Predicate: "subclass", Args: []string{"电子产品", "商品"}},
	}
}

func sampleFacts() []generator.Fact {
	return []generator.Fact{
		{Subject: "订单", Predicate: "包含", Object: "商品"},
		{Subject: "订单", Predicate: "属于", Object: "用户"},
		{Subject: "商品", Predicate: "包含", Object: "库存"},
		{Subject: "用户", Predicate: "拥有", Object: "地址"},
		{Subject: "电子产品", Predicate: "子类", Object: "商品"},
	}
}

func writeOutput(result *generator.Output, outputPath string) error {
	info, err := os.Stat(outputPath)
	isDir := err == nil && info.IsDir()

	var filePath string
	switch {
	case isDir:
		switch result.Format {
		case "go":
			filePath = filepath.Join(outputPath, "types.go")
		case "markdown":
			filePath = filepath.Join(outputPath, result.Type+".md")
		default:
			filePath = filepath.Join(outputPath, result.Type+".txt")
		}
	case strings.HasSuffix(outputPath, "/") || outputPath == "":
		return fmt.Errorf("invalid output path: %s", outputPath)
	default:
		filePath = outputPath
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(result.Content), 0644); err != nil {
		return err
	}

	result.Path = filePath
	return nil
}

func gofmt(path string) error {
	cmd := exec.Command("go", "fmt", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go fmt: %w\n%s", err, string(output))
	}
	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: go run ./cmd/generator -target <code|api|faq> -query <query> [-output <path>] [-rule <rules>]\n\n")
		fmt.Fprintf(os.Stderr, "可用生成器:\n")
		for name, g := range generator.AvailableGenerators {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
			_ = g
		}
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/generator -target code -query \"contains_tc(?X, ?Y)\" -output ./generated\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/generator -target api -query \"contains(?X, ?Y)\" -output ./generated\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/generator -target faq -query \"contains(?X, ?Y)\" -output ./generated\n")
	}
}
