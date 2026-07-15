package document

import (
	"strings"
	"testing"

	"git.hq.quigley.com/products/vane/internal/model"
)

// hand-formatted fixture: comments, an unknown lane key, block and flow
// styles.
const orderFixture = `# priorities, hand-tended
researching:
  - retry-semantics.md   # hot
  - board-capture.md
horizon:
  - frame-composition.md
x-someday:
  - maybe.md
building: [flow-a.md, flow-b.md]
`

func mustParseOrder(t *testing.T, raw string) *OrderDoc {
	t.Helper()
	d, err := ParseOrder([]byte(raw))
	if err != nil {
		t.Fatalf("ParseOrder: %v", err)
	}
	return d
}

func TestParseOrder(t *testing.T) {
	d := mustParseOrder(t, orderFixture)
	want := map[model.State][]string{
		model.Researching: {"retry-semantics.md", "board-capture.md"},
		model.Horizon:     {"frame-composition.md"},
		model.Building:    {"flow-a.md", "flow-b.md"},
	}
	for lane, entries := range want {
		got := d.Lanes[lane]
		if strings.Join(got, ",") != strings.Join(entries, ",") {
			t.Errorf("lane %s = %v, want %v", lane, got, entries)
		}
	}
	if len(d.Lanes) != 3 {
		t.Errorf("unknown key leaked into lanes: %v", d.Lanes)
	}
}

func TestParseOrderUnreadable(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"duplicate lane key", "researching:\n  - a.md\nresearching:\n  - b.md\n"},
		{"syntax failure", "researching:\n  - [unclosed\n"},
		{"lane is not a list", "researching: just-a-scalar\n"},
		{"document is not a mapping", "- a.md\n- b.md\n"},
		{"entry is not a filename", "researching:\n  - {oops: true}\n"},
		{"multiline scalar entry", "researching:\n  - |\n    sneaky.md\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseOrder([]byte(tt.raw)); err == nil {
				t.Error("want unreadable, got nil error")
			}
		})
	}
}

func TestParseOrderTolerates(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"duplicate unknown keys", "x-a: 1\nx-a: 2\nresearching:\n  - a.md\n"},
		{"empty lane", "researching:\nhorizon:\n  - a.md\n"},
		{"empty file", ""},
		{"lane aliasing an unknown anchor", "x-pool: &pool\n  - a.md\nresearching: *pool\n"},
		{"comment and blank between entries", "researching:\n  - a.md\n  # note\n\n  - b.md\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseOrder([]byte(tt.raw)); err != nil {
				t.Errorf("want readable, got %v", err)
			}
		})
	}
}

func TestPruneRemovesPrunableAndUnknownBlocks(t *testing.T) {
	d := mustParseOrder(t, orderFixture)
	got := d.Prune(map[model.State][]model.EntryDisposition{
		model.Researching: {model.EntryActive, model.EntryPrunable},
	})
	want := `# priorities, hand-tended
researching:
  - retry-semantics.md   # hot
horizon:
  - frame-composition.md
building: [flow-a.md, flow-b.md]
`
	if string(got) != want {
		t.Errorf("prune diff beyond prunable lines:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestPruneSparesUnknownBlocksUnderRecognizedAlias(t *testing.T) {
	// the guard is coarse: one recognized alias into unknown territory
	// keeps every unknown block, because unknown blocks may alias each
	// other and a partial prune could dangle.
	raw := "x-pool: &pool\n  - a.md\nresearching: *pool\nx-other:\n  - kept-too.md\n"
	d := mustParseOrder(t, raw)
	got := d.Prune(nil)
	if string(got) != raw {
		t.Errorf("unknown blocks must all survive under a recognized alias:\ngot:\n%s\nwant:\n%s", got, raw)
	}
	if _, err := ParseOrder(got); err != nil {
		t.Errorf("our own write rendered the file unreadable: %v", err)
	}
}

func TestPruneNeverDanglesTransitiveAliases(t *testing.T) {
	// an unknown anchor chain: the recognized lane aliases *pool, whose
	// block in turn aliases *name — pruning x-name would strand *name.
	raw := "x-name: &name a.md\nx-pool: &pool\n  - *name\nresearching: *pool\n"
	d := mustParseOrder(t, raw)
	got := d.Prune(nil)
	if string(got) != raw {
		t.Errorf("transitively required block pruned:\ngot:\n%s\nwant:\n%s", got, raw)
	}
	if _, err := ParseOrder(got); err != nil {
		t.Errorf("our own write rendered the file unreadable: %v", err)
	}
}

func TestRewriteLaneTouchesOneLane(t *testing.T) {
	d := mustParseOrder(t, orderFixture)
	got := d.RewriteLane(model.Researching, []string{"board-capture.md", "retry-semantics.md"}, nil)
	want := `# priorities, hand-tended
researching:
  - board-capture.md
  - retry-semantics.md   # hot
horizon:
  - frame-composition.md
x-someday:
  - maybe.md
building: [flow-a.md, flow-b.md]
`
	if string(got) != want {
		t.Errorf("rewrite leaked beyond the lane:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRewriteLanePreservesInertLines(t *testing.T) {
	raw := "researching:\n  - active-1.md\n  - unreadable.md\n  - active-2.md\n"
	d := mustParseOrder(t, raw)
	got := d.RewriteLane(model.Researching, []string{"active-2.md", "active-1.md"},
		[]model.EntryDisposition{model.EntryActive, model.EntryInert, model.EntryActive})
	want := "researching:\n  - active-2.md\n  - active-1.md\n  - unreadable.md\n"
	if string(got) != want {
		t.Errorf("inert entry lost in rewrite:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRemoveEntry(t *testing.T) {
	d := mustParseOrder(t, orderFixture)
	got := d.RemoveEntry(model.Researching, "retry-semantics.md")
	want := strings.Replace(orderFixture, "  - retry-semantics.md   # hot\n", "", 1)
	if string(got) != want {
		t.Errorf("remove diff beyond one line:\ngot:\n%s\nwant:\n%s", got, want)
	}
	if same := d.RemoveEntry(model.Researching, "absent.md"); string(same) != orderFixture {
		t.Error("removing an absent entry must be a no-op")
	}
}

func TestInsertEntryPositions(t *testing.T) {
	d := mustParseOrder(t, orderFixture)

	head := d.InsertEntry(model.Researching, "new.md", 0, nil)
	wantHead := strings.Replace(orderFixture,
		"researching:\n", "researching:\n  - new.md\n", 1)
	if string(head) != wantHead {
		t.Errorf("insert at 0:\ngot:\n%s\nwant:\n%s", head, wantHead)
	}

	mid := d.InsertEntry(model.Researching, "new.md", 1, nil)
	wantMid := strings.Replace(orderFixture,
		"  - board-capture.md\n", "  - new.md\n  - board-capture.md\n", 1)
	if string(mid) != wantMid {
		t.Errorf("insert mid-list:\ngot:\n%s\nwant:\n%s", mid, wantMid)
	}

	end := d.InsertEntry(model.Researching, "new.md", 2, nil)
	wantEnd := strings.Replace(orderFixture,
		"  - board-capture.md\n", "  - board-capture.md\n  - new.md\n", 1)
	if string(end) != wantEnd {
		t.Errorf("insert at len:\ngot:\n%s\nwant:\n%s", end, wantEnd)
	}

	fresh := d.InsertEntry(model.Evaluating, "new.md", 0, nil)
	wantFresh := orderFixture + "evaluating:\n  - new.md\n"
	if string(fresh) != wantFresh {
		t.Errorf("insert into missing lane:\ngot:\n%s\nwant:\n%s", fresh, wantFresh)
	}
}

func TestReplaceFilenameEveryRetainedOccurrence(t *testing.T) {
	raw := "researching:\n  - old.md   # keep me\nhorizon:\n  - other.md\n  - old.md\n"
	d := mustParseOrder(t, raw)
	got := d.ReplaceFilename("old.md", "new-name.md")
	want := "researching:\n  - new-name.md # keep me\nhorizon:\n  - other.md\n  - new-name.md\n"
	if string(got) != want {
		t.Errorf("replace missed an occurrence or moved a position:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFlowLaneOpsMaterializeBlock(t *testing.T) {
	d := mustParseOrder(t, orderFixture)
	got := d.RemoveEntry(model.Building, "flow-a.md")
	want := strings.Replace(orderFixture,
		"building: [flow-a.md, flow-b.md]", "building:\n  - flow-b.md", 1)
	if string(got) != want {
		t.Errorf("flow lane op:\ngot:\n%s\nwant:\n%s", got, want)
	}
	if _, err := ParseOrder(got); err != nil {
		t.Errorf("materialized lane unreadable: %v", err)
	}
}

func TestNewOrder(t *testing.T) {
	got := NewOrder(model.Researching, []string{"a.md", "#weird.md"})
	d, err := ParseOrder(got)
	if err != nil {
		t.Fatalf("fresh order unreadable: %v", err)
	}
	entries := d.Lanes[model.Researching]
	if len(entries) != 2 || entries[0] != "a.md" || entries[1] != "#weird.md" {
		t.Errorf("entries = %v", entries)
	}
}
