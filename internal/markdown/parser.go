package markdown

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrBinaryFile    = errors.New("binary file not supported")
	ErrEmptyFile     = errors.New("empty file")
	ErrFileTooLarge  = errors.New("file too large")
	ErrUnknownEncode = errors.New("unknown encoding")
)

const MaxFileSize = 10 * 1024 * 1024

type ParseResult struct {
	Frontmatter    map[string]any
	Body           string
	FrontmatterRaw string
	Warnings       []string
	Malformed      bool
}

func Parse(content []byte, path string) (*ParseResult, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrEmptyFile, path)
	}
	if len(content) > MaxFileSize {
		return nil, fmt.Errorf("%w: %s (%d bytes)", ErrFileTooLarge, path, len(content))
	}

	text := string(content)

	if !isLikelyText(text) {
		return nil, fmt.Errorf("%w: %s", ErrBinaryFile, path)
	}

	result := &ParseResult{
		Frontmatter: make(map[string]any),
	}

	stripped := strings.TrimLeft(text, "\n\r ")
	if !strings.HasPrefix(stripped, "---") {
		result.Body = text
		return result, nil
	}

	raw := stripped[3:]
	endIdx := strings.Index(raw, "\n---")
	if endIdx < 0 {
		endIdx = strings.Index(raw, "\r\n---")
	}
	if endIdx < 0 {
		result.Malformed = true
		result.Warnings = append(result.Warnings, "unclosed frontmatter, treating as body")
		result.Body = text
		return result, nil
	}

	frontmatterRaw := raw[:endIdx]
	bodyRaw := raw[endIdx+4:]

	result.FrontmatterRaw = strings.TrimSpace(frontmatterRaw)

	if err := parseSimpleYAML(result.FrontmatterRaw, result.Frontmatter); err != nil {
		result.Malformed = true
		result.Warnings = append(result.Warnings, fmt.Sprintf("YAML parse error: %v, treating frontmatter as plain text", err))
		result.Body = text
		return result, nil
	}

	result.Body = strings.TrimLeft(bodyRaw, "\n\r ")
	return result, nil
}

func isLikelyText(content string) bool {
	nulls := strings.Count(content, "\x00")
	ratio := float64(nulls) / float64(len(content)+1)
	return ratio < 0.01
}

func parseSimpleYAML(raw string, out map[string]any) error {
	lines := strings.Split(raw, "\n")
	var stack []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := countIndent(line)
		for len(stack) > indent/2 {
			stack = stack[:len(stack)-1]
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx < 0 {
			continue
		}

		key := strings.TrimSpace(trimmed[:colonIdx])
		rest := strings.TrimSpace(trimmed[colonIdx+1:])

		if rest == "" {
			stack = append(stack, key)
			continue
		}

		if rest == "[]" || rest == "[ ]" {
			setNested(out, stack, key, []string{})
			continue
		}

		if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
			inner := strings.TrimSpace(rest[1 : len(rest)-1])
			var items []string
			for _, item := range strings.Split(inner, ",") {
				item = strings.TrimSpace(item)
				item = strings.Trim(item, "\"'")
				items = append(items, item)
			}
			var anyItems []any
			for _, item := range items {
				anyItems = append(anyItems, item)
			}
			setNested(out, stack, key, anyItems)
			continue
		}

		rest = strings.Trim(rest, "\"'")
		if len(stack) > 0 {
			parent := getNested(out, stack)
			if parent == nil {
				parent = make(map[string]any)
				setNested(out, stack[:len(stack)-1], stack[len(stack)-1], parent)
			}
			if m, ok := parent.(map[string]any); ok {
				m[key] = rest
			}
		} else {
			out[key] = rest
		}
	}

	return nil
}

func countIndent(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
		} else if r == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

func setNested(m map[string]any, stack []string, key string, val any) {
	target := m
	for _, k := range stack {
		existing, ok := target[k]
		if !ok {
			nested := make(map[string]any)
			target[k] = nested
			target = nested
			continue
		}
		if nested, ok := existing.(map[string]any); ok {
			target = nested
		} else {
			nested := make(map[string]any)
			target[k] = nested
			target = nested
		}
	}
	target[key] = val
}

func getNested(m map[string]any, stack []string) any {
	target := m
	for _, k := range stack {
		existing, ok := target[k]
		if !ok {
			return nil
		}
		if nested, ok := existing.(map[string]any); ok {
			target = nested
		} else {
			return existing
		}
	}
	return target
}


