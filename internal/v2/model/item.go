// Package model computes the v2 views, relationships, progress, and flags
// from item facts and view membership.
package model

import "strings"

type State string

const (
	Inbox      State = "inbox"
	Horizon    State = "horizon"
	Pending    State = "pending"
	Building   State = "building"
	Evaluating State = "evaluating"
)

var laneOrder = [...]State{Inbox, Horizon, Pending, Building, Evaluating}

func ParseState(value string) (State, bool) {
	for _, state := range laneOrder {
		if string(state) == value {
			return state, true
		}
	}
	return "", false
}

type Kind string

const (
	Epic      Kind = "epic"
	Story     Kind = "story"
	Defect    Kind = "defect"
	Punchlist Kind = "punchlist"
)

func ParseKind(value string) (Kind, bool) {
	switch kind := Kind(value); kind {
	case Epic, Story, Defect, Punchlist:
		return kind, true
	default:
		return "", false
	}
}

type LogEntry struct {
	Stamp string
	Note  string
}

type Checklist struct {
	Open  int
	Total int
}

type Item struct {
	Title      string
	TitleOK    bool
	State      State
	Created    string
	Since      string
	Kind       Kind
	Tags       []string
	Subsystems []string
	Source     string
	Milestone  string
	Log        []LogEntry
	Includes   []Target
	Requires   []Target
	BlockedOn  []Target
	Needs      []string

	Checklist   Checklist
	BodyTargets []string
	Diagnostics []string

	HasDesign bool
	Plan      *Plan
	PlanError string
	Stages    map[string]Stage
}

func (i Item) Lane() State {
	if i.State == "" {
		return Inbox
	}
	return i.State
}

type Plan struct {
	Mercurius string
	Stages    []string
}

type StageStatus string

const (
	Started StageStatus = "started"
	Built   StageStatus = "built"
)

type Stage struct {
	Title    string
	Status   StageStatus
	Terminus string
	Needs    []string
	Accepted string
	Error    string
}

type Target struct {
	Project string
	Slug    string
}

func ParseTarget(value string) (Target, bool) {
	project, slug, crossProject := strings.Cut(value, ":")
	if !crossProject {
		return Target{Slug: value}, value != ""
	}
	return Target{Project: project, Slug: slug}, project != "" && slug != ""
}

func (t Target) String() string {
	if t.Project == "" {
		return t.Slug
	}
	return t.Project + ":" + t.Slug
}

type Relation string

const (
	Includes  Relation = "includes"
	Requires  Relation = "requires"
	BlockedOn Relation = "blocked_on"
	BodyLink  Relation = "body"
)

type Link struct {
	From     string
	Relation Relation
	Target   Target
}
