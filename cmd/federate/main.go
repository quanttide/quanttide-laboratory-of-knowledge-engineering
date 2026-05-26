package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/source"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(fs, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	sources := fs.String("sources", "", "comma-separated source definitions: name=type:path")
	strategy := fs.String("strategy", "max_confidence", "fusion strategy (max_confidence / weighted_average)")
	output := fs.String("output", "./fused.json", "output file")
	fs.Parse(args)

	if *sources == "" {
		return fmt.Errorf("error: -sources is required\nformat: name1=file:./docs,name2=api:https://api.example.com/data")
	}

	federator := source.NewFederator()
	for _, def := range strings.Split(*sources, ",") {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}
		src, err := parseSourceDef(def)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %q: %v\n", def, err)
			continue
		}
		federator.AddSource(src)
	}

	if federator.SourceCount() == 0 {
		return fmt.Errorf("no valid sources defined")
	}
	fmt.Printf("✓ 加载 %d 个数据源\n", federator.SourceCount())

	allRecords, err := federator.FetchAll()
	if err != nil {
		return err
	}

	total := 0
	for name, records := range allRecords {
		fmt.Printf("  %s: %d 条记录\n", name, len(records))
		total += len(records)
	}
	fmt.Printf("✓ 共获取 %d 条原始记录\n", total)

	fusionStrategy, err := source.NewFusionStrategy(*strategy, nil)
	if err != nil {
		return err
	}

	fused := fusionStrategy.Fuse(allRecords)
	fmt.Printf("✓ 融合后 %d 条三元组（策略: %s）\n", len(fused), fusionStrategy.Name())

	data, err := json.MarshalIndent(fused, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(*output, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	fmt.Printf("✓ 输出到 %s\n", *output)
	return nil
}

func parseSourceDef(def string) (source.Source, error) {
	parts := strings.SplitN(def, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format, expected name=type:path")
	}
	name := strings.TrimSpace(parts[0])

	value := strings.TrimSpace(parts[1])
	typeParts := strings.SplitN(value, ":", 2)
	if len(typeParts) != 2 {
		return nil, fmt.Errorf("invalid format, expected type:path")
	}

	srcType := typeParts[0]
	srcPath := typeParts[1]

	switch srcType {
	case "file":
		return source.NewFileSource(name, srcPath, nil), nil
	case "api":
		return source.NewAPISource(name, srcPath), nil
	default:
		return nil, fmt.Errorf("unknown source type %q (supported: file, api)", srcType)
	}
}
