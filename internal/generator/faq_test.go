package generator

import (
	"fmt"
	"testing"
)

func TestGenerateQuestion(t *testing.T) {
	f := Fact{Subject: "订单", Predicate: "包含", Object: "商品"}
	q := generateQuestion(f)
	if q == "" {
		t.Error("expected non-empty question")
	}
}

func TestGenerateAnswer(t *testing.T) {
	f := Fact{Subject: "订单", Predicate: "包含", Object: "商品"}
	a := generateAnswer(f)
	if !contains(a, "订单") || !contains(a, "商品") {
		t.Errorf("answer should contain subject and object: %s", a)
	}
}

func TestFAQNoDuplicates(t *testing.T) {
	facts := []Fact{
		{Subject: "A", Predicate: "B", Object: "C"},
		{Subject: "A", Predicate: "B", Object: "C"},
		{Subject: "X", Predicate: "Y", Object: "Z"},
	}
	g := &FAQGenerator{}
	result, err := g.Generate(facts, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Count "Q1:" and "Q2:" headers
	qCount := 0
	for i := 1; i <= 10; i++ {
		header := fmt.Sprintf("Q%d:", i)
		if contains(result.Content, header) {
			qCount++
		}
	}
	if qCount != 2 {
		t.Errorf("expected 2 unique Qs, got %d", qCount)
	}
}
