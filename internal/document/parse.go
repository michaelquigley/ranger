package document

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// the two-pass parse shared by both document kinds. pass 1 decodes the whole
// YAML block into a yaml.Node AST — purely syntactic, so permitted duplicate
// keys don't trip it. pass 2 walks that AST, never re-parsed text: claimed
// keys are recognized and duplicate-checked among themselves, their value
// nodes decode into plain values (aliases resolving against the full tree),
// and unknown material is skipped so its duplicates and anchors are never
// disturbed. the line scanner does the one thing only it can: the key →
// line-range map for surgical write spans.

// keyEntry is one top-level mapping pair with its nodes.
type keyEntry struct {
	name string // "" for a non-scalar key
	key  *yaml.Node
	val  *yaml.Node
}

// parseMapping runs pass 1 over yamlText and returns the top-level mapping's
// key entries in document order. a nil mapping with nil error means an empty
// document.
func parseMapping(yamlText string) ([]keyEntry, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(yamlText), &root); err != nil {
		return nil, err
	}
	if root.Kind == 0 || len(root.Content) == 0 {
		return nil, nil
	}
	m := root.Content[0]
	if m.Kind != yaml.MappingNode {
		return nil, errors.New("document is not a mapping")
	}
	entries := make([]keyEntry, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		name := ""
		if k.Kind == yaml.ScalarNode {
			name = k.Value
		}
		entries = append(entries, keyEntry{name: name, key: k, val: v})
	}
	return entries, nil
}

// fieldSpan is one top-level field's surgical write span: inclusive line
// indices into the file's line slice, plus what a rewrite must preserve.
type fieldSpan struct {
	start, end int
	indent     string
	comment    string // inline comment on the key line, "#..." or ""
}

// scanSpans computes each entry's line span. offset converts a YAML-text
// line number (1-based) into an index in the file's line slice. lastLine is
// the index of the final YAML line in the file. trailing blank lines are
// trimmed from every span; trailing comment-only lines are trimmed too,
// except lines indented as a block scalar's content, which only look like
// comments.
func scanSpans(lines []string, entries []keyEntry, offset, lastLine int) []fieldSpan {
	spans := make([]fieldSpan, len(entries))
	for i, e := range entries {
		start := e.key.Line + offset
		end := lastLine
		if i+1 < len(entries) {
			end = entries[i+1].key.Line + offset - 1
		}
		contentIndent := -1
		if e.val.Style == yaml.LiteralStyle || e.val.Style == yaml.FoldedStyle {
			contentIndent = blockContentIndent(lines, start, end, e.val)
		}
		end = trimSpanEnd(lines, start, end, contentIndent)
		comment := e.val.LineComment
		if comment == "" {
			comment = e.key.LineComment
		}
		spans[i] = fieldSpan{start: start, end: end, indent: leadingSpace(lines[start]), comment: comment}
	}
	return spans
}

// blockContentIndent finds a block scalar's content indentation: an explicit
// indentation indicator on the header (`|2`, relative to the key's indent)
// when present, else the indent of the first non-blank line after the key
// line, provided it is deeper than the key — a shallower line means the
// scalar is empty and nothing follows as content.
func blockContentIndent(lines []string, start, end int, val *yaml.Node) int {
	keyIndent := len(leadingSpace(lines[start]))
	if val.Column-1 >= 0 && val.Column-1 < len(lines[start]) {
		header := lines[start][val.Column-1:]
		if len(header) > 0 && (header[0] == '|' || header[0] == '>') {
			for _, c := range []byte(header[1:min(3, len(header))]) {
				if c >= '1' && c <= '9' {
					return keyIndent + int(c-'0')
				}
				if c != '+' && c != '-' {
					break
				}
			}
		}
	}
	for j := start + 1; j <= end; j++ {
		if strings.TrimSpace(lines[j]) == "" {
			continue
		}
		if ind := len(leadingSpace(lines[j])); ind > keyIndent {
			return ind
		}
		break
	}
	return -1
}

// trimSpanEnd walks end back over trailing blank lines and comment-only
// lines. contentIndent >= 0 marks a block scalar's content indentation:
// comment-looking lines at least that deep are scalar content and stay.
func trimSpanEnd(lines []string, start, end, contentIndent int) int {
	for end > start {
		t := strings.TrimSpace(lines[end])
		if t == "" {
			end--
			continue
		}
		if strings.HasPrefix(t, "#") && (contentIndent < 0 || len(leadingSpace(lines[end])) < contentIndent) {
			end--
			continue
		}
		break
	}
	return end
}

func leadingSpace(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// nodeToPlain decodes a value node into a plain value against the full tree:
// aliases follow their anchor, scalars become their source text (the
// convention's claimed fields are all string-shaped, and source text is what
// shape validation must judge), sequences and mappings recurse. the depth
// guard bounds pathological alias constructions.
func nodeToPlain(n *yaml.Node, depth int) (any, error) {
	if depth > 100 {
		return nil, errors.New("node nesting too deep")
	}
	switch n.Kind {
	case yaml.AliasNode:
		return nodeToPlain(n.Alias, depth+1)
	case yaml.ScalarNode:
		if n.Tag == "!!null" {
			return nil, nil
		}
		return n.Value, nil
	case yaml.SequenceNode:
		out := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := nodeToPlain(c, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.MappingNode:
		out := make(map[string]any, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, err := nodeToPlain(n.Content[i], depth+1)
			if err != nil {
				return nil, err
			}
			ks, ok := k.(string)
			if !ok {
				return nil, errors.New("non-scalar mapping key")
			}
			v, err := nodeToPlain(n.Content[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			out[ks] = v
		}
		return out, nil
	}
	return nil, errors.New("unsupported node kind")
}

// collectAnchors gathers every node in the subtree that defines an anchor.
func collectAnchors(n *yaml.Node, into map[*yaml.Node]bool) {
	if n == nil {
		return
	}
	if n.Anchor != "" {
		into[n] = true
	}
	for _, c := range n.Content {
		collectAnchors(c, into)
	}
}

// aliasesInto reports whether the subtree contains an alias resolving to one
// of the given anchor nodes.
func aliasesInto(n *yaml.Node, anchors map[*yaml.Node]bool) bool {
	if n == nil {
		return false
	}
	if n.Kind == yaml.AliasNode && anchors[n.Alias] {
		return true
	}
	for _, c := range n.Content {
		if aliasesInto(c, anchors) {
			return true
		}
	}
	return false
}
