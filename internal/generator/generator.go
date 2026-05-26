package generator

import (
	"fmt"
	"strings"
)

const MinFacts = 3

type Fact struct {
	Subject   string
	Predicate string
	Object    string
}

type Output struct {
	Type    string
	Content string
	Format  string
	Path    string
}

type Generator interface {
	Generate(facts []Fact, query string) (*Output, error)
	Name() string
}

var AvailableGenerators = map[string]Generator{
	"code": &CodeGenerator{},
	"api":  &APIGenerator{},
	"faq":  &FAQGenerator{},
}

func GetGenerator(name string) (Generator, error) {
	g, ok := AvailableGenerators[name]
	if !ok {
		available := make([]string, 0, len(AvailableGenerators))
		for n := range AvailableGenerators {
			available = append(available, n)
		}
		return nil, fmt.Errorf("unknown generator %q, available: %v", name, available)
	}
	return g, nil
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	first := string(runes[0:1])
	upper := strings.ToUpper(first)
	if len(runes) > 1 {
		return upper + string(runes[1:])
	}
	return upper
}

func uniqueSubjects(facts []Fact) []string {
	seen := map[string]bool{}
	var subjects []string
	for _, f := range facts {
		if !seen[f.Subject] {
			seen[f.Subject] = true
			subjects = append(subjects, f.Subject)
		}
	}
	return subjects
}

func filterFacts(facts []Fact, subject string) []Fact {
	var filtered []Fact
	for _, f := range facts {
		if f.Subject == subject {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
