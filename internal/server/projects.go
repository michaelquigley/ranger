package server

import (
	"fmt"

	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/workspace"
)

// Source yields the current config, consulted fresh on every resolution.
// the daemon's source re-reads the file per call so an edit takes effect on
// the next request; serve's returns its synthesized config constantly.
type Source func() (*config.Config, error)

// Projects resolves project names to workspaces over a config source. no
// cached config, no cached health, no background probing — availability is
// judged by a fresh Load wherever it's asked, same as every other read in
// the system.
type Projects struct {
	source Source
}

// NewProjects returns a project set over source.
func NewProjects(source Source) *Projects {
	return &Projects{source: source}
}

// UnknownProjectError reports a name no configured project carries; the
// handlers map it to the wire's 404.
type UnknownProjectError struct {
	Name string
}

func (e *UnknownProjectError) Error() string {
	return fmt.Sprintf("unknown project %q", e.Name)
}

// Resolve maps a project name to its workspace. a miss is an
// UnknownProjectError; whether the workspace is healthy is the caller's own
// fresh Load to judge.
func (p *Projects) Resolve(name string) (*workspace.Workspace, error) {
	cfg, err := p.source()
	if err != nil {
		return nil, err
	}
	for _, ref := range cfg.Projects {
		if ref.Name == name {
			return workspace.New(ref.Root), nil
		}
	}
	return nil, &UnknownProjectError{Name: name}
}

// ProjectStatus is one project's index entry: its name and the verdict of a
// fresh load at call time.
type ProjectStatus struct {
	Name      string
	Available bool
	Error     string
}

// ProjectIndex is the enumerated project set plus the default's name.
type ProjectIndex struct {
	Projects []ProjectStatus
	Default  string
}

// Index enumerates the configured projects, judging each root's
// availability by a fresh load.
func (p *Projects) Index() (*ProjectIndex, error) {
	cfg, err := p.source()
	if err != nil {
		return nil, err
	}
	idx := &ProjectIndex{Default: cfg.Default}
	for _, ref := range cfg.Projects {
		status := ProjectStatus{Name: ref.Name, Available: true}
		if _, err := workspace.New(ref.Root).Load(); err != nil {
			status.Available = false
			status.Error = err.Error()
		}
		idx.Projects = append(idx.Projects, status)
	}
	return idx, nil
}
