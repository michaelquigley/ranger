package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/workspace"
)

// healthyRoot creates a root whose roadmap directory exists, so a fresh
// load succeeds.
func healthyRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(workspace.RoadmapRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func constantSource(cfg *config.Config) Source {
	return func() (*config.Config, error) { return cfg, nil }
}

func TestResolve(t *testing.T) {
	root := healthyRoot(t)
	p := NewProjects(constantSource(&config.Config{
		Projects: []config.ProjectRef{{Root: root, Name: "ranger"}},
		Default:  "ranger",
	}))

	w, err := p.Resolve("ranger")
	if err != nil {
		t.Fatal(err)
	}
	if w.Root() != root {
		t.Errorf("root: %q, want %q", w.Root(), root)
	}

	_, err = p.Resolve("missing")
	var unknown *UnknownProjectError
	if !errors.As(err, &unknown) || unknown.Name != "missing" {
		t.Errorf("expected UnknownProjectError for %q, got %v", "missing", err)
	}
}

func TestResolveSourceError(t *testing.T) {
	fault := errors.New("config does not parse")
	p := NewProjects(func() (*config.Config, error) { return nil, fault })
	if _, err := p.Resolve("ranger"); !errors.Is(err, fault) {
		t.Errorf("expected the source's error, got %v", err)
	}
	if _, err := p.Index(); !errors.Is(err, fault) {
		t.Errorf("expected the source's error, got %v", err)
	}
}

func TestIndexAvailability(t *testing.T) {
	p := NewProjects(constantSource(&config.Config{
		Projects: []config.ProjectRef{
			{Root: healthyRoot(t), Name: "healthy"},
			{Root: filepath.Join(t.TempDir(), "moved"), Name: "broken"},
		},
		Default: "healthy",
	}))

	idx, err := p.Index()
	if err != nil {
		t.Fatal(err)
	}
	if idx.Default != "healthy" {
		t.Errorf("default: %q", idx.Default)
	}
	if len(idx.Projects) != 2 {
		t.Fatalf("projects: %d", len(idx.Projects))
	}
	if !idx.Projects[0].Available || idx.Projects[0].Error != "" {
		t.Errorf("healthy project flagged: %+v", idx.Projects[0])
	}
	if idx.Projects[1].Available || idx.Projects[1].Error == "" {
		t.Errorf("broken project not flagged: %+v", idx.Projects[1])
	}
}

// TestSourceConsultedFresh pins the no-cached-config promise: a source
// change is visible on the very next resolution.
func TestSourceConsultedFresh(t *testing.T) {
	current := &config.Config{
		Projects: []config.ProjectRef{{Root: healthyRoot(t), Name: "before"}},
		Default:  "before",
	}
	p := NewProjects(func() (*config.Config, error) { return current, nil })

	if _, err := p.Resolve("before"); err != nil {
		t.Fatal(err)
	}
	current = &config.Config{
		Projects: []config.ProjectRef{{Root: healthyRoot(t), Name: "after"}},
		Default:  "after",
	}
	if _, err := p.Resolve("before"); err == nil {
		t.Error("stale name still resolves; the config was cached")
	}
	if _, err := p.Resolve("after"); err != nil {
		t.Errorf("new name does not resolve: %v", err)
	}
	idx, err := p.Index()
	if err != nil {
		t.Fatal(err)
	}
	if idx.Default != "after" || len(idx.Projects) != 1 || idx.Projects[0].Name != "after" {
		t.Errorf("index reflects a stale config: %+v", idx)
	}
}
