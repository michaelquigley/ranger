package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV2CLIIgnoresSymlinkTargets(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	future := filepath.Join(root, "docs", "future")
	ranger := filepath.Join(future, ".ranger")

	writeCLIFile(t, filepath.Join(ranger, "version.yaml"), "version: 2\n")
	writeCLIFile(t, filepath.Join(future, "visible", "index.md"),
		"---\ntitle: Visible\nstate: inbox\ncreated: 2026-11-24\n---\n\nThe real prompt.\n")
	const foreignItem = "---\ntitle: Foreign title\nstate: inbox\ncreated: 2026-11-24\n---\n\nForeign prompt.\n"
	writeCLIFile(t, filepath.Join(outside, "foreign-directory", "index.md"), foreignItem)
	writeCLIFile(t, filepath.Join(outside, "foreign-index.md"), foreignItem)
	writeCLIFile(t, filepath.Join(outside, "foreign-view.yaml"), "- foreign-slug\n")
	linkCLIPath(t, filepath.Join(outside, "foreign-directory"), filepath.Join(future, "linked-dir"))
	linkCLIPath(t, filepath.Join(outside, "foreign-index.md"), filepath.Join(future, "linked-index", "index.md"))
	linkCLIPath(t, filepath.Join(outside, "foreign-view.yaml"), filepath.Join(ranger, "queue.yaml"))

	t.Chdir(filepath.Join(future, "visible"))
	for _, tt := range []struct {
		args []string
		text string
		fail bool
	}{
		{[]string{"roadmap"}, "Visible (visible)", false},
		{[]string{"queue"}, "queue.yaml: symlink refused", true},
		{[]string{"needs"}, "nothing needs you", false},
		{[]string{"show", "visible"}, "The real prompt.", false},
		{[]string{"show", "linked-dir"}, "no item named linked-dir", true},
		{[]string{"show", "linked-index"}, "no item named linked-index", true},
		{[]string{"flags"}, "no flags", false},
	} {
		out, err := runCLI(tt.args...)
		output := fmt.Sprintf("%s\n%v", out, err)
		if (err != nil) != tt.fail || !strings.Contains(output, tt.text) {
			t.Errorf("%v output = %q, want %q (error = %v)", tt.args, output, tt.text, tt.fail)
		}
		if strings.Contains(strings.ToLower(output), "foreign") {
			t.Errorf("%v exposed symlink target content or names: %q", tt.args, output)
		}
	}
}

func TestV2CLINamesSymlinkedProjectComponent(t *testing.T) {
	outside := t.TempDir()
	// a valid project behind the link must still be refused.
	writeCLIFile(t, filepath.Join(outside, "docs", "future", ".ranger", "version.yaml"), "version: 2\n")
	for _, rel := range []string{"docs", "docs/future", "docs/future/.ranger", "docs/future/.ranger/version.yaml"} {
		t.Run(rel, func(t *testing.T) {
			root := t.TempDir()
			component := filepath.Join(root, rel)
			linkCLIPath(t, filepath.Join(outside, rel), component)
			start := filepath.Join(root, "nested")
			if err := os.Mkdir(start, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(start)
			out, err := runCLI("roadmap")
			if out != "" || err == nil || !strings.Contains(err.Error(), component+": symlink refused") ||
				strings.Contains(err.Error(), outside) {
				t.Fatalf("roadmap output = %q, error = %v, want error naming %s", out, err, component)
			}
		})
	}
}

func writeCLIFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func linkCLIPath(t *testing.T, target, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}
