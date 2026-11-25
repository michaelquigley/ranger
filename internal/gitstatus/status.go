// Package gitstatus reads worktree marks for a project-relative directory.
package gitstatus

import (
	"os/exec"
	"strings"
)

type Change struct {
	Staged    bool
	Unstaged  bool
	Untracked bool
}

type Status struct {
	Known bool
	// Paths are relative to the requested directory.
	Paths map[string]Change
}

// Read invokes the same status command for either convention.
func Read(root, scope string) Status {
	out, err := exec.Command("git", "-C", root, "status", "--porcelain", "-z", "--untracked-files=all", "--", scope).Output()
	if err != nil {
		return Status{}
	}
	status := Status{Known: true, Paths: map[string]Change{}}
	fields := strings.Split(string(out), "\x00")
	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if len(entry) < 4 {
			continue
		}
		xy, path := entry[:2], entry[3:]
		// the second path is the rename or copy source. keep only the destination.
		if strings.ContainsAny(xy, "RC") {
			i++
		}
		if rel, ok := strings.CutPrefix(path, scope+"/"); ok {
			change := Change{Untracked: xy == "??"}
			if !change.Untracked {
				change.Staged, change.Unstaged = xy[0] != ' ', xy[1] != ' '
			}
			status.Paths[rel] = change
		}
	}
	return status
}
