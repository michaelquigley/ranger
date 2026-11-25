package document

import (
	"fmt"

	shared "github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

var itemFields = []string{
	"title", "state", "created", "since", "kind", "tags", "subsystems", "source",
	"milestone", "log", "includes", "requires", "blocked_on", "needs",
}

type itemSchema struct {
	Title      string
	State      string
	Created    string
	Since      string
	Kind       string
	Tags       []string
	Subsystems []string
	Source     string
	Milestone  string
	Log        []model.LogEntry
	Includes   []string
	Requires   []string
	BlockedOn  []string
	Needs      []string
}

// ParseItem returns the readable fields and the body even when some fields
// fail validation. slug is the directory name used to resolve body links.
func ParseItem(raw []byte, slug string) (model.Item, string) {
	front, body, err := splitFrontmatter(raw, true)
	checklist, targets := scanBody(body, slug)
	item := model.Item{Checklist: checklist, BodyTargets: targets}
	if err != nil {
		item.Diagnostics = []string{err.Error()}
		return item, body
	}

	var schema itemSchema
	readable, diagnostics := bindFields(front, &schema, itemFields, []string{"title", "state", "created"})
	item.Diagnostics = diagnostics
	item.Title, item.TitleOK = schema.Title, readable["title"]
	item.Tags, item.Subsystems = schema.Tags, schema.Subsystems
	item.Source, item.Milestone = schema.Source, schema.Milestone
	item.Log, item.Needs = schema.Log, schema.Needs
	if readable["state"] {
		state, ok := model.ParseState(schema.State)
		if !ok {
			item.Diagnostics = append(item.Diagnostics, fmt.Sprintf("invalid state: %q", schema.State))
		}
		item.State = state
	}
	if readable["kind"] {
		kind, ok := model.ParseKind(schema.Kind)
		if !ok {
			item.Diagnostics = append(item.Diagnostics, fmt.Sprintf("invalid kind: %q", schema.Kind))
		}
		item.Kind = kind
	}
	if readable["created"] {
		item.Created = itemDate(&item, "created", schema.Created)
	}
	if readable["since"] {
		item.Since = itemDate(&item, "since", schema.Since)
	}
	for _, entry := range item.Log {
		if !shared.ValidDate(entry.Stamp) {
			item.Diagnostics = append(item.Diagnostics, fmt.Sprintf("invalid log stamp: %q", entry.Stamp))
		}
	}
	item.Includes = itemTargets(&item, model.Includes, schema.Includes)
	item.Requires = itemTargets(&item, model.Requires, schema.Requires)
	item.BlockedOn = itemTargets(&item, model.BlockedOn, schema.BlockedOn)
	return item, body
}

func itemDate(item *model.Item, field, value string) string {
	if shared.ValidDate(value) {
		return value
	}
	item.Diagnostics = append(item.Diagnostics, fmt.Sprintf("invalid %s date: %q", field, value))
	return ""
}

func itemTargets(item *model.Item, relation model.Relation, values []string) []model.Target {
	var targets []model.Target
	for _, value := range values {
		target, ok := model.ParseTarget(value)
		if !ok {
			item.Diagnostics = append(item.Diagnostics, fmt.Sprintf("invalid %s target: %q", relation, value))
			continue
		}
		targets = append(targets, target)
	}
	return targets
}
