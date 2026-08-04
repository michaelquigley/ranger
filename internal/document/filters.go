package document

import (
	"fmt"
	"strings"

	"github.com/michaelquigley/df/dd"
)

// SavedFilter is one named board filter from .ranger/filters.yaml: the
// operator's saved lens over the board, each dimension carrying include
// and exclude values plus an absence flag — match only cards where the
// dimension is not specified at all. empty milestone fields mean unset,
// matching the item schema's convention.
type SavedFilter struct {
	Name          string
	Tags          []string
	NotTags       []string
	NoTags        bool
	Subsystems    []string
	NotSubsystems []string
	NoSubsystems  bool
	Milestone     string
	NotMilestone  string
	NoMilestone   bool
}

// filtersSchema is the dd-bind target for the whole document.
type filtersSchema struct {
	Filters []SavedFilter
}

// ParseFilters parses raw as a filters document. unlike items, which
// degrade field-by-field to flagged cards, a filters file reads whole or
// not at all — it is tool-rendered and small, and a filter missing its
// name has no identity to degrade to. empty and duplicate names refuse
// for the same reason duplicate lane keys do in order.yaml: guessing
// which entry the name means is coin-flipping.
func ParseFilters(raw []byte) ([]SavedFilter, error) {
	var schema filtersSchema
	if err := dd.BindYAML(&schema, raw); err != nil {
		return nil, fmt.Errorf("filters document does not parse: %w", err)
	}
	seen := map[string]bool{}
	for _, f := range schema.Filters {
		if f.Name == "" {
			return nil, fmt.Errorf("filter with empty name")
		}
		if seen[f.Name] {
			return nil, fmt.Errorf("duplicate filter name: %s", f.Name)
		}
		seen[f.Name] = true
	}
	return schema.Filters, nil
}

// RenderFilters emits a filters document fresh. filters.yaml is
// tool-rendered: unlike items and order.yaml, whose bytes the operator
// authors and the tool patches surgically, this file is born from a board
// gesture and re-rendered whole — deterministically, by hand, never
// through a YAML encoder — on every save. a hand edit is honored on read;
// the next save re-renders the file in canonical form.
func RenderFilters(filters []SavedFilter) []byte {
	if len(filters) == 0 {
		return []byte("filters: []\n")
	}
	var b strings.Builder
	b.WriteString("filters:\n")
	for _, f := range filters {
		b.WriteString("  - name: " + emitScalar(f.Name) + "\n")
		renderList(&b, "tags", f.Tags)
		renderList(&b, "not_tags", f.NotTags)
		renderList(&b, "subsystems", f.Subsystems)
		renderList(&b, "not_subsystems", f.NotSubsystems)
		renderScalar(&b, "milestone", f.Milestone)
		renderScalar(&b, "not_milestone", f.NotMilestone)
	}
	return []byte(b.String())
}

// renderList emits one flow-style list field — the same form card
// frontmatter carries tags in — omitting empty lists entirely.
func renderList(b *strings.Builder, key string, values []string) {
	if len(values) == 0 {
		return
	}
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = emitScalar(v)
	}
	b.WriteString("    " + key + ": [" + strings.Join(quoted, ", ") + "]\n")
}

func renderScalar(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	b.WriteString("    " + key + ": " + emitScalar(value) + "\n")
}
