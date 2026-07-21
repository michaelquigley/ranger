// Package server implements the generated api.Handler over a project set.
// every handler call resolves its project and rebuilds from a fresh disk
// read — the server holds no snapshot — and translation between model and
// wire types happens at this edge only.
package server

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/model"
	"github.com/michaelquigley/ranger/internal/workspace"
)

// Server is the api.Handler implementation: a project set and nothing else.
type Server struct {
	projects *Projects
}

// New returns a server over the given project set.
func New(projects *Projects) *Server {
	return &Server{projects: projects}
}

var _ api.Handler = (*Server)(nil)

// NewError shapes any unmapped handler error as the default response —
// repository-level failures, partial two-file reports, and the like, the
// message verbatim.
func (s *Server) NewError(_ context.Context, err error) *api.ServerErrorStatusCode {
	return &api.ServerErrorStatusCode{
		StatusCode: 500,
		Response:   api.ErrorResponse{Message: err.Error()},
	}
}

// resolve maps the {project} path segment to its workspace. an unknown
// name returns the operation's 404 body; a config source failure is the
// request's plain error, healed on the next good save.
func (s *Server) resolve(name string) (*workspace.Workspace, *api.ErrorResponse, error) {
	w, err := s.projects.Resolve(name)
	if err != nil {
		var unknown *UnknownProjectError
		if errors.As(err, &unknown) {
			return nil, &api.ErrorResponse{Message: err.Error()}, nil
		}
		return nil, nil, err
	}
	return w, nil, nil
}

// asConflict maps the document layer's typed refusals onto the wire's 409
// family: guard mismatches split item/order by the file they guard, and
// no-clobber collisions carry their structured recovery paths.
func asConflict(err error) (*api.Conflict, bool) {
	var conflict *document.ConflictError
	if errors.As(err, &conflict) {
		reason := api.ConflictReasonItemConflict
		if filepath.Base(conflict.Path) == "order.yaml" {
			reason = api.ConflictReasonOrderConflict
		}
		return &api.Conflict{Reason: reason, Message: err.Error()}, true
	}
	var collision *document.CollisionError
	if errors.As(err, &collision) {
		return &api.Conflict{
			Reason:     api.ConflictReasonSlugCollision,
			Message:    err.Error(),
			SourcePath: api.NewOptString(collision.Src),
			DestPath:   api.NewOptString(collision.Dst),
		}, true
	}
	return nil, false
}

func wireCard(snap *workspace.Snapshot, c model.CardInput) api.Card {
	card := api.Card{
		Filename: c.Filename,
		Title:    c.Title,
		Flags:    []api.Flag{},
	}
	if c.State != "" {
		card.State = api.NewOptState(api.State(c.State))
	}
	if c.Created != "" {
		card.Created = api.NewOptString(c.Created)
	}
	for _, f := range c.Flags {
		card.Flags = append(card.Flags, api.Flag{Kind: api.FlagKind(f.Kind), Diagnostic: f.Diagnostic})
	}
	if it, ok := snap.Item(c.Filename); ok {
		card.Hash = it.Hash
		card.Tags = it.Doc.Tags
		card.Subsystems = it.Doc.Subsystems
		if it.Doc.Source != "" {
			card.Source = api.NewOptString(it.Doc.Source)
		}
		if it.Doc.Milestone != "" {
			card.Milestone = api.NewOptString(it.Doc.Milestone)
		}
		for _, l := range it.Doc.Log {
			card.Log = append(card.Log, api.LogEntry{Stamp: l.Stamp, Note: l.Note})
		}
	}
	return card
}

func wireBoard(snap *workspace.Snapshot, project string) *api.Board {
	board := snap.Board()
	out := &api.Board{Project: project, OrderVersion: snap.OrderVersion}
	for _, lane := range board.Lanes {
		wl := api.Lane{
			State:       api.State(lane.State),
			Cards:       []api.Card{},
			RankedCount: lane.RankedCount,
		}
		for _, c := range lane.Cards {
			wl.Cards = append(wl.Cards, wireCard(snap, c))
		}
		out.Lanes = append(out.Lanes, wl)
	}
	return out
}

// freshBoard reloads from disk after a mutation so the client repaints from
// disk truth, never from what the mutation thinks it did. project is the
// configured name, carried onto the wire.
func freshBoard(w *workspace.Workspace, project string) (*api.Board, error) {
	snap, err := w.Load()
	if err != nil {
		return nil, err
	}
	return wireBoard(snap, project), nil
}
