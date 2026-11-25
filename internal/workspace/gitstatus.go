package workspace

import "github.com/michaelquigley/ranger/internal/gitstatus"

// GitStatus is one fresh read of the roadmap directory's uncommitted
// state. read is the operative word: the 2026-07-21 design change opened
// read-only status inspection through the git CLI and nothing more — the
// tool still never writes git state; commit, push, and history remain the
// operator's judgment gate.
type GitStatus struct {
	// Known reports whether git answered: false without a git binary,
	// outside a repository, or on any git failure. unknown is the absence
	// of information, never an error and never cleanliness.
	Known bool
	// Dirty is true when anything under the roadmap directory is
	// uncommitted — items, order.yaml, and assets alike.
	Dirty bool
	// Files holds the dirty paths relative to the roadmap directory, so
	// flat item filenames key directly.
	Files map[string]bool
}

// GitStatus shells out to `git status --porcelain` scoped to the roadmap
// directory — read-only, fresh per call, and modified, staged, and
// untracked all count: every one is work the operator hasn't committed.
func (w *Workspace) GitStatus() GitStatus {
	git := gitstatus.Read(w.root, RoadmapRel)
	if !git.Known {
		return GitStatus{}
	}
	status := GitStatus{Known: true, Dirty: len(git.Paths) != 0, Files: make(map[string]bool, len(git.Paths))}
	for path := range git.Paths {
		status.Files[path] = true
	}
	return status
}
