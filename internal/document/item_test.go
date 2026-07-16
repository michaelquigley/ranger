package document

import (
	"strings"
	"testing"

	"git.hq.quigley.com/products/vane/internal/model"
)

// the hand-formatted fixture: comments, unusual spacing, unknown fields,
// quoted and unquoted scalars, and a body that must survive verbatim.
const itemFixture = `---
# provenance header, hand-written
title: Retry Semantics (v2)   # the working name
state:    researching
created: 2026-07-13
tags: [retry, "zrok"]
source: github:openziti/zrok#412
milestone: v0.1.x
x-priority-notes: |
  someone else's extension data
  with: colons and # hashes
log:
  - stamp: 2026-07-13
    note: spec drawn
---

body prose here.

more body.
`

func TestParseItemValid(t *testing.T) {
	d := ParseItem([]byte(itemFixture))
	if d.Malformed {
		t.Fatalf("fixture parsed malformed: %v", d.Diagnostics)
	}
	if !d.TitleOK || d.Title != "Retry Semantics (v2)" {
		t.Errorf("title = %q (ok=%v)", d.Title, d.TitleOK)
	}
	if d.State != model.Researching {
		t.Errorf("state = %q", d.State)
	}
	if d.Created != "2026-07-13" {
		t.Errorf("created = %q", d.Created)
	}
	if len(d.Tags) != 2 || d.Tags[0] != "retry" || d.Tags[1] != "zrok" {
		t.Errorf("tags = %v", d.Tags)
	}
	if d.Source != "github:openziti/zrok#412" {
		t.Errorf("source = %q", d.Source)
	}
	if d.Milestone != "v0.1.x" {
		t.Errorf("milestone = %q", d.Milestone)
	}
	if len(d.Log) != 1 || d.Log[0].Stamp != "2026-07-13" || d.Log[0].Note != "spec drawn" {
		t.Errorf("log = %v", d.Log)
	}
}

func TestSetStateSurgical(t *testing.T) {
	d := ParseItem([]byte(itemFixture))
	got, err := d.SetState(model.Building)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(itemFixture, "state:    researching", "state: building", 1)
	if string(got) != want {
		t.Errorf("SetState diff beyond the state line:\ngot:\n%s\nwant:\n%s", got, want)
	}
	if reparsed := ParseItem(got); reparsed.State != model.Building || reparsed.Malformed {
		t.Errorf("patched doc reparsed: state=%q malformed=%v %v", reparsed.State, reparsed.Malformed, reparsed.Diagnostics)
	}
}

func TestSetTitlePreservesInlineComment(t *testing.T) {
	d := ParseItem([]byte(itemFixture))
	got, err := d.SetTitle("Retry Semantics (v3)")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(itemFixture,
		"title: Retry Semantics (v2)   # the working name",
		"title: Retry Semantics (v3) # the working name", 1)
	if string(got) != want {
		t.Errorf("SetTitle diff beyond the title line:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSetTitleReplacesBlockScalarRange(t *testing.T) {
	fixture := `---
title: |
  a long title
  spanning lines
state: inbox
created: 2026-07-13
---
body stays.
`
	d := ParseItem([]byte(fixture))
	if d.Malformed {
		t.Fatalf("fixture malformed: %v", d.Diagnostics)
	}
	got, err := d.SetTitle("shorter now")
	if err != nil {
		t.Fatal(err)
	}
	want := `---
title: shorter now
state: inbox
created: 2026-07-13
---
body stays.
`
	if string(got) != want {
		t.Errorf("block scalar range not replaced whole:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSetTitleSparesCommentAfterBlockScalar(t *testing.T) {
	fixture := `---
title: |
  a long title
# keep this comment
state: inbox
created: 2026-07-13
---
body stays.
`
	d := ParseItem([]byte(fixture))
	if d.Malformed {
		t.Fatalf("fixture malformed: %v", d.Diagnostics)
	}
	got, err := d.SetTitle("shorter now")
	if err != nil {
		t.Fatal(err)
	}
	want := `---
title: shorter now
# keep this comment
state: inbox
created: 2026-07-13
---
body stays.
`
	if string(got) != want {
		t.Errorf("comment outside the scalar was swallowed:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSetTitleKeepsCommentLookingScalarContent(t *testing.T) {
	fixture := `---
title: |
  a long title
  # literally part of the title
state: inbox
created: 2026-07-13
---
`
	d := ParseItem([]byte(fixture))
	if d.Title != "a long title\n# literally part of the title\n" {
		t.Fatalf("title = %q", d.Title)
	}
	got, err := d.SetTitle("shorter now")
	if err != nil {
		t.Fatal(err)
	}
	want := `---
title: shorter now
state: inbox
created: 2026-07-13
---
`
	if string(got) != want {
		t.Errorf("scalar content stranded:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSetTitleHonorsIndentationIndicator(t *testing.T) {
	fixture := `---
title: |2
    over-indented first line
  # content at the declared indent
state: inbox
created: 2026-07-13
---
`
	d := ParseItem([]byte(fixture))
	if d.Title != "  over-indented first line\n# content at the declared indent\n" {
		t.Fatalf("title = %q", d.Title)
	}
	got, err := d.SetTitle("shorter now")
	if err != nil {
		t.Fatal(err)
	}
	want := `---
title: shorter now
state: inbox
created: 2026-07-13
---
`
	if string(got) != want {
		t.Errorf("declared-indent content stranded:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSetTitleQuotesUnsafeScalars(t *testing.T) {
	d := ParseItem([]byte(itemFixture))
	got, err := d.SetTitle("watch: the # signs")
	if err != nil {
		t.Fatal(err)
	}
	reparsed := ParseItem(got)
	if reparsed.Malformed {
		t.Fatalf("unsafe title broke the document: %v", reparsed.Diagnostics)
	}
	if reparsed.Title != "watch: the # signs" {
		t.Errorf("title round-trip = %q", reparsed.Title)
	}
}

func TestParseItemMalformedClassification(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		malformed bool
	}{
		{"duplicate claimed key", "---\ntitle: a\ntitle: b\nstate: inbox\ncreated: 2026-07-13\n---\n", true},
		{"duplicate unknown key", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nx-a: 1\nx-a: 2\n---\n", false},
		{"broken syntax inside unknown field", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nx-b: [unclosed\n---\n", true},
		{"missing required field", "---\ntitle: a\nstate: inbox\n---\n", true},
		{"invalid state", "---\ntitle: a\nstate: shipped\ncreated: 2026-07-13\n---\n", true},
		{"invalid created", "---\ntitle: a\nstate: inbox\ncreated: 2026-7-3\n---\n", true},
		{"claimed field violating shape", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\ntags: {oops: true}\n---\n", true},
		{"no frontmatter at all", "just prose\n", true},
		{"unterminated fence", "---\ntitle: a\n", true},
		{"dotted fence close", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\n...\nbody\n", false},
		{"alias into unknown anchor", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nx-shared: &st [a, b]\ntags: *st\n---\n", false},
		{"invalid log stamp", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nlog:\n  - stamp: soon\n    note: n\n---\n", true},
		{"null claimed value", "---\ntitle:\nstate: inbox\ncreated: 2026-07-13\n---\n", true},
		{"milestone violating shape", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nmilestone: [v0.1.x]\n---\n", true},
		{"duplicate milestone", "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nmilestone: v0.1.x\nmilestone: v0.5.x\n---\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ParseItem([]byte(tt.raw))
			if d.Malformed != tt.malformed {
				t.Errorf("malformed = %v, want %v (diagnostics: %v)", d.Malformed, tt.malformed, d.Diagnostics)
			}
		})
	}
}

func TestMalformedIsVerdictNotBlackout(t *testing.T) {
	raw := "---\ntitle: still here\nstate: researching\ncreated: 2026-07-13\ntags: {oops: true}\n---\nbody\n"
	d := ParseItem([]byte(raw))
	if !d.Malformed {
		t.Fatal("broken tags should flag the item")
	}
	if d.State != model.Researching {
		t.Errorf("broken sibling erased state: %q", d.State)
	}
	if !d.TitleOK || d.Title != "still here" {
		t.Errorf("broken sibling erased title: %q (ok=%v)", d.Title, d.TitleOK)
	}
	if d.Created != "2026-07-13" {
		t.Errorf("broken sibling erased created: %q", d.Created)
	}
}

func TestUnreadableStateFallsToZero(t *testing.T) {
	raw := "---\ntitle: a\nstate: shipped\ncreated: 2026-07-13\n---\n"
	d := ParseItem([]byte(raw))
	if d.State != "" {
		t.Errorf("invalid state should be unreadable, got %q", d.State)
	}
}

func TestAliasedTagsResolve(t *testing.T) {
	raw := "---\ntitle: a\nstate: inbox\ncreated: 2026-07-13\nx-shared: &st [one, two]\ntags: *st\n---\n"
	d := ParseItem([]byte(raw))
	if d.Malformed {
		t.Fatalf("aliased tags flagged: %v", d.Diagnostics)
	}
	if len(d.Tags) != 2 || d.Tags[0] != "one" || d.Tags[1] != "two" {
		t.Errorf("tags = %v", d.Tags)
	}
}

func TestPatchRefusedWithoutField(t *testing.T) {
	d := ParseItem([]byte("---\ntitle: a\ncreated: 2026-07-13\n---\n"))
	if _, err := d.SetState(model.Building); err == nil {
		t.Error("SetState with no state: field should refuse")
	}
	dup := ParseItem([]byte("---\ntitle: a\ntitle: b\nstate: inbox\ncreated: 2026-07-13\n---\n"))
	if _, err := dup.SetTitle("c"); err == nil {
		t.Error("SetTitle with a duplicated title: field should refuse")
	}
}
