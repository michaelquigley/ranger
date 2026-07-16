package server

import (
	"context"
	"errors"
	"fmt"

	"git.hq.quigley.com/products/vane/internal/api"
	"git.hq.quigley.com/products/vane/internal/model"
	"git.hq.quigley.com/products/vane/internal/workspace"
)

func (s *Server) GetBoard(_ context.Context) (*api.Board, error) {
	return s.freshBoard()
}

func (s *Server) GetItem(_ context.Context, params api.GetItemParams) (api.GetItemRes, error) {
	snap, err := s.w.Load()
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

func (s *Server) CreateItem(_ context.Context, req *api.CreateItemReq) (api.CreateItemRes, error) {
	// prevalidated: no draft file exists for a title capture can never
	// finalize, so the form loses nothing and the tree stays clean.
	if req.Title == "" {
		return &api.ErrorResponse{Message: "title must not be empty"}, nil
	}
	if model.Slug(req.Title) == "" {
		return &api.ErrorResponse{Message: "title reduces to an empty slug; there is no filename to derive"}, nil
	}

	temp, err := s.w.CreateDraft(req.Title, req.Body.Or(""))
	if err != nil {
		return nil, err
	}
	fin, err := s.w.FinalizeDraft(temp)
	if err != nil {
		return nil, err
	}
	switch fin.Outcome {
	case workspace.Finalized:
		board, err := s.freshBoard()
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
	err := s.w.SaveContent(params.Filename, []byte(req.Content), req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	return s.boardRes()
}

func (s *Server) TransitionItem(_ context.Context, req *api.TransitionItemReq, params api.TransitionItemParams) (api.TransitionItemRes, error) {
	var position *int
	if p, ok := req.Position.Get(); ok {
		position = &p
	}
	err := s.w.Transition(params.Filename, model.State(req.State), req.ExpectedHash, req.ExpectedOrderVersion, position)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	return s.boardRes()
}

func (s *Server) ReorderLane(_ context.Context, req *api.ReorderLaneReq, params api.ReorderLaneParams) (api.ReorderLaneRes, error) {
	err := s.w.Reorder(model.State(params.Lane), req.Filenames, req.ExpectedVersion)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	return s.boardRes()
}

func (s *Server) DeleteItem(_ context.Context, req *api.DeleteItemReq, params api.DeleteItemParams) (api.DeleteItemRes, error) {
	err := s.w.Delete(params.Filename, req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	return s.boardRes()
}

func (s *Server) RetitleItem(_ context.Context, req *api.RetitleItemReq, params api.RetitleItemParams) (api.RetitleItemRes, error) {
	if req.Title == "" {
		return &api.ErrorResponse{Message: "title must not be empty"}, nil
	}
	newName, err := s.w.Retitle(params.Filename, req.Title, req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := s.freshBoard()
	if err != nil {
		return nil, err
	}
	return &api.ItemLanded{Filename: newName, Board: *board}, nil
}

func (s *Server) RenameToSlug(_ context.Context, req *api.RenameToSlugReq, params api.RenameToSlugParams) (api.RenameToSlugRes, error) {
	newName, err := s.w.RenameToSlug(params.Filename, req.ExpectedHash, req.ExpectedOrderVersion)
	if err != nil {
		var refusal *workspace.Refusal
		if errors.As(err, &refusal) {
			return &api.ErrorResponse{Message: refusal.Msg}, nil
		}
		if conflict, ok := asConflict(err); ok {
			return conflict, nil
		}
		return nil, err
	}
	board, err := s.freshBoard()
	if err != nil {
		return nil, err
	}
	return &api.ItemLanded{Filename: newName, Board: *board}, nil
}

// boardRes wraps the fresh board for the operations whose success response
// is the board itself.
func (s *Server) boardRes() (*api.Board, error) {
	return s.freshBoard()
}
