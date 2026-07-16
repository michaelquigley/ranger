package workspace

import (
	"errors"
	"strings"
	"testing"

	"git.hq.quigley.com/products/vane/internal/document"
	"git.hq.quigley.com/products/vane/internal/model"
)

// the standard gesture fixture: two ranked researching items, one ranked
// horizon item, one unranked inbox item, and a stale order entry whose file
// is gone — bait for the opportunistic prune.
const (
	retryItem = `---
title: retry semantics
state: researching
created: 2026-07-01
---

retry body.
`
	secondItem = `---
title: second thread
state: researching
created: 2026-07-02
---
`
	frameItem = `---
title: frame composition
state: horizon
created: 2026-07-02
---
`
	captureItem = `---
title: board capture
state: inbox
created: 2026-07-03
---
`
	orderFixture = `# hand-tended
researching:
  - retry-semantics.md   # hot
  - second-thread.md
  - stale-gone.md
horizon:
  - frame-composition.md
`
)

func gestureFixture(t *testing.T) (*Workspace, *Snapshot, map[string]string) {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/retry-semantics.md":   retryItem,
		"docs/future/roadmap/second-thread.md":     secondItem,
		"docs/future/roadmap/frame-composition.md": frameItem,
		"docs/future/roadmap/board-capture.md":     captureItem,
		"docs/future/roadmap/order.yaml":           orderFixture,
	})
	w := New(root)
	snap, err := w.Load()
	if err != nil {
		t.Fatal(err)
	}
	return w, snap, treeState(t, root)
}

func item(t *testing.T, snap *Snapshot, filename string) *Item {
	t.Helper()
	it, ok := snap.Item(filename)
	if !ok {
		t.Fatalf("fixture item %s missing", filename)
	}
	return it
}

func TestTransitionUnrankedTouchesOnlyTheItem(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "board-capture.md")
	if err := w.Transition("board-capture.md", model.Horizon, it.Hash, snap.OrderVersion, nil); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/board-capture.md": strings.Replace(captureItem, "state: inbox", "state: horizon", 1),
	})
}

func TestTransitionRankedRemovesOldLaneEntry(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	if err := w.Transition("retry-semantics.md", model.Building, it.Hash, snap.OrderVersion, nil); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics.md": strings.Replace(retryItem, "state: researching", "state: building", 1),
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
horizon:
  - frame-composition.md
`,
	})
}

func TestTransitionAndPlaceMovesTheEntry(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "board-capture.md")
	pos := 0
	if err := w.Transition("board-capture.md", model.Researching, it.Hash, snap.OrderVersion, &pos); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/board-capture.md": strings.Replace(captureItem, "state: inbox", "state: researching", 1),
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - board-capture.md
  - retry-semantics.md   # hot
  - second-thread.md
horizon:
  - frame-composition.md
`,
	})
}

func TestTransitionAndPlaceRankedMovesAcrossLanes(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	pos := 1
	if err := w.Transition("retry-semantics.md", model.Horizon, it.Hash, snap.OrderVersion, &pos); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics.md": strings.Replace(retryItem, "state: researching", "state: horizon", 1),
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
horizon:
  - frame-composition.md
  - retry-semantics.md
`,
	})
}

func TestTransitionSameStateKeepsRank(t *testing.T) {
	// the gesture leaves no lane, so the ranked cleanup must not fire: a
	// same-state transition is a no-op that keeps priority.
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	if err := w.Transition("retry-semantics.md", model.Researching, it.Hash, snap.OrderVersion, nil); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{})
}

func TestTransitionAndPlaceWithinLaneMovesTheEntry(t *testing.T) {
	// a placement to the same lane is a legitimate move gesture and must
	// keep working.
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "second-thread.md")
	pos := 0
	if err := w.Transition("second-thread.md", model.Researching, it.Hash, snap.OrderVersion, &pos); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
  - retry-semantics.md   # hot
horizon:
  - frame-composition.md
`,
	})
}

func TestTransitionWithPlaceCreatesAbsentOrder(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/board-capture.md": captureItem,
	})
	w := New(root)
	snap, err := w.Load()
	if err != nil {
		t.Fatal(err)
	}
	before := treeState(t, root)
	it := item(t, snap, "board-capture.md")
	pos := 0
	if err := w.Transition("board-capture.md", model.Researching, it.Hash, snap.OrderVersion, &pos); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, root), map[string]string{
		"docs/future/roadmap/board-capture.md": strings.Replace(captureItem, "state: inbox", "state: researching", 1),
		"docs/future/roadmap/order.yaml":       "researching:\n  - board-capture.md\n",
	})
}

func TestReorderMovesOriginalLines(t *testing.T) {
	w, snap, before := gestureFixture(t)
	err := w.Reorder(model.Researching, []string{"second-thread.md", "retry-semantics.md"}, snap.OrderVersion)
	if err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
  - retry-semantics.md   # hot
horizon:
  - frame-composition.md
`,
	})
}

func TestRetitleRankedIsATwoFileGesture(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	newName, err := w.Retitle("retry-semantics.md", "retry semantics v2", it.Hash, snap.OrderVersion)
	if err != nil {
		t.Fatal(err)
	}
	if newName != "retry-semantics-v2.md" {
		t.Fatalf("newName = %s", newName)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics-v2.md": strings.Replace(retryItem, "title: retry semantics", "title: retry semantics v2", 1),
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - retry-semantics-v2.md # hot
  - second-thread.md
horizon:
  - frame-composition.md
`,
	}, "docs/future/roadmap/retry-semantics.md")
}

func TestRetitleToEmptySlugKeepsFilenameAndRank(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	newName, err := w.Retitle("retry-semantics.md", "((()))", it.Hash, snap.OrderVersion)
	if err != nil {
		t.Fatal(err)
	}
	if newName != "retry-semantics.md" {
		t.Fatalf("newName = %s", newName)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics.md": strings.Replace(retryItem, "title: retry semantics", `title: "((()))"`, 1),
	})
}

func TestRetitleCollisionChangesNothing(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	_, err := w.Retitle("retry-semantics.md", "second thread", it.Hash, snap.OrderVersion)
	var collision *document.CollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("want collision, got %v", err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{})
}

func TestRenameToSlugRepairsMismatch(t *testing.T) {
	root := t.TempDir()
	misnamed := "---\ntitle: proper name\nstate: inbox\ncreated: 2026-07-01\n---\n"
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/wrong.md": misnamed,
	})
	w := New(root)
	snap, _ := w.Load()
	before := treeState(t, root)
	it := item(t, snap, "wrong.md")
	newName, err := w.RenameToSlug("wrong.md", it.Hash, snap.OrderVersion)
	if err != nil {
		t.Fatal(err)
	}
	if newName != "proper-name.md" {
		t.Fatalf("newName = %s", newName)
	}
	assertTreeDiff(t, before, treeState(t, root), map[string]string{
		"docs/future/roadmap/proper-name.md": misnamed,
	}, "docs/future/roadmap/wrong.md")
}

func TestRenameToSlugRefusesEmptySlug(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/hand-picked.md": "---\ntitle: \"((()))\"\nstate: inbox\ncreated: 2026-07-01\n---\n",
	})
	w := New(root)
	snap, _ := w.Load()
	it := item(t, snap, "hand-picked.md")
	if _, err := w.RenameToSlug("hand-picked.md", it.Hash, snap.OrderVersion); err == nil {
		t.Error("rename-to-slug must refuse an empty-slug title")
	}
}

func TestMalformedRankedRepairKeepsPriority(t *testing.T) {
	// the work order's named scenario: rename a malformed ranked card,
	// repair its state, priority survives. the researching entry is inert
	// while the state is unreadable — retained through the rename, active
	// again after the repair.
	root := t.TempDir()
	mangled := "---\ntitle: mangled item\nstate: researchin\ncreated: 2026-07-01\n---\n"
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/x.md":       mangled,
		"docs/future/roadmap/order.yaml": "researching:\n  - x.md\n",
	})
	w := New(root)

	snap, _ := w.Load()
	before := treeState(t, root)
	it := item(t, snap, "x.md")
	newName, err := w.Retitle("x.md", "repaired item", it.Hash, snap.OrderVersion)
	if err != nil {
		t.Fatal(err)
	}
	renamed := strings.Replace(mangled, "title: mangled item", "title: repaired item", 1)
	assertTreeDiff(t, before, treeState(t, root), map[string]string{
		"docs/future/roadmap/repaired-item.md": renamed,
		"docs/future/roadmap/order.yaml":       "researching:\n  - repaired-item.md\n",
	}, "docs/future/roadmap/x.md")

	snap, _ = w.Load()
	before = treeState(t, root)
	it = item(t, snap, newName)
	repaired := strings.Replace(renamed, "state: researchin\n", "state: researching\n", 1)
	if err := w.SaveContent(newName, []byte(repaired), it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, root), map[string]string{
		"docs/future/roadmap/repaired-item.md": repaired,
	})

	snap, _ = w.Load()
	board := snap.Board()
	for _, lane := range board.Lanes {
		if lane.State == model.Researching {
			if lane.RankedCount != 1 || lane.Cards[0].Filename != newName {
				t.Errorf("priority did not survive the repair: %+v", lane)
			}
		}
	}
}

func TestSaveContentBodyOnly(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	content := strings.Replace(retryItem, "retry body.", "retry body, expanded with new thinking.", 1)
	if err := w.SaveContent("retry-semantics.md", []byte(content), it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics.md": content,
	})
}

func TestSaveContentStateChangeIsATransition(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	content := strings.Replace(retryItem, "state: researching", "state: building", 1)
	if err := w.SaveContent("retry-semantics.md", []byte(content), it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/retry-semantics.md": content,
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
horizon:
  - frame-composition.md
`,
	})
}

func TestSaveContentRepairOutOfInboxCostsTheInboxRank(t *testing.T) {
	// an unreadable-state card actively ranked in inbox, repaired to a
	// valid other lane: the effective-lane comparison fires and the inbox
	// entry is removed, exactly like any other departure.
	root := t.TempDir()
	mangled := "---\ntitle: mangled\nstate: bogus\ncreated: 2026-07-01\n---\n"
	writeFiles(t, root, map[string]string{
		"docs/future/roadmap/y.md":       mangled,
		"docs/future/roadmap/order.yaml": "inbox:\n  - y.md\n",
	})
	w := New(root)
	snap, _ := w.Load()
	before := treeState(t, root)
	it := item(t, snap, "y.md")
	repaired := strings.Replace(mangled, "state: bogus", "state: horizon", 1)
	if err := w.SaveContent("y.md", []byte(repaired), it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, root), map[string]string{
		"docs/future/roadmap/y.md":       repaired,
		"docs/future/roadmap/order.yaml": "inbox:\n",
	})
}

func TestDeleteUnrankedRemovesOnlyTheFile(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "board-capture.md")
	if err := w.Delete("board-capture.md", it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{}, "docs/future/roadmap/board-capture.md")
}

func TestDeleteRankedRemovesEntries(t *testing.T) {
	w, snap, before := gestureFixture(t)
	it := item(t, snap, "retry-semantics.md")
	if err := w.Delete("retry-semantics.md", it.Hash, snap.OrderVersion); err != nil {
		t.Fatal(err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{
		"docs/future/roadmap/order.yaml": `# hand-tended
researching:
  - second-thread.md
horizon:
  - frame-composition.md
`,
	}, "docs/future/roadmap/retry-semantics.md")
}

func TestDeleteStaleHashChangesNothing(t *testing.T) {
	w, snap, before := gestureFixture(t)
	var conflict *document.ConflictError
	err := w.Delete("retry-semantics.md", document.Hash([]byte("stale")), snap.OrderVersion)
	if !errors.As(err, &conflict) {
		t.Fatalf("stale delete must conflict, got %v", err)
	}
	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{})
}

func TestStaleGuardsRefuseAndChangeNothing(t *testing.T) {
	w, snap, before := gestureFixture(t)
	var conflict *document.ConflictError

	err := w.Transition("retry-semantics.md", model.Building, document.Hash([]byte("stale view")), snap.OrderVersion, nil)
	if !errors.As(err, &conflict) {
		t.Errorf("stale item hash must conflict, got %v", err)
	}

	it := item(t, snap, "retry-semantics.md")
	err = w.Transition("retry-semantics.md", model.Building, it.Hash, document.Hash([]byte("stale order")), nil)
	if !errors.As(err, &conflict) {
		t.Errorf("stale order version must conflict, got %v", err)
	}

	err = w.Reorder(model.Researching, []string{"retry-semantics.md"}, document.VersionAbsent)
	if !errors.As(err, &conflict) {
		t.Errorf("absent expectation against a present order.yaml must conflict, got %v", err)
	}

	assertTreeDiff(t, before, treeState(t, w.Root()), map[string]string{})
}
