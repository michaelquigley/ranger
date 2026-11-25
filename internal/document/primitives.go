package document

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/michaelquigley/df/dd"
	"gopkg.in/yaml.v3"
)

// LineSpan is an inclusive, zero-based line range in the parsed input.
type LineSpan struct {
	Start, End int
	Indent     string
	Comment    string
}

// Field holds one claimed mapping entry, including its line span and any
// error converting the value. duplicate entries are returned in file order.
type Field struct {
	Name  string
	Value any
	// Span is suitable for replacement only when the field occupies its own
	// lines in a block mapping.
	Span LineSpan
	Err  error
}

// ParseFields reads claimed top-level fields from a YAML mapping. values
// resolve aliases against the entire mapping before binding.
func ParseFields(raw []byte, claimed []string) ([]Field, error) {
	entries, err := parseMapping(string(raw))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(raw), "\n")
	spans := scanSpans(lines, entries, -1, len(lines)-1)
	var fields []Field
	for i, entry := range entries {
		if !slices.Contains(claimed, entry.name) {
			continue
		}
		value, err := nodeToPlain(entry.val, 0)
		span := spans[i]
		fields = append(fields, Field{
			Name: entry.name, Value: value, Err: err,
			Span: LineSpan{Start: span.start, End: span.end, Indent: span.indent, Comment: span.comment},
		})
	}
	return fields, nil
}

// SequenceEntry is a string entry and its one-based source line.
type SequenceEntry struct {
	Value string
	Line  int
}

// ParseSequence reads a top-level string list in block or flow form. block
// entries occupy one line each. a blank or comment-only document is empty.
func ParseSequence(raw []byte) (entries []SequenceEntry, flow bool, err error) {
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, false, err
	}
	if len(root.Content) == 0 {
		return nil, false, nil
	}
	list := root.Content[0]
	if list.Kind != yaml.SequenceNode {
		return nil, false, errors.New("document is not a list")
	}

	flow = list.Style&yaml.FlowStyle != 0
	lines := strings.Split(string(raw), "\n")
	for i, node := range list.Content {
		plain, err := nodeToPlain(node, 0)
		if err != nil {
			return nil, false, fmt.Errorf("line %d: %w", node.Line, err)
		}
		if _, ok := plain.(string); !ok {
			return nil, false, fmt.Errorf("line %d: expected a string entry", node.Line)
		}
		var schema struct{ Entry string }
		if err := dd.Bind(&schema, map[string]any{"entry": plain}); err != nil {
			return nil, false, fmt.Errorf("line %d: %w", node.Line, err)
		}
		if !flow {
			start, end := node.Line-1, len(lines)-1
			if i+1 < len(list.Content) {
				end = list.Content[i+1].Line - 2
			}
			end = trimSpanEnd(lines, start, end, -1)
			if end != start || node.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 ||
				!strings.HasPrefix(strings.TrimSpace(lines[start]), "-") {
				return nil, false, fmt.Errorf("line %d: expected one entry per line", node.Line)
			}
		}
		entries = append(entries, SequenceEntry{Value: schema.Entry, Line: node.Line})
	}
	return entries, flow, nil
}

// EmitScalar emits one YAML scalar using ranger's shared quoting rules.
func EmitScalar(value string) string { return emitScalar(value) }

// ReplaceSpan replaces the given line range with lines, preserving all bytes
// outside the span. the replacement lines do not include newline separators.
func ReplaceSpan(raw []byte, span LineSpan, replacement []string) []byte {
	lines := strings.Split(string(raw), "\n")
	lines = slices.Replace(lines, span.Start, span.End+1, replacement...)
	return []byte(strings.Join(lines, "\n"))
}

// ValidDate reports whether value is a calendar date in YYYY-MM-DD form.
func ValidDate(value string) bool { return validDate(value) }
