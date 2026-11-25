package server

import (
	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/v2/model"
	v2 "github.com/michaelquigley/ranger/internal/v2/workspace"
)

func wireV2View(view v2.ViewFile, lanes []model.Lane, git v2.ProjectGitStatus, name model.ViewName) api.V2View {
	out := api.V2View{
		Lanes:   make([]api.V2Lane, 0, len(lanes)),
		Version: optText(view.Version),
		Git:     wireV2Git(git.Known, git.Views[name]),
	}
	if view.Err != nil {
		out.Error = api.NewOptString(view.Err.Error())
	}
	for _, lane := range lanes {
		out.Lanes = append(out.Lanes, api.V2Lane{State: api.V2State(lane.State), Items: append([]string{}, lane.Items...)})
	}
	return out
}

func wireV2Summary(item v2.Item, summary *model.ItemSummary, git api.OptV2GitMark) api.V2ItemSummary {
	out := api.V2ItemSummary{
		Title: item.Title, State: api.V2State(item.Lane()),
		Created: optText(item.Created), Since: optText(item.Since),
		Tags:     append([]string{}, item.Tags...),
		Includes: wireV2Targets(item.Includes), Requires: wireV2Targets(item.Requires), BlockedOn: wireV2Targets(item.BlockedOn),
		Checklist: api.V2Checklist{Open: item.Checklist.Open, Total: item.Checklist.Total},
		Needs:     wireV2Needs(summary.Needs), Flags: wireV2Flags(summary.Flags), Git: git,
	}
	if item.Kind != "" {
		out.Kind = api.NewOptV2ItemSummaryKind(api.V2ItemSummaryKind(item.Kind))
	}
	if summary.Progress.Position != "" {
		out.Progress = api.NewOptV2Progress(wireV2Progress(summary.Progress))
	}
	return out
}

func wireV2Targets(targets []model.Target) []string {
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		out = append(out, target.String())
	}
	return out
}

func wireV2Progress(progress model.Progress) api.V2Progress {
	out := api.V2Progress{
		Position: api.V2ProgressPosition(progress.Position), Stage: optText(progress.Stage),
		Status: wireV2StageStatus(progress.Status),
	}
	if progress.Stage != "" {
		out.Number = api.NewOptInt(progress.Number)
		out.Total = api.NewOptInt(progress.Total)
	}
	if progress.Verdict.Broker != "" {
		out.Verdict = api.NewOptV2Verdict(api.V2Verdict{
			Broker: api.V2VerdictBroker(progress.Verdict.Broker), Value: progress.Verdict.Value,
		})
	}
	return out
}

func wireV2StageStatus(status model.StageStatus) api.OptV2StageStatus {
	if status == "" {
		return api.OptV2StageStatus{}
	}
	return api.NewOptV2StageStatus(api.V2StageStatus(status))
}

func wireV2Needs(needs []model.Need) []api.V2Need {
	out := make([]api.V2Need, 0, len(needs))
	for _, need := range needs {
		out = append(out, api.V2Need{
			Item: need.Item, Stage: optText(need.Stage), Request: optText(need.Request), Acceptance: need.Acceptance,
		})
	}
	return out
}

func wireV2Flags(flags []model.Flag) []api.V2Flag {
	out := make([]api.V2Flag, 0, len(flags))
	for _, flag := range flags {
		wf := api.V2Flag{
			Kind: api.V2FlagKind(flag.Kind), Item: optText(flag.Item),
			Target: optText(flag.Target), Diagnostic: optText(flag.Diagnostic),
		}
		if flag.View != "" {
			wf.View = api.NewOptV2FlagView(api.V2FlagView(flag.View))
			wf.Line = api.NewOptInt(flag.Line)
		}
		if flag.Relation != "" {
			wf.Relation = api.NewOptV2Relation(api.V2Relation(flag.Relation))
		}
		out = append(out, wf)
	}
	return out
}

func wireV2Links(links []model.Link) []api.V2Link {
	out := make([]api.V2Link, 0, len(links))
	for _, link := range links {
		out = append(out, api.V2Link{
			From: link.From, Relation: api.V2Relation(link.Relation),
			Slug: link.Target.Slug, Project: optText(link.Target.Project),
		})
	}
	return out
}

func wireV2Git(known bool, mark v2.GitMark) api.OptV2GitMark {
	if !known {
		return api.OptV2GitMark{}
	}
	return api.NewOptV2GitMark(api.V2GitMark{Staged: mark.Staged, Unstaged: mark.Unstaged, New: mark.New})
}

func optText(text string) api.OptString {
	if text == "" {
		return api.OptString{}
	}
	return api.NewOptString(text)
}
