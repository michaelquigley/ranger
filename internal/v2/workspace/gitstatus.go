package workspace

import (
	"strings"

	"github.com/michaelquigley/ranger/internal/gitstatus"
	"github.com/michaelquigley/ranger/internal/v2/model"
)

type GitMark struct {
	Staged   bool
	Unstaged bool
	New      bool
}

type ProjectGitStatus struct {
	Known bool
	Dirty bool
	Items map[string]GitMark
	Views map[model.ViewName]GitMark
}

// GitStatus combines path changes for each item and reports the two view files.
func GitStatus(root string) ProjectGitStatus {
	git := gitstatus.Read(root, FutureRel)
	status := ProjectGitStatus{
		Known: git.Known, Dirty: len(git.Paths) != 0,
		Items: map[string]GitMark{}, Views: map[model.ViewName]GitMark{},
	}
	for _, view := range []model.ViewName{model.Roadmap, model.Queue} {
		change := git.Paths[".ranger/"+string(view)+".yaml"]
		status.Views[view] = GitMark{Staged: change.Staged, Unstaged: change.Unstaged, New: change.Untracked}
	}
	for path, change := range git.Paths {
		slug, file, ok := strings.Cut(path, "/")
		if !ok || strings.HasPrefix(slug, ".") {
			continue
		}
		mark := status.Items[slug]
		newIndex := file == "index.md" && change.Untracked
		mark.New = mark.New || newIndex
		mark.Staged = mark.Staged || change.Staged
		mark.Unstaged = mark.Unstaged || change.Unstaged || (change.Untracked && !newIndex)
		status.Items[slug] = mark
	}
	return status
}
