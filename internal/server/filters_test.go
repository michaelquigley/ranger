package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/document"
)

func saveFilters(t *testing.T, s *Server, filters []api.SavedFilter, expectedVersion string) api.SaveFiltersRes {
	t.Helper()
	res, err := s.SaveFilters(context.Background(), &api.SaveFiltersReq{
		Filters:         filters,
		ExpectedVersion: expectedVersion,
	}, api.SaveFiltersParams{Project: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestSaveFiltersLifecycle(t *testing.T) {
	s, w := fixture(t, false)

	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if board.FiltersVersion.Or("") != document.VersionAbsent {
		t.Fatalf("filtersVersion = %+v, want the absent sentinel", board.FiltersVersion)
	}
	if len(board.SavedFilters) != 0 {
		t.Fatalf("savedFilters = %+v, want empty", board.SavedFilters)
	}

	first := api.SavedFilter{Name: "active work", Tags: []string{"feature"}, NotTags: []string{"spike"}}
	board = mustBoard(t, saveFilters(t, s, []api.SavedFilter{first}, document.VersionAbsent), nil)
	if len(board.SavedFilters) != 1 || board.SavedFilters[0].Name != "active work" {
		t.Fatalf("savedFilters = %+v", board.SavedFilters)
	}
	onDisk, err := os.ReadFile(filepath.Join(w.Root(), "docs", "future", "roadmap", ".ranger", "filters.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if board.FiltersVersion.Or("") != document.Hash(onDisk) {
		t.Error("filtersVersion must be the on-disk hash")
	}

	second := api.SavedFilter{Name: "quiet", NotMilestone: api.NewOptString("v0.1.x")}
	board = mustBoard(t, saveFilters(t, s, []api.SavedFilter{first, second}, board.FiltersVersion.Or("")), nil)
	if len(board.SavedFilters) != 2 || board.SavedFilters[1].NotMilestone.Or("") != "v0.1.x" {
		t.Fatalf("savedFilters = %+v", board.SavedFilters)
	}

	third := api.SavedFilter{Name: "untriaged", NoTags: api.NewOptBool(true), NoMilestone: api.NewOptBool(true)}
	board = mustBoard(t, saveFilters(t, s, []api.SavedFilter{first, second, third}, board.FiltersVersion.Or("")), nil)
	if len(board.SavedFilters) != 3 || !board.SavedFilters[2].NoTags.Or(false) || !board.SavedFilters[2].NoMilestone.Or(false) {
		t.Fatalf("savedFilters = %+v", board.SavedFilters)
	}
}

func TestSaveFiltersStaleVersionConflicts(t *testing.T) {
	s, _ := fixture(t, false)
	one := []api.SavedFilter{{Name: "a"}}
	mustBoard(t, saveFilters(t, s, one, document.VersionAbsent), nil)

	res := saveFilters(t, s, one, document.VersionAbsent)
	conflict, ok := res.(*api.Conflict)
	if !ok {
		t.Fatalf("res = %#v, want a conflict", res)
	}
	if conflict.Reason != api.ConflictReasonFiltersConflict {
		t.Errorf("reason = %s, want filters_conflict", conflict.Reason)
	}
}

func TestSaveFiltersValidation(t *testing.T) {
	s, _ := fixture(t, false)
	cases := map[string][]api.SavedFilter{
		"empty name":     {{Name: ""}},
		"duplicate name": {{Name: "a"}, {Name: "a"}},
	}
	for label, filters := range cases {
		res := saveFilters(t, s, filters, document.VersionAbsent)
		if _, ok := res.(*api.SaveFiltersBadRequest); !ok {
			t.Errorf("%s: res = %#v, want a validation refusal", label, res)
		}
	}
}

func TestUnreadableFiltersDegrades(t *testing.T) {
	s, w := fixture(t, false)
	path := filepath.Join(w.Root(), "docs", "future", "roadmap", ".ranger", "filters.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("filters: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if board.FiltersVersion.Set {
		t.Errorf("filtersVersion = %+v, want absent — degraded, never a board failure", board.FiltersVersion)
	}
	if board.SavedFilters != nil {
		t.Errorf("savedFilters = %+v, want absent", board.SavedFilters)
	}
}
