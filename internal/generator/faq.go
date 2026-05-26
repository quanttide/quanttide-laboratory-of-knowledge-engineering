package generator

import (
	"fmt"
	"strings"
)

type FAQGenerator struct{}

func (g *FAQGenerator) Name() string { return "faq" }

func (g *FAQGenerator) Generate(facts []Fact, query string) (*Output, error) {
	if len(facts) < MinFacts {
		return nil, fmt.Errorf("数据不足：仅 %d 条事实，需要至少 %d 条才能生成 FAQ", len(facts), MinFacts)
	}

	var faq strings.Builder
	faq.WriteString("# FAQ - 常见问题\n\n")
	faq.WriteString(fmt.Sprintf("> 基于 %d 条知识事实自动生成\n\n", len(facts)))

	seen := make(map[string]bool)

	for i, f := range facts {
		q := generateQuestion(f)
		if seen[q] {
			continue
		}
		seen[q] = true

		a := generateAnswer(f)
		faq.WriteString(fmt.Sprintf("## Q%d: %s\n\n", i+1, q))
		faq.WriteString(fmt.Sprintf("%s\n\n", a))
	}

	return &Output{
		Type:    "faq",
		Content: faq.String(),
		Format:  "markdown",
	}, nil
}

func generateQuestion(f Fact) string {
	templates := []string{
		fmt.Sprintf("%s 的 %s 是什么？", f.Subject, f.Predicate),
		fmt.Sprintf("%s 如何关联 %s？", f.Subject, f.Object),
		fmt.Sprintf("什么是 %s 的 %s？", f.Subject, f.Object),
		fmt.Sprintf("%s 与 %s 的关系是什么？", f.Subject, f.Object),
	}
	return templates[0]
}

func generateAnswer(f Fact) string {
	return fmt.Sprintf("%s 的 %s 是 %s。", f.Subject, f.Predicate, f.Object)
}
