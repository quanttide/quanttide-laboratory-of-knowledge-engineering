package generator

import (
	"strings"
	"testing"
)

func sampleTestFacts() []Fact {
	return []Fact{
		{Subject: "订单", Predicate: "包含", Object: "商品"},
		{Subject: "订单", Predicate: "属于", Object: "用户"},
		{Subject: "商品", Predicate: "包含", Object: "库存"},
		{Subject: "用户", Predicate: "拥有", Object: "地址"},
		{Subject: "电子产品", Predicate: "子类", Object: "商品"},
	}
}

func TestCodeGenerator(t *testing.T) {
	g := &CodeGenerator{}
	result, err := g.Generate(sampleTestFacts(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "code" {
		t.Errorf("expected type code, got %s", result.Type)
	}
	if result.Format != "go" {
		t.Errorf("expected format go, got %s", result.Format)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	if !contains(result.Content, "type 订单 struct") &&
		!contains(result.Content, "type Order struct") {
		t.Error("expected Go struct definition in output")
	}
}

func TestAPIGenerator(t *testing.T) {
	g := &APIGenerator{}
	result, err := g.Generate(sampleTestFacts(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "api" {
		t.Errorf("expected type api, got %s", result.Type)
	}
	if result.Format != "markdown" {
		t.Errorf("expected format markdown, got %s", result.Format)
	}
	if !contains(result.Content, "/api/v1/") {
		t.Error("expected API paths in output")
	}
}

func TestFAQGenerator(t *testing.T) {
	g := &FAQGenerator{}
	result, err := g.Generate(sampleTestFacts(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "faq" {
		t.Errorf("expected type faq, got %s", result.Type)
	}
	if !contains(result.Content, "Q1:") {
		t.Error("expected Q&A format in output")
	}
}

func TestGeneratorInsufficientFacts(t *testing.T) {
	gens := []Generator{&CodeGenerator{}, &APIGenerator{}, &FAQGenerator{}}
	for _, g := range gens {
		t.Run(g.Name(), func(t *testing.T) {
			_, err := g.Generate([]Fact{{Subject: "A", Predicate: "B", Object: "C"}}, "")
			if err == nil {
				t.Error("expected error for < 3 facts")
			}
		})
	}
}

func TestGetGenerator(t *testing.T) {
	g, err := GetGenerator("code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Name() != "code" {
		t.Errorf("expected code, got %s", g.Name())
	}

	_, err = GetGenerator("nonexistent")
	if err == nil {
		t.Error("expected error for unknown generator")
	}
}

func TestToGoName(t *testing.T) {
	cases := map[string]string{
		"订单":      "订单",
		"user":    "User",
		"user_name": "UserName",
		"":        "",
	}
	for input, expected := range cases {
		result := toGoName(input)
		if result != expected {
			t.Errorf("toGoName(%q) = %q, want %q", input, result, expected)
		}
	}
}

func TestGuessGoType(t *testing.T) {
	if guessGoType("包含", "商品") != "[]商品" {
		t.Errorf("expected []商品 for 包含 predicate")
	}
	if guessGoType("名称", "订单") != "string" {
		t.Errorf("expected string for 名称 predicate")
	}
	if guessGoType("数量", "10") != "int" {
		t.Errorf("expected int for 数量 predicate")
	}
}

func TestCapitalize(t *testing.T) {
	if capitalize("hello") != "Hello" {
		t.Errorf("expected Hello, got %s", capitalize("hello"))
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
