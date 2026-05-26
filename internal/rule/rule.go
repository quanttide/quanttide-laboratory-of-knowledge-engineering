package rule

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Rule struct {
	Name       string
	Head       Atom
	Body       []Atom
	Raw        string
}

type Atom struct {
	Predicate string
	Vars      []string
}

type RuleSet struct {
	Rules      []Rule
	Predicates map[string]bool
	Errors     []string
}

func Load(path string) (*RuleSet, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("rule path error: %w", err)
	}

	rs := &RuleSet{
		Predicates: make(map[string]bool),
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mgl") {
				if err := rs.loadFile(filepath.Join(path, entry.Name())); err != nil {
					rs.Errors = append(rs.Errors, fmt.Sprintf("%s: %v", entry.Name(), err))
				}
			}
		}
	} else {
		if err := rs.loadFile(path); err != nil {
			rs.Errors = append(rs.Errors, err.Error())
		}
	}

	rs.detectUndefined()
	rs.detectCycles()

	return rs, nil
}

func (rs *RuleSet) loadFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(raw)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		rule, err := parseRule(line)
		if err != nil {
			return fmt.Errorf("parse error in %s: %w", path, err)
		}
		if rule != nil {
			rs.Rules = append(rs.Rules, *rule)
			rs.Predicates[rule.Head.Predicate] = true
			for _, bodyAtom := range rule.Body {
				rs.Predicates[bodyAtom.Predicate] = true
			}
		}
	}

	return nil
}

func parseRule(line string) (*Rule, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	parts := strings.SplitN(line, ":-", 2)
	if len(parts) < 1 {
		return nil, nil
	}

	headStr := strings.TrimSpace(parts[0])
	head, err := ParseAtom(headStr)
	if err != nil {
		return nil, fmt.Errorf("head atom: %w", err)
	}

	rule := &Rule{
		Name: head.Predicate,
		Head: *head,
		Raw:  line,
	}

	if len(parts) == 2 {
		bodyStr := strings.TrimSpace(parts[1])
		bodyAtoms := splitTopLevel(bodyStr, ',')
		for _, atomStr := range bodyAtoms {
			atomStr = strings.TrimSpace(atomStr)
			if atomStr == "" {
				continue
			}
			atom, err := ParseAtom(atomStr)
			if err != nil {
				return nil, fmt.Errorf("body atom %q: %w", atomStr, err)
			}
			rule.Body = append(rule.Body, *atom)
		}
	}

	return rule, nil
}

func isVariable(s string) bool {
	if strings.HasPrefix(s, "?") {
		return true
	}
	if len(s) == 1 && s[0] >= 'A' && s[0] <= 'Z' {
		return true
	}
	return false
}

func splitTopLevel(s string, sep rune) []string {
	var result []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case sep:
			if depth == 0 {
				result = append(result, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if start < len(s) {
		result = append(result, strings.TrimSpace(s[start:]))
	}
	return result
}

func ParseAtom(s string) (*Atom, error) {
	s = strings.TrimSpace(s)
	parenIdx := strings.Index(s, "(")
	if parenIdx < 0 {
		return nil, fmt.Errorf("missing '(' in atom: %s", s)
	}
	if !strings.HasSuffix(s, ")") {
		return nil, fmt.Errorf("missing ')' in atom: %s", s)
	}

	predicate := strings.TrimSpace(s[:parenIdx])
	varsStr := s[parenIdx+1 : len(s)-1]

	var vars []string
	for _, v := range strings.Split(varsStr, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			vars = append(vars, v)
		}
	}

	return &Atom{
		Predicate: predicate,
		Vars:      vars,
	}, nil
}

func (rs *RuleSet) detectUndefined() {
	defined := make(map[string]bool)
	for _, r := range rs.Rules {
		defined[r.Head.Predicate] = true
	}

	for _, r := range rs.Rules {
		for _, body := range r.Body {
			if !defined[body.Predicate] {
				rs.Errors = append(rs.Errors,
					fmt.Sprintf("undefined predicate %q in rule %q", body.Predicate, r.Name))
			}
		}
	}
}

func (rs *RuleSet) detectCycles() {
	graph := make(map[string][]string)
	for _, r := range rs.Rules {
		for _, body := range r.Body {
			graph[r.Head.Predicate] = append(graph[r.Head.Predicate], body.Predicate)
		}
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true
		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if inStack[neighbor] {
				rs.Errors = append(rs.Errors,
					fmt.Sprintf("cyclic dependency detected: %s -> %s", node, neighbor))
				return true
			}
		}
		inStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] {
			dfs(node)
		}
	}
}

const RuleTimeout = 10 * time.Second

func (rs *RuleSet) EvaluateQuery(query Atom, facts []Fact, timeout time.Duration) ([]Fact, error) {
	if timeout <= 0 {
		timeout = RuleTimeout
	}

	done := make(chan []Fact, 1)
	errCh := make(chan error, 1)

	go func() {
		results, err := rs.evaluate(query, facts)
		if err != nil {
			errCh <- err
			return
		}
		done <- results
	}()

	select {
	case results := <-done:
		return results, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("query %q timed out after %v", query.Predicate, timeout)
	}
}

type Fact struct {
	Predicate string
	Args      []string
}

func (rs *RuleSet) evaluate(query Atom, facts []Fact) ([]Fact, error) {
	var results []Fact

	for _, r := range rs.Rules {
		if r.Head.Predicate != query.Predicate {
			continue
		}

		if len(r.Body) == 0 {
			results = append(results, Fact{
				Predicate: r.Head.Predicate,
				Args:      r.Head.Vars,
			})
			continue
		}

		matches := matchRule(r, facts)
		results = append(results, matches...)
	}

	return results, nil
}

func matchRule(rule Rule, facts []Fact) []Fact {
	if len(rule.Body) == 0 {
		return nil
	}

	type binding map[string]string
	var allBindings []binding

	for _, bodyAtom := range rule.Body {
		var newBindings []binding
		for _, fact := range facts {
			if bodyAtom.Predicate != fact.Predicate {
				continue
			}
			if len(bodyAtom.Vars) != len(fact.Args) {
				continue
			}

			b := make(binding)
			match := true
			for i, v := range bodyAtom.Vars {
				if isVariable(v) {
					if existing, ok := b[v]; ok && existing != fact.Args[i] {
						match = false
						break
					}
					b[v] = fact.Args[i]
				} else if v != fact.Args[i] {
					match = false
					break
				}
			}
			if match {
				newBindings = append(newBindings, b)
			}
		}

		if allBindings == nil {
			allBindings = newBindings
		} else {
			var merged []binding
			for _, existing := range allBindings {
				for _, newb := range newBindings {
					compatible := true
					m := make(binding)
					for k, v := range existing {
						m[k] = v
					}
					for k, v := range newb {
						if old, ok := m[k]; ok && old != v {
							compatible = false
							break
						}
						m[k] = v
					}
					if compatible {
						merged = append(merged, m)
					}
				}
			}
			allBindings = merged
		}

		if len(allBindings) == 0 {
			return nil
		}
	}

	var results []Fact
	for _, b := range allBindings {
		args := make([]string, len(rule.Head.Vars))
		for i, v := range rule.Head.Vars {
			if val, ok := b[v]; ok {
				args[i] = val
			} else if !isVariable(v) {
				args[i] = v
			}
		}
		results = append(results, Fact{
			Predicate: rule.Head.Predicate,
			Args:      args,
		})
	}

	return results
}
