package rule_test

import (
	"fmt"

	"github.com/quanttide/quanttide-example-of-knowledge-engineering/internal/rule"
)

func ExampleParseAtom() {
	atom, err := rule.ParseAtom("contains_tc(A, C)")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(atom.Predicate)
	fmt.Println(atom.Vars)
	// Output:
	// contains_tc
	// [A C]
}

func ExampleRuleSet_EvaluateQuery() {
	rs := &rule.RuleSet{}
	rs.Rules = append(rs.Rules, rule.Rule{
		Name: "child_of",
		Head: rule.Atom{Predicate: "child_of", Vars: []string{"?X", "?Y"}},
		Body: []rule.Atom{
			{Predicate: "parent", Vars: []string{"?Y", "?X"}},
		},
	})

	facts := []rule.Fact{
		{Predicate: "parent", Args: []string{"Alice", "Bob"}},
	}

	query := rule.Atom{Predicate: "child_of", Vars: []string{"?X", "?Y"}}
	results, err := rs.EvaluateQuery(query, facts, 0)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, f := range results {
		fmt.Printf("%s(%s)\n", f.Predicate, f.Args[0])
	}
	// Output:
	// child_of(Bob)
}
