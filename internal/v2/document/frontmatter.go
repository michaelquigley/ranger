// Package document reads the v2 file shapes through the shared document
// primitives and binds each claimed field through dd.
package document

import (
	"fmt"
	"strings"

	"github.com/michaelquigley/df/dd"
	shared "github.com/michaelquigley/ranger/internal/document"
)

func splitFrontmatter(raw []byte, required bool) (front []byte, body string, err error) {
	text := string(raw)
	opener, rest, _ := strings.Cut(text, "\n")
	if strings.TrimRight(opener, " \t\r") != "---" {
		if required {
			return nil, text, fmt.Errorf("missing frontmatter fence")
		}
		return nil, text, nil
	}
	offset := 0
	for line := range strings.SplitSeq(rest, "\n") {
		switch strings.TrimRight(line, " \t\r") {
		case "---", "...":
			return []byte(rest[:offset]), rest[min(offset+len(line)+1, len(rest)):], nil
		}
		offset += len(line) + 1
	}
	return nil, text, fmt.Errorf("unterminated frontmatter fence")
}

func bindFields(raw []byte, schema any, claimed, required []string) (map[string]bool, []string) {
	fields, err := shared.ParseFields(raw, claimed)
	if err != nil {
		return nil, []string{fmt.Sprintf("frontmatter does not parse: %v", err)}
	}

	counts := map[string]int{}
	for _, field := range fields {
		counts[field.Name]++
	}
	readable := map[string]bool{}
	var diagnostics []string
	for _, field := range fields {
		switch {
		case counts[field.Name] > 1:
			continue
		case field.Err != nil:
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %v", field.Name, field.Err))
		case field.Value == nil:
			diagnostics = append(diagnostics, field.Name+": null value")
		default:
			if err := dd.Bind(schema, map[string]any{field.Name: field.Value}); err != nil {
				diagnostics = append(diagnostics, fmt.Sprintf("%s: %v", field.Name, err))
			} else {
				readable[field.Name] = true
			}
		}
	}
	for _, key := range claimed {
		if counts[key] > 1 {
			diagnostics = append(diagnostics, "duplicate key: "+key)
		}
	}
	for _, key := range required {
		if counts[key] == 0 {
			diagnostics = append(diagnostics, "missing required field: "+key)
		}
	}
	return readable, diagnostics
}
