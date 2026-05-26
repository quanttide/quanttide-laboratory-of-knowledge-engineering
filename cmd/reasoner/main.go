package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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
	rulePath := fs.String("rule", "./rules", "rule file or directory")
	query := fs.String("query", "", "query target (e.g. contains_tc(订单, ?X))")
	timeout := fs.Int("timeout", 10, "single rule timeout in seconds")
	fs.Parse(args)

	if *query == "" {
		return fmt.Errorf("error: -query is required")
	}

	rs, err := rule.Load(*rulePath)
	if err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	if len(rs.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "规则编译警告:\n")
		for _, e := range rs.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
	}

	if len(rs.Rules) == 0 {
		fmt.Println("! 没有加载到任何规则")
		return nil
	}
	fmt.Printf("✓ 加载 %d 条规则", len(rs.Rules))
	if len(rs.Errors) > 0 {
		fmt.Printf("，%d 个警告", len(rs.Errors))
	}
	fmt.Println()

	queryAtom, err := rule.ParseAtom(*query)
	if err != nil {
		return fmt.Errorf("parse query: %w", err)
	}

	facts := loadSampleFacts()

	results, err := rs.EvaluateQuery(*queryAtom, facts, time.Duration(*timeout)*time.Second)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("! 查询无结果")
		return nil
	}

	fmt.Printf("查询 %s 返回 %d 条结果:\n", *query, len(results))
	for _, f := range results {
		args := strings.Join(f.Args, ", ")
		fmt.Printf("  %s(%s)\n", f.Predicate, args)
	}
	return nil
}

func loadSampleFacts() []rule.Fact {
	return []rule.Fact{
		{Predicate: "contains", Args: []string{"订单", "商品"}},
		{Predicate: "contains", Args: []string{"商品", "库存"}},
		{Predicate: "subclass", Args: []string{"电子产品", "商品"}},
		{Predicate: "subclass", Args: []string{"手机", "电子产品"}},
		{Predicate: "role", Args: []string{"alice", "admin"}},
		{Predicate: "role_permission", Args: []string{"admin", "delete_order"}},
		{Predicate: "order", Args: []string{"order-001"}},
		{Predicate: "item", Args: []string{"order-001", "item-xyz"}},
		{Predicate: "insufficient_stock", Args: []string{"item-xyz"}},
	}
}
