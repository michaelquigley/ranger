package server

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/document"
	"github.com/michaelquigley/ranger/internal/workspace"
)

const (
	retryItem = `---
title: retry semantics
state: researching
created: 2026-07-01
tags: [retry]
source: github:openziti/zrok#412
log:
  - stamp: 2026-07-02
    note: spec drawn
---

retry body.
`
	captureItem = `---
title: board capture
state: inbox
created: 2026-07-03
---
`
	orderFixture = `researching:
  - retry-semantics.md
`
)

func fixture(t *testing.T, withOrder bool) (*Server, *workspace.Workspace) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"retry-semantics.md": retryItem,
		"board-capture.md":   captureItem,
	}
	if withOrder {
		files["order.yaml"] = orderFixture
	}
	for name, content := range files {
		path := filepath.Join(root, "docs", "future", "roadmap", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w := workspace.New(root)
	cfg := &config.Config{
		Projects: []config.ProjectRef{{Root: root, Name: "test"}},
		Default:  "test",
		Port:     config.DefaultPort,
	}
	return New(NewProjects(constantSource(cfg))), w
}

func hashes(t *testing.T, w *workspace.Workspace, filename string) (string, string) {
	t.Helper()
	snap, err := w.Load()
	if err != nil {
		t.Fatal(err)
	}
	it, ok := snap.Item(filename)
	if !ok {
		t.Fatalf("fixture item %s missing", filename)
	}
	return it.Hash, snap.OrderVersion
}

func laneOf(t *testing.T, board *api.Board, state api.State) api.Lane {
	t.Helper()
	for _, lane := range board.Lanes {
		if lane.State == state {
			return lane
		}
	}
	t.Fatalf("no %s lane", state)
	return api.Lane{}
}

func mustBoard(t *testing.T, res any, err error) *api.Board {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	board, ok := res.(*api.Board)
	if !ok {
		t.Fatalf("res = %#v", res)
	}
	return board
}

func TestGetBoard(t *testing.T) {
	s, _ := fixture(t, true)
	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if board.Project != "test" {
		t.Errorf("project = %s, want the configured name", board.Project)
	}
	if board.OrderVersion != document.Hash([]byte(orderFixture)) {
		t.Errorf("orderVersion = %s", board.OrderVersion)
	}
	if len(board.Lanes) != 5 {
		t.Fatalf("lanes = %d, want 5", len(board.Lanes))
	}
	researching := laneOf(t, board, api.StateResearching)
	if researching.RankedCount != 1 || len(researching.Cards) != 1 {
		t.Fatalf("researching lane = %+v", researching)
	}
	card := researching.Cards[0]
	if card.Hash != document.Hash([]byte(retryItem)) {
		t.Error("card must carry its content hash")
	}
	if card.Title != "retry semantics" || card.Tags[0] != "retry" || card.Source.Or("") != "github:openziti/zrok#412" {
		t.Errorf("card = %+v", card)
	}
	if len(card.Log) != 1 || card.Log[0].Note != "spec drawn" {
		t.Errorf("log = %+v", card.Log)
	}
}

func TestGetBoardAbsentOrderVersion(t *testing.T) {
	s, _ := fixture(t, false)
	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if board.OrderVersion != document.VersionAbsent {
		t.Errorf("orderVersion = %s, want the absent sentinel", board.OrderVersion)
	}
}

func TestUnknownProjectIs404(t *testing.T) {
	s, _ := fixture(t, true)
	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if _, is404 := res.(*api.ErrorResponse); !is404 {
		t.Fatalf("board res = %#v", res)
	}
	createRes, err := s.CreateItem(context.Background(), &api.CreateItemReq{Title: "anything"}, api.CreateItemParams{Project: "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if _, is404 := createRes.(*api.CreateItemNotFound); !is404 {
		t.Fatalf("create res = %#v", createRes)
	}
}

func TestSearchItems(t *testing.T) {
	s, _ := fixture(t, true)
	tests := []struct {
		q    string
		want []string
	}{
		{"RETRY", []string{"retry-semantics.md"}},
		{"retry body", []string{"retry-semantics.md"}},
		{"board", []string{"board-capture.md"}},
		{"nothing-matches-this", []string{}},
	}
	for _, tt := range tests {
		res, err := s.SearchItems(context.Background(), api.SearchItemsParams{Project: "test", Q: tt.q})
		if err != nil {
			t.Fatal(err)
		}
		found, isOK := res.(*api.SearchItemsOK)
		if !isOK {
			t.Fatalf("res = %#v", res)
		}
		if len(found.Filenames) != len(tt.want) {
			t.Errorf("search %q = %v, want %v", tt.q, found.Filenames, tt.want)
			continue
		}
		for i, f := range tt.want {
			if found.Filenames[i] != f {
				t.Errorf("search %q = %v, want %v", tt.q, found.Filenames, tt.want)
			}
		}
	}
}

func TestGetItem(t *testing.T) {
	s, _ := fixture(t, true)
	res, err := s.GetItem(context.Background(), api.GetItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	ok, isOK := res.(*api.GetItemOK)
	if !isOK {
		t.Fatalf("res = %T", res)
	}
	if ok.Content != retryItem || ok.Hash != document.Hash([]byte(retryItem)) {
		t.Error("content and hash must be the raw truth")
	}

	missing, err := s.GetItem(context.Background(), api.GetItemParams{Project: "test", Filename: "gone.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, is404 := missing.(*api.ErrorResponse); !is404 {
		t.Fatalf("missing item res = %T", missing)
	}
}

func TestStaleItemHashConflicts(t *testing.T) {
	s, w := fixture(t, true)
	_, orderVersion := hashes(t, w, "retry-semantics.md")
	res, err := s.TransitionItem(context.Background(),
		&api.TransitionItemReq{State: api.StateBuilding, ExpectedHash: document.Hash([]byte("stale view")), ExpectedOrderVersion: orderVersion},
		api.TransitionItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	conflict, isConflict := res.(*api.Conflict)
	if !isConflict || conflict.Reason != api.ConflictReasonItemConflict {
		t.Fatalf("res = %#v", res)
	}
}

func TestStaleOrderVersionConflicts(t *testing.T) {
	s, w := fixture(t, true)
	hash, _ := hashes(t, w, "retry-semantics.md")
	res, err := s.TransitionItem(context.Background(),
		&api.TransitionItemReq{State: api.StateBuilding, ExpectedHash: hash, ExpectedOrderVersion: document.Hash([]byte("stale order"))},
		api.TransitionItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	conflict, isConflict := res.(*api.Conflict)
	if !isConflict || conflict.Reason != api.ConflictReasonOrderConflict {
		t.Fatalf("res = %#v", res)
	}
}

func TestExpectedAbsentVersusRacingCreation(t *testing.T) {
	// an expected-absent write against a present order.yaml is the racing
	// creator's loss: refuse and reload.
	s, _ := fixture(t, true)
	res, err := s.ReorderLane(context.Background(),
		&api.ReorderLaneReq{Filenames: []string{"retry-semantics.md"}, ExpectedVersion: document.VersionAbsent},
		api.ReorderLaneParams{Project: "test", Lane: api.StateResearching})
	if err != nil {
		t.Fatal(err)
	}
	conflict, isConflict := res.(*api.Conflict)
	if !isConflict || conflict.Reason != api.ConflictReasonOrderConflict {
		t.Fatalf("res = %#v", res)
	}

	// against a genuinely absent order.yaml the same expectation creates it
	sAbsent, _ := fixture(t, false)
	absentRes, err := sAbsent.ReorderLane(context.Background(),
		&api.ReorderLaneReq{Filenames: []string{"retry-semantics.md"}, ExpectedVersion: document.VersionAbsent},
		api.ReorderLaneParams{Project: "test", Lane: api.StateResearching})
	board := mustBoard(t, absentRes, err)
	if laneOf(t, board, api.StateResearching).RankedCount != 1 {
		t.Error("first-ever ranking must land")
	}
	if board.OrderVersion == document.VersionAbsent {
		t.Error("fresh board must carry the created order.yaml's version")
	}
}

func TestTransitionReturnsFreshBoard(t *testing.T) {
	s, w := fixture(t, true)
	hash, orderVersion := hashes(t, w, "board-capture.md")
	pos := 0
	req := &api.TransitionItemReq{State: api.StateResearching, ExpectedHash: hash, ExpectedOrderVersion: orderVersion}
	req.Position = api.NewOptInt(pos)
	res, err := s.TransitionItem(context.Background(), req,
		api.TransitionItemParams{Project: "test", Filename: "board-capture.md"})
	board := mustBoard(t, res, err)
	researching := laneOf(t, board, api.StateResearching)
	if researching.RankedCount != 2 || researching.Cards[0].Filename != "board-capture.md" {
		t.Errorf("transition-and-place did not paint through: %+v", researching)
	}
}

func TestPartialTwoFileFailureIsReportedVerbatim(t *testing.T) {
	s, w := fixture(t, true)
	hash, orderVersion := hashes(t, w, "retry-semantics.md")
	orderPath := filepath.Join(w.Root(), "docs", "future", "roadmap", "order.yaml")
	if err := os.Chmod(orderPath, 0o444); err != nil {
		t.Fatal(err)
	}
	_, err := s.TransitionItem(context.Background(),
		&api.TransitionItemReq{State: api.StateBuilding, ExpectedHash: hash, ExpectedOrderVersion: orderVersion},
		api.TransitionItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err == nil {
		t.Fatal("write against a read-only order.yaml must fail")
	}
	wire := s.NewError(context.Background(), err)
	if !strings.Contains(wire.Response.Message, "but order.yaml was not updated") {
		t.Errorf("partial failure must be reported plainly, got %q", wire.Response.Message)
	}
}

func TestCreateItem(t *testing.T) {
	t.Run("captured into inbox with a fresh board", func(t *testing.T) {
		s, _ := fixture(t, true)
		req := &api.CreateItemReq{Title: "new idea"}
		req.Body = api.NewOptString("the body.\n")
		res, err := s.CreateItem(context.Background(), req, api.CreateItemParams{Project: "test"})
		if err != nil {
			t.Fatal(err)
		}
		landed, isLanded := res.(*api.ItemLanded)
		if !isLanded || landed.Filename != "new-idea.md" {
			t.Fatalf("res = %#v", res)
		}
		inbox := laneOf(t, &landed.Board, api.StateInbox)
		found := false
		for _, c := range inbox.Cards {
			if c.Filename == "new-idea.md" {
				found = true
			}
		}
		if !found {
			t.Error("fresh board must show the captured card")
		}
	})

	t.Run("prevalidation writes no draft", func(t *testing.T) {
		s, w := fixture(t, true)
		for _, title := range []string{"", "((()))"} {
			res, err := s.CreateItem(context.Background(), &api.CreateItemReq{Title: title}, api.CreateItemParams{Project: "test"})
			if err != nil {
				t.Fatal(err)
			}
			if _, is400 := res.(*api.CreateItemBadRequest); !is400 {
				t.Fatalf("res = %#v", res)
			}
		}
		entries, _ := os.ReadDir(filepath.Join(w.Root(), "docs", "future", "roadmap"))
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".capture-") {
				t.Errorf("prevalidated capture left a draft: %s", e.Name())
			}
		}
	})

	t.Run("collision preserves the draft and reports both paths", func(t *testing.T) {
		s, _ := fixture(t, true)
		res, err := s.CreateItem(context.Background(), &api.CreateItemReq{Title: "board capture"}, api.CreateItemParams{Project: "test"})
		if err != nil {
			t.Fatal(err)
		}
		conflict, isConflict := res.(*api.Conflict)
		if !isConflict || conflict.Reason != api.ConflictReasonSlugCollision {
			t.Fatalf("res = %#v", res)
		}
		temp := conflict.TempPath.Or("")
		if temp == "" || conflict.DestPath.Or("") == "" {
			t.Fatalf("collision must carry structured recovery paths: %#v", conflict)
		}
		if _, err := os.Stat(temp); err != nil {
			t.Error("the preserved draft must exist at the reported path")
		}
	})
}

func TestRetitleCollisionCarriesPaths(t *testing.T) {
	s, w := fixture(t, true)
	hash, orderVersion := hashes(t, w, "retry-semantics.md")
	res, err := s.RetitleItem(context.Background(),
		&api.RetitleItemReq{Title: "board capture", ExpectedHash: hash, ExpectedOrderVersion: orderVersion},
		api.RetitleItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	conflict, isConflict := res.(*api.Conflict)
	if !isConflict || conflict.Reason != api.ConflictReasonSlugCollision {
		t.Fatalf("res = %#v", res)
	}
	if conflict.SourcePath.Or("") == "" || conflict.DestPath.Or("") == "" {
		t.Errorf("retitle collision must carry source and destination: %#v", conflict)
	}
}

func TestRenameToSlugRefusalIsValidation(t *testing.T) {
	s, w := fixture(t, true)
	hash, orderVersion := hashes(t, w, "retry-semantics.md")
	// retry-semantics.md already matches its slug; retitle it to an
	// empty-slug title first so the rename has nothing to repair toward.
	res, err := s.RetitleItem(context.Background(),
		&api.RetitleItemReq{Title: "((()))", ExpectedHash: hash, ExpectedOrderVersion: orderVersion},
		api.RetitleItemParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, isLanded := res.(*api.ItemLanded); !isLanded {
		t.Fatalf("res = %#v", res)
	}
	hash, orderVersion = hashes(t, w, "retry-semantics.md")
	renameRes, err := s.RenameToSlug(context.Background(),
		&api.RenameToSlugReq{ExpectedHash: hash, ExpectedOrderVersion: orderVersion},
		api.RenameToSlugParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, is400 := renameRes.(*api.RenameToSlugBadRequest); !is400 {
		t.Fatalf("refusal must be a validation error, got %#v", renameRes)
	}
}

func TestSaveContentConflictOnStaleHash(t *testing.T) {
	s, _ := fixture(t, true)
	res, err := s.SaveContent(context.Background(),
		&api.SaveContentReq{Content: "anything", ExpectedHash: document.Hash([]byte("stale")), ExpectedOrderVersion: document.VersionAbsent},
		api.SaveContentParams{Project: "test", Filename: "retry-semantics.md"})
	if err != nil {
		t.Fatal(err)
	}
	conflict, isConflict := res.(*api.Conflict)
	if !isConflict || conflict.Reason != api.ConflictReasonItemConflict {
		t.Fatalf("res = %#v", res)
	}
}

// TestHTTPRoundTrip drives one read and one mutation through the real wire
// — generated client through generated router to handler — proving the
// project-scoped paths compose; a method-level test cannot catch a stale
// path.
func TestHTTPRoundTrip(t *testing.T) {
	s, w := fixture(t, true)
	h, err := api.NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()
	c, err := api.NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	readRes, err := c.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, readRes, err)
	if board.Project != "test" || board.OrderVersion == "" {
		t.Errorf("wire board = project %q, orderVersion %q", board.Project, board.OrderVersion)
	}

	hash, orderVersion := hashes(t, w, "board-capture.md")
	mutRes, err := c.TransitionItem(context.Background(),
		&api.TransitionItemReq{State: api.StateResearching, ExpectedHash: hash, ExpectedOrderVersion: orderVersion},
		api.TransitionItemParams{Project: "test", Filename: "board-capture.md"})
	mutated := mustBoard(t, mutRes, err)
	found := false
	for _, card := range laneOf(t, mutated, api.StateResearching).Cards {
		if card.Filename == "board-capture.md" {
			found = true
		}
	}
	if !found {
		t.Error("mutation must land through the wire and paint the fresh board")
	}

	res, err := c.GetBoard(context.Background(), api.GetBoardParams{Project: "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if _, is404 := res.(*api.ErrorResponse); !is404 {
		t.Fatalf("unknown project over the wire = %#v", res)
	}
}
