package generator

import (
	"fmt"
	"sort"
	"strings"
)

type APIGenerator struct{}

func (g *APIGenerator) Name() string { return "api" }

func (g *APIGenerator) Generate(facts []Fact, query string) (*Output, error) {
	if len(facts) < MinFacts {
		return nil, fmt.Errorf("数据不足：仅 %d 条事实，需要至少 %d 条才能生成 API 设计", len(facts), MinFacts)
	}

	var doc strings.Builder
	doc.WriteString("# API 设计建议\n\n")
	doc.WriteString(fmt.Sprintf("> 基于 %d 条知识事实生成的 RESTful API 设计\n\n", len(facts)))

	subjects := uniqueSubjects(facts)
	sort.Strings(subjects)

	for _, subject := range subjects {
		resourcePath := toKebabCase(subject)

		doc.WriteString(fmt.Sprintf("## %s\n\n", toGoName(subject)))
		doc.WriteString(fmt.Sprintf("资源路径: `/api/v1/%s`\n\n", resourcePath))

		related := filterFacts(facts, subject)

		doc.WriteString("| 方法 | 路径 | 说明 |\n")
		doc.WriteString("|------|------|------|\n")
		doc.WriteString(fmt.Sprintf("| GET | `/api/v1/%s` | 列表查询 |\n", resourcePath))
		doc.WriteString(fmt.Sprintf("| POST | `/api/v1/%s` | 创建 |\n", resourcePath))
		doc.WriteString(fmt.Sprintf("| GET | `/api/v1/%s/:id` | 详情 |\n", resourcePath))
		doc.WriteString(fmt.Sprintf("| PUT | `/api/v1/%s/:id` | 更新 |\n", resourcePath))
		doc.WriteString(fmt.Sprintf("| DELETE | `/api/v1/%s/:id` | 删除 |\n", resourcePath))

		if len(related) > 0 {
			doc.WriteString("\n关联字段:\n\n")
			doc.WriteString("| 字段 | 类型 | 说明 |\n")
			doc.WriteString("|------|------|------|\n")
			for _, f := range related {
				typeName := guessGoType(f.Predicate, f.Object)
				doc.WriteString(fmt.Sprintf("| %s | %s | %s |\n", toSnakeCase(f.Object), typeName, f.Predicate))
			}
		}

		doc.WriteString("\n---\n\n")
	}

	return &Output{
		Type:    "api",
		Content: doc.String(),
		Format:  "markdown",
	}, nil
}

func toKebabCase(s string) string {
	parts := splitNonAlpha(s)
	return strings.Join(parts, "-")
}
