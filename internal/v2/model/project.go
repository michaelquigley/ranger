package model

import (
	"cmp"
	"maps"
	"slices"

	"github.com/michaelquigley/ranger/internal/slug"
)

type ViewName string

const (
	Roadmap ViewName = "roadmap"
	Queue   ViewName = "queue"
)

type ViewEntry struct {
	Slug string
	Line int
}

type Membership struct {
	Entries  []ViewEntry
	Readable bool
}

type Lane struct {
	State State
	Items []string
}

type FlagKind string

const (
	Malformed       FlagKind = "malformed"
	SlugMismatch    FlagKind = "slug-mismatch"
	StillBuilding   FlagKind = "still-building"
	MissingLink     FlagKind = "missing-link"
	MissingViewItem FlagKind = "missing-view-item"
	InNeitherView   FlagKind = "in-neither-view"
	MissingStage    FlagKind = "missing-stage"
	UnlistedStage   FlagKind = "unlisted-stage"
	UnreadablePlan  FlagKind = "unreadable-plan"
	UnreadableStage FlagKind = "unreadable-stage"
)

type Flag struct {
	Kind       FlagKind
	Item       string
	View       ViewName
	Line       int
	Target     string
	Relation   Relation
	Diagnostic string
}

type ItemSummary struct {
	Progress Progress
	LinksOut []Link
	LinksIn  []Link
	Needs    []Need
	Flags    []Flag
}

type Project struct {
	Roadmap []Lane
	Queue   []Lane
	Items   map[string]*ItemSummary
	Needs   []Need
	Flags   []Flag
}

// ComputeProject classifies both views and builds the item summaries and
// project-wide lists from the same facts.
func ComputeProject(items map[string]Item, directories map[string]bool, roadmap, queue Membership) Project {
	project := Project{Items: make(map[string]*ItemSummary, len(items))}
	slugs := slices.Sorted(maps.Keys(items))
	for _, slug := range slugs {
		project.Items[slug] = &ItemSummary{}
	}

	var roadmapMembers, queueMembers map[string]bool
	project.Roadmap, roadmapMembers = project.computeView(Roadmap, roadmap, items, slugs)
	project.Queue, queueMembers = project.computeView(Queue, queue, items, slugs)

	for _, slug := range slugs {
		item := items[slug]
		summary := project.Items[slug]
		stageNames := itemStageNames(item)
		summary.Needs = itemNeeds(slug, item, stageNames)
		project.Needs = append(project.Needs, summary.Needs...)
		if item.Lane() == Building {
			summary.Progress = BuildingProgress(item)
		}
		project.classifyItem(slug, item, stageNames)
		if roadmap.Readable && queue.Readable && item.Lane() != Inbox &&
			!roadmapMembers[slug] && !queueMembers[slug] {
			project.flag(Flag{Kind: InNeitherView, Item: slug})
		}
		for _, link := range itemLinks(slug, item, items, directories) {
			summary.LinksOut = append(summary.LinksOut, link)
			if link.Target.Project != "" {
				continue
			}
			if target := project.Items[link.Target.Slug]; target != nil {
				target.LinksIn = append(target.LinksIn, link)
			} else {
				project.flag(Flag{Kind: MissingLink, Item: slug, Target: link.Target.Slug, Relation: link.Relation})
			}
		}
	}
	return project
}

func (p *Project) computeView(name ViewName, view Membership, items map[string]Item, slugs []string) ([]Lane, map[string]bool) {
	if !view.Readable {
		return nil, nil
	}
	byState := map[State][]string{}
	seen := map[string]bool{}
	for _, entry := range view.Entries {
		item, exists := items[entry.Slug]
		if !exists {
			p.flag(Flag{Kind: MissingViewItem, View: name, Line: entry.Line, Target: entry.Slug})
			continue
		}
		if seen[entry.Slug] {
			continue
		}
		seen[entry.Slug] = true
		if state := item.Lane(); state != Inbox {
			byState[state] = append(byState[state], entry.Slug)
		}
	}
	for _, slug := range slugs {
		if items[slug].Lane() == Inbox {
			byState[Inbox] = append(byState[Inbox], slug)
		}
	}
	slices.SortFunc(byState[Inbox], func(a, b string) int {
		return dateOrder(items[a].Created, items[b].Created, a, b, false)
	})
	for _, state := range []State{Building, Evaluating} {
		slices.SortFunc(byState[state], func(a, b string) int {
			return dateOrder(items[a].Since, items[b].Since, a, b, true)
		})
	}

	lanes := make([]Lane, 0, len(laneOrder))
	for _, state := range laneOrder {
		lanes = append(lanes, Lane{State: state, Items: byState[state]})
	}
	return lanes, seen
}

func dateOrder(a, b, slugA, slugB string, newestFirst bool) int {
	switch {
	case a == b:
		return cmp.Compare(slugA, slugB)
	case a == "":
		return 1
	case b == "":
		return -1
	case newestFirst:
		return cmp.Compare(b, a)
	default:
		return cmp.Compare(a, b)
	}
}

func itemStageNames(item Item) []string {
	var names []string
	if item.Plan != nil {
		for _, name := range item.Plan.Stages {
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(item.Stages)) {
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	return names
}

func (p *Project) classifyItem(itemName string, item Item, stages []string) {
	for _, diagnostic := range item.Diagnostics {
		p.flag(Flag{Kind: Malformed, Item: itemName, Diagnostic: diagnostic})
	}
	if expected := slug.FromTitle(item.Title); item.TitleOK && expected != "" && expected != itemName {
		p.flag(Flag{Kind: SlugMismatch, Item: itemName, Target: expected})
	}
	if p.Items[itemName].Progress.Position == Complete {
		p.flag(Flag{Kind: StillBuilding, Item: itemName})
	}
	if item.PlanError != "" {
		p.flag(Flag{Kind: UnreadablePlan, Item: itemName, Diagnostic: item.PlanError})
	}
	for _, name := range stages {
		stage, exists := item.Stages[name]
		switch {
		case !exists:
			p.flag(Flag{Kind: MissingStage, Item: itemName, Target: name})
		case item.PlanError == "" && (item.Plan == nil || !slices.Contains(item.Plan.Stages, name)):
			p.flag(Flag{Kind: UnlistedStage, Item: itemName, Target: name})
		}
		if stage.Error != "" {
			p.flag(Flag{Kind: UnreadableStage, Item: itemName, Target: name, Diagnostic: stage.Error})
		}
	}
}

func (p *Project) flag(flag Flag) {
	p.Flags = append(p.Flags, flag)
	if flag.Item != "" {
		summary := p.Items[flag.Item]
		summary.Flags = append(summary.Flags, flag)
	}
}

func itemLinks(slug string, item Item, items map[string]Item, directories map[string]bool) []Link {
	var links []Link
	for _, group := range []struct {
		relation Relation
		targets  []Target
	}{
		{Includes, item.Includes},
		{Requires, item.Requires},
		{BlockedOn, item.BlockedOn},
	} {
		for _, target := range group.targets {
			links = append(links, Link{From: slug, Relation: group.relation, Target: target})
		}
	}
	for _, target := range item.BodyTargets {
		if _, itemExists := items[target]; itemExists || !directories[target] {
			links = append(links, Link{From: slug, Relation: BodyLink, Target: Target{Slug: target}})
		}
	}
	return links
}
