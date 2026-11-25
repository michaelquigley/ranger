package model

import "strings"

type Position string

const (
	NoDesign     Position = "no-design"
	Designing    Position = "designing"
	PlanInReview Position = "plan-in-review"
	Ready        Position = "ready-to-hand-off"
	NoStages     Position = "no-stages"
	CurrentStage Position = "current-stage"
	Complete     Position = "complete"
)

type Broker string

const (
	Mercurius Broker = "mercurius"
	Terminus  Broker = "terminus"
)

type Verdict struct {
	Broker Broker
	Value  string
}

type Progress struct {
	Position Position
	Stage    string
	Number   int
	Total    int
	Status   StageStatus
	Verdict  Verdict
}

// BuildingProgress follows the plan's order, including missing or invalid
// stages, which hold their places without contributing any facts.
func BuildingProgress(item Item) Progress {
	if item.Plan == nil {
		if item.HasDesign {
			return Progress{Position: Designing}
		}
		return Progress{Position: NoDesign}
	}

	plan := item.Plan
	started := false
	firstUnaccepted := -1
	for i, name := range plan.Stages {
		stage := item.Stages[name]
		if stage.Error != "" {
			stage = Stage{}
		}
		started = started || stage.Status != "" || stage.Accepted != ""
		if firstUnaccepted == -1 && stage.Accepted == "" {
			firstUnaccepted = i
		}
	}
	if len(plan.Stages) > 0 && firstUnaccepted == -1 {
		return Progress{Position: Complete}
	}
	if started {
		name := plan.Stages[firstUnaccepted]
		stage := item.Stages[name]
		if stage.Error != "" {
			stage = Stage{}
		}
		return Progress{
			Position: CurrentStage, Stage: name, Number: firstUnaccepted + 1,
			Total: len(plan.Stages), Status: stage.Status,
			Verdict: postedVerdict(Terminus, stage.Terminus),
		}
	}

	position := PlanInReview
	switch {
	case len(plan.Stages) == 0:
		position = NoStages
	case strings.HasPrefix(plan.Mercurius, "ready_to_build"):
		position = Ready
	}
	return Progress{Position: position, Verdict: postedVerdict(Mercurius, plan.Mercurius)}
}

func postedVerdict(broker Broker, value string) Verdict {
	if value == "" {
		return Verdict{}
	}
	return Verdict{Broker: broker, Value: value}
}

type Need struct {
	Item       string
	Stage      string
	Request    string
	Acceptance bool
}

func itemNeeds(slug string, item Item, stages []string) []Need {
	var needs []Need
	for _, request := range item.Needs {
		needs = append(needs, Need{Item: slug, Request: request})
	}
	for _, name := range stages {
		stage := item.Stages[name]
		if stage.Error != "" {
			continue
		}
		for _, request := range stage.Needs {
			needs = append(needs, Need{Item: slug, Stage: name, Request: request})
		}
		if stage.Status == Built && stage.Accepted == "" {
			needs = append(needs, Need{Item: slug, Stage: name, Acceptance: true})
		}
	}
	return needs
}
