package generator

import (
	"fmt"
	"strings"
)

type CodeGenerator struct{}

func (g *CodeGenerator) Name() string { return "code" }

func (g *CodeGenerator) Generate(facts []Fact, query string) (*Output, error) {
	if len(facts) < MinFacts {
		return nil, fmt.Errorf("数据不足：仅 %d 条事实，需要至少 %d 条才能生成有意义的代码", len(facts), MinFacts)
	}

	var code strings.Builder
	code.WriteString("package generated\n\n")
	code.WriteString("import (\n\t\"time\"\n)\n\n")

	subjects := uniqueSubjects(facts)
	for _, subject := range subjects {
		related := filterFacts(facts, subject)
		structName := toGoName(subject)

		code.WriteString(fmt.Sprintf("type %s struct {\n", structName))

		for _, f := range related {
			fieldName := toGoName(f.Object)
			typeName := guessGoType(f.Predicate, f.Object)
			tag := fmt.Sprintf("`json:\"%s\"`", toSnakeCase(f.Object))
			code.WriteString(fmt.Sprintf("\t%s %s %s\n", fieldName, typeName, tag))
		}

		code.WriteString("}\n\n")
	}

	return &Output{
		Type:    "code",
		Content: code.String(),
		Format:  "go",
	}, nil
}

func toGoName(s string) string {
	if s == "" {
		return ""
	}
	parts := splitNonAlpha(s)
	var result strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		result.WriteString(strings.ToUpper(string(runes[0:1])))
		if len(runes) > 1 {
			result.WriteString(string(runes[1:]))
		}
	}
	return result.String()
}

func toSnakeCase(s string) string {
	parts := splitNonAlpha(s)
	return strings.Join(parts, "_")
}

func splitNonAlpha(s string) []string {
	var parts []string
	var buf []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r > 127 {
			buf = append(buf, r)
		} else {
			if len(buf) > 0 {
				parts = append(parts, string(buf))
				buf = nil
			}
		}
	}
	if len(buf) > 0 {
		parts = append(parts, string(buf))
	}
	return parts
}

func guessGoType(predicate, object string) string {
	switch predicate {
	case "包含", "has", "拥有", "属于":
		return "[]" + toGoName(object)
	case "名称", "name", "标题", "title":
		return "string"
	case "数量", "count", "price", "价格":
		return "int"
	case "时间", "date", "created_at":
		return "time.Time"
	default:
		if looksLikeNumber(object) {
			return "float64"
		}
		return "string"
	}
}

func looksLikeNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
