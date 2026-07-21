package server

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/config"
)

// treeOf snapshots every path and file's bytes under root, so a refused
// gesture can be proven to have left the filesystem untouched.
func treeOf(t *testing.T, root string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			tree[rel] = "(dir)"
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tree[rel] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

// TestDegradationAndHeal is the executable form of the error-by-tier
// census entry: on one server instance, a broken root degrades — flagged
// in the index, its board and capture refusing with the repository error,
// bytes untouched — and heals on the next request after the root heals on
// disk, with no rebuild. it is the test that catches an accidental startup
// gate or cached health.
func TestDegradationAndHeal(t *testing.T) {
	ctx := context.Background()

	healthyRt := t.TempDir()
	if err := os.MkdirAll(filepath.Join(healthyRt, "docs", "future", "roadmap"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(healthyRt, "docs", "future", "roadmap", "retry-semantics.md"), []byte(retryItem), 0o644); err != nil {
		t.Fatal(err)
	}
	brokenRt := t.TempDir() // no roadmap directory: a moved root

	cfg := &config.Config{
		Projects: []config.ProjectRef{
			{Root: healthyRt, Name: "healthy"},
			{Root: brokenRt, Name: "broken"},
		},
		Default: "healthy",
		Port:    config.DefaultPort,
	}
	s := New(NewProjects(constantSource(cfg)))

	// the index reports one available and one flagged with its error
	idx, err := s.GetProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Projects) != 2 || !idx.Projects[0].Available || idx.Projects[1].Available {
		t.Fatalf("index = %+v", idx)
	}
	if diag := idx.Projects[1].Error.Or(""); !strings.Contains(diag, "roadmap directory") {
		t.Errorf("broken project must carry its diagnostic, got %q", diag)
	}

	// the broken board returns the repository error; the healthy board works
	if _, err := s.GetBoard(ctx, api.GetBoardParams{Project: "broken"}); err == nil || !strings.Contains(err.Error(), "roadmap directory") {
		t.Fatalf("broken board err = %v", err)
	}
	healthyRes, err := s.GetBoard(ctx, api.GetBoardParams{Project: "healthy"})
	board := mustBoard(t, healthyRes, err)
	if board.Project != "healthy" {
		t.Errorf("healthy board = %+v", board)
	}

	// capture against the broken project refuses with the repository error
	// and leaves the filesystem byte-for-byte unchanged: no recreated
	// roadmap directory, no landed draft.
	before := treeOf(t, brokenRt)
	if _, err := s.CreateItem(ctx, &api.CreateItemReq{Title: "stray thought"}, api.CreateItemParams{Project: "broken"}); err == nil || !strings.Contains(err.Error(), "roadmap directory") {
		t.Fatalf("broken capture err = %v", err)
	}
	if after := treeOf(t, brokenRt); !reflect.DeepEqual(before, after) {
		t.Errorf("refused capture touched the tree: before %v, after %v", before, after)
	}

	// heal the root on disk; index and board recover on the next request,
	// same server instance, no rebuild.
	if err := os.MkdirAll(filepath.Join(brokenRt, "docs", "future", "roadmap"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brokenRt, "docs", "future", "roadmap", "board-capture.md"), []byte(captureItem), 0o644); err != nil {
		t.Fatal(err)
	}
	idx, err = s.GetProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !idx.Projects[0].Available || !idx.Projects[1].Available {
		t.Fatalf("healed index = %+v", idx)
	}
	healedRes, err := s.GetBoard(ctx, api.GetBoardParams{Project: "broken"})
	healed := mustBoard(t, healedRes, err)
	if len(laneOf(t, healed, api.StateInbox).Cards) != 1 {
		t.Errorf("healed board = %+v", healed)
	}
}
