package server

import (
	"context"
	"fmt"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/v2/model"
	v2 "github.com/michaelquigley/ranger/internal/v2/workspace"
)

func (s *Server) GetV2Project(_ context.Context, params api.GetV2ProjectParams) (api.GetV2ProjectRes, error) {
	project, notFound, wrong, err := s.resolve(params.Project, v2.V2)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	if wrong != nil {
		return wrong, nil
	}
	snap, err := v2.Load(project.Root())
	if err != nil {
		return nil, err
	}
	git := v2.GitStatus(project.Root())
	out := &api.V2Project{
		Project: params.Project,
		Roadmap: wireV2View(snap.Roadmap, snap.Project.Roadmap, git, model.Roadmap),
		Queue:   wireV2View(snap.Queue, snap.Project.Queue, git, model.Queue),
		Items:   make(api.V2ProjectItems, len(snap.Items)),
		Needs:   wireV2Needs(snap.Project.Needs),
		Flags:   wireV2Flags(snap.Project.Flags),
	}
	for slug, item := range snap.Items {
		out.Items[slug] = wireV2Summary(item, snap.Project.Items[slug], wireV2Git(git.Known, git.Items[slug]))
	}
	return out, nil
}

func (s *Server) GetV2Item(_ context.Context, params api.GetV2ItemParams) (api.GetV2ItemRes, error) {
	project, notFound, wrong, err := s.resolve(params.Project, v2.V2)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	if wrong != nil {
		return wrong, nil
	}
	snap, err := v2.Load(project.Root())
	if err != nil {
		return nil, err
	}
	item, ok := snap.Items[params.Slug]
	if !ok {
		return &api.ErrorResponse{Message: fmt.Sprintf("no item named %s", params.Slug)}, nil
	}
	git := v2.GitStatus(project.Root())
	summary := snap.Project.Items[params.Slug]
	out := &api.V2Item{
		Content: string(item.Raw), Hash: item.Hash,
		Summary:    wireV2Summary(item, summary, wireV2Git(git.Known, git.Items[params.Slug])),
		Subsystems: item.Subsystems, Source: optText(item.Source), Milestone: optText(item.Milestone),
		LinksOut: wireV2Links(summary.LinksOut), LinksIn: wireV2Links(summary.LinksIn),
		HasDesign: item.HasDesign, PlanError: optText(item.PlanError),
		Stages: make(api.V2ItemStages, len(item.Stages)),
		Files:  append([]string{}, item.Files...),
	}
	for _, entry := range item.Log {
		out.Log = append(out.Log, api.LogEntry{Stamp: entry.Stamp, Note: entry.Note})
	}
	if item.Plan != nil {
		out.Plan = api.NewOptV2Plan(api.V2Plan{
			Mercurius: optText(item.Plan.Mercurius),
			Stages:    append([]string{}, item.Plan.Stages...),
		})
	}
	for name, stage := range item.Stages {
		out.Stages[name] = api.V2Stage{
			Title: optText(stage.Title), Status: wireV2StageStatus(stage.Status),
			Terminus: optText(stage.Terminus), Needs: append([]string{}, stage.Needs...),
			Accepted: optText(stage.Accepted), Error: optText(stage.Error),
		}
	}
	return out, nil
}
