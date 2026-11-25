package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV2CLIReadsLeaveRehearsalsUnchanged(t *testing.T) {
	for _, project := range []struct {
		name, item, queue, detail, needs string
	}{
		{"baab", "agent-mcp-score-creation-from-a-local-file", "stage 1 of 1 (file-backed-score-calls): built [terminus: clean rehearsal-run]", "file: stages/file-backed-score-calls.md", "awaiting acceptance"},
		{"screeningroom", "fullscreen-in-the-client", "2 open of 4", "link in (body): look-and-feel-pass", "nothing needs you"},
	} {
		t.Run(project.name, func(t *testing.T) {
			root := t.TempDir()
			fixture := filepath.Join("..", "..", "internal", "v2", "workspace", "testdata", project.name)
			if err := os.CopyFS(root, os.DirFS(fixture)); err != nil {
				t.Fatal(err)
			}
			t.Chdir(filepath.Join(root, "docs", "future", project.item))
			before := cliTreeHash(t, root)
			for _, command := range []struct {
				args []string
				want string
			}{
				{[]string{"roadmap"}, "building:"},
				{[]string{"queue"}, project.queue},
				{[]string{"needs"}, project.needs},
				{[]string{"show", project.item}, project.detail},
				{[]string{"flags"}, "no flags"},
			} {
				out, err := runCLI(command.args...)
				if err != nil || !strings.Contains(out, command.want) {
					t.Errorf("%v = %q, %v, want %q", command.args, out, err, command.want)
				}
				if after := cliTreeHash(t, root); after != before {
					t.Errorf("%v modified the project tree", command.args)
				}
			}
			for _, command := range []struct {
				args []string
				want string
			}{
				{[]string{"list"}, "ranger roadmap or ranger queue"},
				{[]string{"state", project.item, "pending"}, "state is not built yet"},
				{[]string{"new work"}, "capture is not built yet"},
			} {
				if _, err := runCLI(command.args...); err == nil || !strings.Contains(err.Error(), command.want) {
					t.Errorf("%v error = %v, want %q", command.args, err, command.want)
				}
			}
			if after := cliTreeHash(t, root); after != before {
				t.Error("refused commands modified the project tree")
			}
		})
	}
}

func TestV1CaptureStillWorksWithQuotedCommandWord(t *testing.T) {
	root := t.TempDir()
	roadmap := filepath.Join(root, "docs", "future", "roadmap")
	if err := os.MkdirAll(roadmap, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	t.Setenv("RANGER_EDITOR", "true")
	if _, err := runCLI("queue", "follow-up"); err == nil {
		t.Fatal("an unquoted command word should reach the command")
	}
	for _, args := range [][]string{{"roadmap"}, {"queue"}, {"needs"}, {"show", "anything"}, {"flags"}} {
		if _, err := runCLI(args...); err == nil || !strings.Contains(err.Error(), "uses v1") {
			t.Errorf("%v error = %v", args, err)
		}
	}
	if _, err := runCLI("queue follow-up"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI("list"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI("state", "queue-follow-up", "researching"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(roadmap, "queue-follow-up.md"))
	if err != nil || !bytes.Contains(raw, []byte("title: queue follow-up\nstate: researching\n")) {
		t.Errorf("captured item = %q, error = %v", raw, err)
	}
}

func TestV2CLIReportsBrokenViewOnlyWhenRequested(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join("..", "..", "internal", "v2", "workspace", "testdata", "baab")
	if err := os.CopyFS(root, os.DirFS(fixture)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	err := os.WriteFile("docs/future/.ranger/queue.yaml", []byte("not a list"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI("queue"); err == nil || !strings.Contains(err.Error(), "queue.yaml") {
		t.Errorf("queue error = %v", err)
	}
	if out, err := runCLI("roadmap"); err != nil || !strings.Contains(out, "building:") {
		t.Errorf("roadmap = %q, %v", out, err)
	}
}

func runCLI(args ...string) (string, error) {
	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	err := cmd.Execute()
	return out.String(), err
}

func cliTreeHash(t *testing.T, root string) [sha256.Size]byte {
	t.Helper()
	hash := sha256.New()
	tree := os.DirFS(root)
	err := fs.WalkDir(tree, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%q %v\n", path, entry.Type())
		if entry.IsDir() {
			return nil
		}
		raw, err := fs.ReadFile(tree, path)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%x\n", sha256.Sum256(raw))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return [sha256.Size]byte(hash.Sum(nil))
}
