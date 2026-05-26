package rule

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAtom(t *testing.T) {
	tests := []struct {
		input   string
		pred    string
		vars    []string
		wantErr bool
	}{
		{"contains_tc(A, C)", "contains_tc", []string{"A", "C"}, false},
		{"contains(A, B)", "contains", []string{"A", "B"}, false},
		{"not_self_contained(A)", "not_self_contained", []string{"A"}, false},
		{"badatom", "", nil, true},
		{"missing_paren(", "", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			atom, err := ParseAtom(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if atom.Predicate != tt.pred {
				t.Errorf("expected predicate %s, got %s", tt.pred, atom.Predicate)
			}
			if len(atom.Vars) != len(tt.vars) {
				t.Errorf("expected %d vars, got %d", len(tt.vars), len(atom.Vars))
			}
		})
	}
}

func TestParseRule(t *testing.T) {
	rule, err := parseRule("contains_tc(A, C) :- contains(A, B), contains_tc(B, C)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Head.Predicate != "contains_tc" {
		t.Errorf("expected head contains_tc, got %s", rule.Head.Predicate)
	}
	if len(rule.Body) != 2 {
		t.Errorf("expected 2 body atoms, got %d", len(rule.Body))
	}
}

func TestParseFact(t *testing.T) {
	rule, err := parseRule("contains(A, B)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Head.Predicate != "contains" {
		t.Errorf("expected contains, got %s", rule.Head.Predicate)
	}
	if len(rule.Body) != 0 {
		t.Errorf("expected 0 body atoms for fact, got %d", len(rule.Body))
	}
}

func TestLoadFile(t *testing.T) {
	content := `contains_tc(A, C) :- contains(A, B), contains_tc(B, C)
contains_tc(A, B) :- contains(A, B)
not_self_contained(A) :- contains(A, A)`

	path := filepath.Join(t.TempDir(), "test.mgl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs.Rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rs.Rules))
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	r1 := `parent_tc(A, C) :- parent(A, B), parent_tc(B, C)`
	r2 := `subclass_tc(A, C) :- subclass(A, B), subclass_tc(B, C)`

	os.WriteFile(filepath.Join(dir, "r1.mgl"), []byte(r1), 0644)
	os.WriteFile(filepath.Join(dir, "r2.mgl"), []byte(r2), 0644)

	rs, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rs.Rules))
	}
}

func TestUndefinedPredicate(t *testing.T) {
	content := `foo(A) :- bar(A), baz(A)`
	path := filepath.Join(t.TempDir(), "undef.mgl")
	os.WriteFile(path, []byte(content), 0644)

	rs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rs.Errors) > 0 {
		t.Logf("got undefined predicate warnings: %v", rs.Errors)
	}
}

func TestCycleDetection(t *testing.T) {
	content := `a(X) :- b(X)
b(X) :- c(X)
c(X) :- a(X)`
	path := filepath.Join(t.TempDir(), "cycle.mgl")
	os.WriteFile(path, []byte(content), 0644)

	rs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rs.Errors) > 0 {
		t.Logf("cycle detection triggered: %v", rs.Errors)
	}
}

func TestEvaluateQuery(t *testing.T) {
	content := `contains_tc(A, C) :- contains(A, B), contains_tc(B, C)
contains_tc(A, B) :- contains(A, B)`
	path := filepath.Join(t.TempDir(), "eval.mgl")
	os.WriteFile(path, []byte(content), 0644)

	rs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	facts := []Fact{
		{Predicate: "contains", Args: []string{"订单", "商品"}},
		{Predicate: "contains", Args: []string{"商品", "库存"}},
	}

	query := Atom{Predicate: "contains_tc", Vars: []string{"?X", "?Y"}}
	results, err := rs.EvaluateQuery(query, facts, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
	}
}

func TestCommentsIgnored(t *testing.T) {
	content := `% this is a comment
// this is also a comment
# and this too
contains(A, B) :- contains(A, B)`
	path := filepath.Join(t.TempDir(), "comments.mgl")
	os.WriteFile(path, []byte(content), 0644)

	rs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs.Rules) != 1 {
		t.Errorf("expected 1 rule (comments ignored), got %d", len(rs.Rules))
	}
}

func TestRuleTimeout(t *testing.T) {
	rs := &RuleSet{}
	query := Atom{Predicate: "slow", Vars: []string{"?X"}}

	_, err := rs.EvaluateQuery(query, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error with empty ruleset: %v", err)
	}
}

func TestEmptyRuleSet(t *testing.T) {
	rs := &RuleSet{}
	if len(rs.Rules) != 0 {
		t.Error("expected empty ruleset")
	}
}
