package server

import (
	"fmt"

	"github.com/michaelquigley/ranger/internal/config"
	v2 "github.com/michaelquigley/ranger/internal/v2/workspace"
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

// Project is a configured root with its freshly detected convention.
type Project struct {
	root       string
	Convention v2.Convention
}

func (p Project) Root() string { return p.root }

// LoadForAvailability loads with the detected convention. the v1 load
// can migrate a legacy order.yaml.
func (p Project) LoadForAvailability() error {
	switch p.Convention {
	case v2.V1:
		_, err := workspace.New(p.root).Load()
		return err
	case v2.V2:
		_, err := v2.Load(p.root)
		return err
	default:
		return fmt.Errorf("no detected convention for %s", p.root)
	}
}

func detectProject(root string) (Project, error) {
	convention, err := v2.Detect(root)
	return Project{root: root, Convention: convention}, err
}

// Resolve maps a configured name to its root and detects its convention.
func (p *Projects) Resolve(name string) (Project, error) {
	root, err := p.rootFor(name)
	if err != nil {
		return Project{}, err
	}
	return detectProject(root)
}

func (p *Projects) rootFor(name string) (string, error) {
	cfg, err := p.source()
	if err != nil {
		return "", err
	}
	for _, ref := range cfg.Projects {
		if ref.Name == name {
			return ref.Root, nil
		}
	}
	return "", &UnknownProjectError{Name: name}
}

// ProjectStatus is one project's index entry, judged by a fresh load.
type ProjectStatus struct {
	Name       string
	Convention v2.Convention
	Available  bool
	Error      string
	// Dirty carries the git verdict only when DirtyKnown.
	Dirty      bool
	DirtyKnown bool
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
		project, err := detectProject(ref.Root)
		status := ProjectStatus{Name: ref.Name, Convention: project.Convention}
		if err == nil {
			err = project.LoadForAvailability()
			switch project.Convention {
			case v2.V1:
				git := workspace.New(project.root).GitStatus()
				status.Dirty, status.DirtyKnown = git.Dirty, git.Known
			case v2.V2:
				git := v2.GitStatus(project.root)
				status.Dirty, status.DirtyKnown = git.Dirty, git.Known
			}
		}
		status.Available = err == nil
		if err != nil {
			status.Error = err.Error()
		}
		idx.Projects = append(idx.Projects, status)
	}
	return idx, nil
}
