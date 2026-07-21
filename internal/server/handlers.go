package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/model"
	"github.com/michaelquigley/ranger/internal/workspace"
)

func (s *Server) GetProjects(_ context.Context) (*api.ProjectIndex, error) {
	idx, err := s.projects.Index()
	if err != nil {
		return nil, err
	}
	out := &api.ProjectIndex{Default: idx.Default, Projects: []api.ProjectStatus{}}
	for _, p := range idx.Projects {
		status := api.ProjectStatus{Name: p.Name, Available: p.Available}
		if p.Error != "" {
			status.Error = api.NewOptString(p.Error)
		}
		out.Projects = append(out.Projects, status)
	}
	return out, nil
}

func (s *Server) GetBoard(_ context.Context, params api.GetBoardParams) (api.GetBoardRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (s *Server) GetItem(_ context.Context, params api.GetItemParams) (api.GetItemRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	snap, err := w.Load()
	if err != nil {
		return nil, err
	}
	it, ok := snap.Item(params.Filename)
	if !ok {
		return &api.ErrorResponse{Message: fmt.Sprintf("no item named %s", params.Filename)}, nil
	}
	card := wireCard(snap, cardFor(snap, params.Filename))
	return &api.GetItemOK{Content: string(it.Raw), Card: card, Hash: it.Hash}, nil
}

// cardFor finds one item's classification among the snapshot's cards.
func cardFor(snap *workspace.Snapshot, filename string) model.CardInput {
	for _, c := range snap.Cards() {
		if c.Filename == filename {
			return c
		}
	}
	return model.CardInput{Filename: filename}
}

func (s *Server) SearchItems(_ context.Context, params api.SearchItemsParams) (api.SearchItemsRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	snap, err := w.Load()
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(params.Q)
	out := &api.SearchItemsOK{Filenames: []string{}}
	for _, it := range snap.Items {
		if strings.Contains(strings.ToLower(it.Doc.Title), q) || strings.Contains(strings.ToLower(it.Doc.Body()), q) {
			out.Filenames = append(out.Filenames, it.Filename)
		}
	}
	return out, nil
}

func (s *Server) CreateItem(_ context.Context, req *api.CreateItemReq, params api.CreateItemParams) (api.CreateItemRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return (*api.CreateItemNotFound)(notFound), nil
	}
	// prevalidated: no draft file exists for a title capture can never
	// finalize, so the form loses nothing and the tree stays clean.
	if req.Title == "" {
		return &api.CreateItemBadRequest{Message: "title must not be empty"}, nil
	}
	if model.Slug(req.Title) == "" {
		return &api.CreateItemBadRequest{Message: "title reduces to an empty slug; there is no filename to derive"}, nil
	}

	// capture is the one mutation that never reads the repository first —
	// CreateDraft opens with MkdirAll — and under a moved root it would
	// silently recreate the roadmap in the dead tree. the preflight makes a
	// degraded project refuse with its repository error, bytes untouched.
	if _, err := w.Load(); err != nil {
		return nil, err
	}

	temp, err := w.CreateDraft(req.Title, req.Body.Or(""))
	if err != nil {
		return nil, err
	}
	fin, err := w.FinalizeDraft(temp)
	if err != nil {
		return nil, err
	}
	switch fin.Outcome {
	case workspace.Finalized:
		board, err := freshBoard(w, params.Project)
		if err != nil {
			return nil, err
		}
		return &api.ItemLanded{Filename: fin.Filename, Board: *board}, nil
	case workspace.Collision:
		return &api.Conflict{
			Reason:   api.ConflictReasonSlugCollision,
			Message:  fmt.Sprintf("%s already exists; the draft is preserved", fin.DestPath),
			TempPath: api.NewOptString(fin.TempPath),
			DestPath: api.NewOptString(fin.DestPath),
		}, nil
	default:
		// prevalidation makes empty-title and empty-slug unreachable here
		return nil, fmt.Errorf("capture finalized unexpectedly (outcome %d); draft kept at %s", fin.Outcome, fin.TempPath)
	}
}

func (s *Server) SaveContent(_ context.Context, req *api.SaveContentReq, params api.SaveContentParams) (api.SaveContentRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	if err := w.SaveContent(params.Filename, []byte(req.Content), req.ExpectedHash, req.ExpectedOrderVersion); err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (s *Server) TransitionItem(_ context.Context, req *api.TransitionItemReq, params api.TransitionItemParams) (api.TransitionItemRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	var position *int
	if p, ok := req.Position.Get(); ok {
		position = &p
	}
	if err := w.Transition(params.Filename, model.State(req.State), req.ExpectedHash, req.ExpectedOrderVersion, position); err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (s *Server) ReorderLane(_ context.Context, req *api.ReorderLaneReq, params api.ReorderLaneParams) (api.ReorderLaneRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	if err := w.Reorder(model.State(params.Lane), req.Filenames, req.ExpectedVersion); err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (s *Server) DeleteItem(_ context.Context, req *api.DeleteItemReq, params api.DeleteItemParams) (api.DeleteItemRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return notFound, nil
	}
	if err := w.Delete(params.Filename, req.ExpectedHash, req.ExpectedOrderVersion); err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (s *Server) RetitleItem(_ context.Context, req *api.RetitleItemReq, params api.RetitleItemParams) (api.RetitleItemRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return (*api.RetitleItemNotFound)(notFound), nil
	}
	if req.Title == "" {
		return &api.RetitleItemBadRequest{Message: "title must not be empty"}, nil
	}
	newName, err := w.Retitle(params.Filename, req.Title, req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return &api.ItemLanded{Filename: newName, Board: *board}, nil
}

func (s *Server) RenameToSlug(_ context.Context, req *api.RenameToSlugReq, params api.RenameToSlugParams) (api.RenameToSlugRes, error) {
	w, notFound, err := s.resolve(params.Project)
	if err != nil {
		return nil, err
	}
	if notFound != nil {
		return (*api.RenameToSlugNotFound)(notFound), nil
	}
	newName, err := w.RenameToSlug(params.Filename, req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		var refusal *workspace.Refusal
		if errors.As(err, &refusal) {
			return &api.RenameToSlugBadRequest{Message: refusal.Msg}, nil
		}
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := freshBoard(w, params.Project)
	if err != nil {
		return nil, err
	}
	return &api.ItemLanded{Filename: newName, Board: *board}, nil
}
