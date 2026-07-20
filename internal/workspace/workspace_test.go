package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.hq.quigley.com/products/ranger/internal/document"
	"git.hq.quigley.com/products/ranger/internal/model"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// treeState snapshots every file under root for whole-tree diff assertions.
func treeState(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// assertTreeDiff asserts the gesture changed exactly the files in changed
// (to the given content), removed exactly the files in deleted, and left
// every other byte in the tree untouched.
func assertTreeDiff(t *testing.T, before, after, changed map[string]string, deleted ...string) {
	t.Helper()
	gone := map[string]bool{}
	for _, d := range deleted {
		gone[d] = true
	}
	for path, b := range before {
		a, exists := after[path]
		switch {
		case gone[path]:
			if exists {
				t.Errorf("%s should have been removed", path)
			}
		case !exists:
			t.Errorf("%s disappeared unexpectedly", path)
		default:
			if w, isChanged := changed[path]; isChanged {
				if a != w {
					t.Errorf("%s content:\ngot:\n%s\nwant:\n%s", path, a, w)
				}
			} else if a != b {
				t.Errorf("%s changed unexpectedly:\nbefore:\n%s\nafter:\n%s", path, b, a)
			}
		}
	}
	for path, a := range after {
		if _, existed := before[path]; existed {
			continue
		}
		w, wanted := changed[path]
		if !wanted {
			t.Errorf("%s appeared unexpectedly:\n%s", path, a)
		} else if a != w {
			t.Errorf("%s content:\ngot:\n%s\nwant:\n%s", path, a, w)
		}
	}
}

func TestDiscoverRoot(t *testing.T) {
	t.Run("roadmap claims the root from nested directories", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, map[string]string{"docs/future/roadmap/x.md": "x"})
		deep := filepath.Join(root, "a", "b", "c")
		os.MkdirAll(deep, 0o755)
		if got := DiscoverRoot(deep); got != root {
			t.Errorf("root = %s, want %s", got, root)
		}
	})
	t.Run("nested repo walls the walk before an enclosing roadmap", func(t *testing.T) {
		outer := t.TempDir()
		writeFiles(t, outer, map[string]string{"docs/future/roadmap/x.md": "x"})
		nested := filepath.Join(outer, "vendor", "other")
		os.MkdirAll(filepath.Join(nested, ".git"), 0o755)
		start := filepath.Join(nested, "src")
		os.MkdirAll(start, 0o755)
		if got := DiscoverRoot(start); got != nested {
			t.Errorf("root = %s, want nested %s", got, nested)
		}
	})
	t.Run("worktree file-form .git claims the root", func(t *testing.T) {
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: elsewhere\n"), 0o644)
		start := filepath.Join(root, "sub")
		os.MkdirAll(start, 0o755)
		if got := DiscoverRoot(start); got != root {
			t.Errorf("root = %s, want %s", got, root)
		}
	})
	t.Run("no markers falls back to the start directory", func(t *testing.T) {
		start := t.TempDir()
		if got := DiscoverRoot(start); got != start {
			t.Errorf("root = %s, want %s", got, start)
		}
	})
}

func TestLoad(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/good.md":         "---\ntitle: good\nstate: inbox\ncreated: 2026-07-01\n---\n",
		"docs/future/roadmap/broken.md":       "---\ntitle: broken\nstate: nonsense\ncreated: 2026-07-01\n---\n",
		"docs/future/roadmap/.capture-x.md":   "in flight",
		"docs/future/roadmap/notes.txt":       "not an item",
		"docs/future/roadmap/nested/deep.md":  "not enumerated",
	})
	w := New(root)
	snap, err := w.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Items) != 2 {
		names := []string{}
		for _, it := range snap.Items {
			names = append(names, it.Filename)
		}
		t.Fatalf("items = %v, want good.md and broken.md only", names)
	}
	if snap.OrderVersion != document.VersionAbsent {
		t.Errorf("absent order.yaml must report the absent sentinel, got %s", snap.OrderVersion)
	}
	cards := snap.Cards()
	for _, c := range cards {
		if c.Filename == "broken.md" {
			if len(c.Flags) == 0 || c.Flags[0].Kind != model.FlagMalformed {
				t.Errorf("broken.md should carry the malformed flag, got %v", c.Flags)
			}
			if c.EffectiveLane() != model.Inbox {
				t.Errorf("unreadable state should degrade to inbox, got %s", c.EffectiveLane())
			}
		}
	}
}

func TestLoadRepositoryTiers(t *testing.T) {
	t.Run("missing roadmap directory is repository-level", func(t *testing.T) {
		if _, err := New(t.TempDir()).Load(); err == nil {
			t.Error("want error for missing roadmap directory")
		}
	})
	t.Run("unreadable order.yaml is repository-level", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, map[string]string{
			"docs/future/roadmap/good.md":    "---\ntitle: good\nstate: inbox\ncreated: 2026-07-01\n---\n",
			"docs/future/roadmap/order.yaml": "researching:\n  - a.md\nresearching:\n  - b.md\n",
		})
		if _, err := New(root).Load(); err == nil {
			t.Error("want error for duplicate lane key")
		}
	})
}

func TestCaptureLifecycle(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	t.Run("draft skeleton, directory on demand, finalize exact bytes", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		temp, err := w.CreateDraft("my idea", "")
		if err != nil {
			t.Fatal(err)
		}
		wantSkeleton := "---\ntitle: my idea\nstate: inbox\ncreated: " + today + "\n---\n\n"
		raw, _ := os.ReadFile(temp)
		if string(raw) != wantSkeleton {
			t.Errorf("skeleton:\ngot:\n%s\nwant:\n%s", raw, wantSkeleton)
		}
		fin, err := w.FinalizeDraft(temp)
		if err != nil {
			t.Fatal(err)
		}
		if fin.Outcome != Finalized || fin.Filename != "my-idea.md" {
			t.Fatalf("finalization = %+v", fin)
		}
		landed, _ := os.ReadFile(w.itemPath("my-idea.md"))
		if string(landed) != wantSkeleton {
			t.Error("finalize must land the saved bytes unchanged")
		}
		if _, err := os.Stat(temp); !os.IsNotExist(err) {
			t.Error("temp must be gone after finalize")
		}
	})

	t.Run("title edited in the editor lands under the edited slug", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		temp, _ := w.CreateDraft("first thought", "")
		edited := "---\ntitle:   Better Name (v2)   # renamed mid-edit\nstate: inbox\ncreated: " + today + "\n---\n\nhand-typed body.\n"
		os.WriteFile(temp, []byte(edited), 0o644)
		fin, err := w.FinalizeDraft(temp)
		if err != nil {
			t.Fatal(err)
		}
		if fin.Outcome != Finalized || fin.Filename != "better-name-v2.md" {
			t.Fatalf("finalization = %+v", fin)
		}
		landed, _ := os.ReadFile(w.itemPath("better-name-v2.md"))
		if string(landed) != edited {
			t.Error("hand-formatted frontmatter must survive byte-for-byte")
		}
	})

	t.Run("empty title cancels, temp survives", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		temp, _ := w.CreateDraft("", "words already typed\n")
		fin, err := w.FinalizeDraft(temp)
		if err != nil {
			t.Fatal(err)
		}
		if fin.Outcome != EmptyTitle || fin.TempPath != temp {
			t.Fatalf("finalization = %+v", fin)
		}
		if _, err := os.Stat(temp); err != nil {
			t.Error("temp must survive a canceled capture")
		}
	})

	t.Run("empty slug keeps temp with rename-by-hand outcome", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		temp, _ := w.CreateDraft("((()))", "")
		fin, err := w.FinalizeDraft(temp)
		if err != nil {
			t.Fatal(err)
		}
		if fin.Outcome != EmptySlug {
			t.Fatalf("finalization = %+v", fin)
		}
		if _, err := os.Stat(temp); err != nil {
			t.Error("temp must survive an empty-slug capture")
		}
	})

	t.Run("collision keeps temp and reports both paths", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		writeFiles(t, root, map[string]string{
			"docs/future/roadmap/my-idea.md": "already here",
		})
		temp, _ := w.CreateDraft("my idea", "")
		fin, err := w.FinalizeDraft(temp)
		if err != nil {
			t.Fatal(err)
		}
		if fin.Outcome != Collision || fin.DestPath != w.itemPath("my-idea.md") {
			t.Fatalf("finalization = %+v", fin)
		}
		if _, err := os.Stat(temp); err != nil {
			t.Error("temp must survive a collision")
		}
		onDisk, _ := os.ReadFile(w.itemPath("my-idea.md"))
		if string(onDisk) != "already here" {
			t.Error("collision must not clobber")
		}
	})

	t.Run("capture temps are invisible to enumeration", func(t *testing.T) {
		root := t.TempDir()
		w := New(root)
		if _, err := w.CreateDraft("draft", ""); err != nil {
			t.Fatal(err)
		}
		snap, err := w.Load()
		if err != nil {
			t.Fatal(err)
		}
		if len(snap.Items) != 0 {
			t.Errorf("capture temp leaked into enumeration: %v", snap.Items)
		}
	})
}
