package model

import (
	"slices"
	"testing"
)

func TestViewOrder(t *testing.T) {
	items := map[string]Item{
		"inbox-later":      {State: Inbox, Created: "2026-11-02"},
		"inbox-a":          {State: Inbox, Created: "2026-11-01"},
		"inbox-z":          {Created: "2026-11-01"},
		"undated":          {},
		"horizon-a":        {State: Horizon},
		"horizon-b":        {State: Horizon},
		"pending-a":        {State: Pending},
		"pending-b":        {State: Pending},
		"building-a":       {State: Building, Since: "2026-11-04"},
		"building-z":       {State: Building, Since: "2026-11-04"},
		"building-old":     {State: Building, Since: "2026-11-02"},
		"building-undated": {State: Building},
		"evaluating-new":   {State: Evaluating, Since: "2026-11-07"},
		"evaluating-old":   {State: Evaluating, Since: "2026-11-01"},
	}
	membership := func(names ...string) Membership {
		view := Membership{Readable: true}
		for i, name := range names {
			view.Entries = append(view.Entries, ViewEntry{Slug: name, Line: i + 1})
		}
		return view
	}
	project := ComputeProject(items, nil,
		membership("pending-b", "pending-a", "pending-b", "horizon-b", "horizon-a", "inbox-later",
			"building-z", "building-undated", "building-old", "building-a", "evaluating-old", "evaluating-new"),
		membership("horizon-a", "horizon-b", "pending-a", "pending-b", "building-z", "building-a", "evaluating-old"),
	)
	inbox := []string{"inbox-a", "inbox-z", "inbox-later", "undated"}
	for _, tt := range []struct {
		name  ViewName
		lanes []Lane
		want  [][]string
	}{
		{Roadmap, project.Roadmap, [][]string{inbox, {"horizon-b", "horizon-a"}, {"pending-b", "pending-a"},
			{"building-a", "building-z", "building-old", "building-undated"}, {"evaluating-new", "evaluating-old"}}},
		{Queue, project.Queue, [][]string{inbox, {"horizon-a", "horizon-b"}, {"pending-a", "pending-b"},
			{"building-a", "building-z"}, {"evaluating-old"}}},
	} {
		t.Run(string(tt.name), func(t *testing.T) {
			wantStates := []State{Inbox, Horizon, Pending, Building, Evaluating}
			if len(tt.lanes) != len(wantStates) {
				t.Fatalf("lanes = %v", tt.lanes)
			}
			for i, lane := range tt.lanes {
				if lane.State != wantStates[i] || !slices.Equal(lane.Items, tt.want[i]) {
					t.Errorf("lane %d = %v, want %s %v", i, lane, wantStates[i], tt.want[i])
				}
			}
		})
	}
}

func TestProjectFlagsAndNeeds(t *testing.T) {
	items := map[string]Item{
		"origin": {
			State: Building, Title: "changed title", TitleOK: true,
			Diagnostics: []string{"invalid kind"}, HasDesign: true,
			Plan: &Plan{Stages: []string{"main", "missing", "broken"}},
			Stages: map[string]Stage{
				"main":   {Title: "Main", Status: Built, Needs: []string{"audition"}},
				"extra":  {Needs: []string{"a request from an unlisted stage"}},
				"broken": {Error: "invalid date", Status: Built},
			},
			Needs:       []string{"choose a direction"},
			Includes:    []Target{{Slug: "target"}},
			Requires:    []Target{{Project: "other", Slug: "gone"}},
			BlockedOn:   []Target{{Slug: "gone"}},
			BodyTargets: []string{"target", "gone", "notes"},
		},
		"target": {},
		"done": {
			State: Building, Plan: &Plan{Stages: []string{"last"}},
			Stages: map[string]Stage{"last": {Accepted: "2026-11-24", Status: Built}},
		},
		"bad-plan": {State: Building, PlanError: "invalid YAML"},
	}
	roadmap := Membership{Readable: true, Entries: []ViewEntry{{Slug: "gone", Line: 2}, {Slug: "origin", Line: 3}, {Slug: "done", Line: 4}}}
	project := ComputeProject(items, map[string]bool{"notes": true}, roadmap, Membership{Readable: true})
	var gotKinds []FlagKind
	for _, flag := range project.Flags {
		gotKinds = append(gotKinds, flag.Kind)
	}
	wantKinds := []FlagKind{MissingViewItem, UnreadablePlan, InNeitherView, StillBuilding,
		Malformed, SlugMismatch, MissingStage, UnreadableStage, UnlistedStage, MissingLink, MissingLink}
	if !slices.Equal(gotKinds, wantKinds) {
		t.Errorf("flags = %+v, want kinds %v", project.Flags, wantKinds)
	}
	if project.Flags[0] != (Flag{Kind: MissingViewItem, View: Roadmap, Line: 2, Target: "gone"}) {
		t.Errorf("view flag = %+v", project.Flags[0])
	}

	wantLinks := []Link{
		{From: "origin", Relation: Includes, Target: Target{Slug: "target"}},
		{From: "origin", Relation: BodyLink, Target: Target{Slug: "target"}},
	}
	if links := project.Items["target"].LinksIn; !slices.Equal(links, wantLinks) {
		t.Errorf("incoming links = %+v", links)
	}
	linksOut := project.Items["origin"].LinksOut
	if len(linksOut) != 5 || linksOut[1].Target.String() != "other:gone" {
		t.Errorf("outgoing links = %+v", linksOut)
	}
	wantNeeds := []Need{
		{Item: "origin", Request: "choose a direction"},
		{Item: "origin", Stage: "main", Request: "audition"},
		{Item: "origin", Stage: "main", Acceptance: true},
		{Item: "origin", Stage: "extra", Request: "a request from an unlisted stage"},
	}
	if !slices.Equal(project.Needs, wantNeeds) || !slices.Equal(project.Items["origin"].Needs, wantNeeds) {
		t.Errorf("needs = %+v", project.Needs)
	}
	for _, p := range []Project{
		ComputeProject(items, nil, roadmap, Membership{}),
		ComputeProject(items, nil, Membership{}, roadmap),
	} {
		if slices.ContainsFunc(p.Flags, func(flag Flag) bool { return flag.Kind == InNeitherView }) {
			t.Errorf("membership unknown but flagged: %+v", p.Flags)
		}
	}
}
