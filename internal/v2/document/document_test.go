package document

import (
	"slices"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/v2/model"
)

func TestItemFieldsValidateIndependently(t *testing.T) {
	raw := []byte(`---
title: A story
state: pending
created: 2026-11-24
since: 2026-11-31
kind: unknown
tags: [spike]
subsystems: [viewer]
source: ../readme.md
milestone: beta
log:
  - stamp: 2026-11-24
    note: journal entry
blocked_on: ["", dependency]
needs: [review]
unknown: ignored
unknown: still ignored
---

A body.
`)
	item, body := ParseItem(raw, "a-story")
	if item.Lane() != model.Pending || !item.TitleOK || item.Title != "A story" ||
		item.Created != "2026-11-24" || item.Since != "" || item.Kind != "" {
		t.Errorf("item = %+v", item)
	}
	if !slices.Equal(item.Diagnostics, []string{`invalid kind: "unknown"`, `invalid since date: "2026-11-31"`, `invalid blocked_on target: ""`}) {
		t.Errorf("diagnostics = %v", item.Diagnostics)
	}
	if len(item.Log) != 1 || item.Log[0].Note != "journal entry" || len(item.BlockedOn) != 1 ||
		item.BlockedOn[0].Slug != "dependency" || !slices.Equal(item.Needs, []string{"review"}) ||
		!slices.Equal(item.Tags, []string{"spike"}) || !slices.Equal(item.Subsystems, []string{"viewer"}) ||
		item.Source != "../readme.md" || item.Milestone != "beta" {
		t.Errorf("optional fields = %+v", item)
	}
	if body != "\nA body.\n" {
		t.Errorf("body = %q", body)
	}
	for _, state := range []string{"state: researching\n", "state: building\nstate: pending\n", "state: [building]\n"} {
		item, _ := ParseItem([]byte("---\ntitle: Valid\n"+state+"created: 2026-11-24\n---\n"), "valid")
		if item.Lane() != model.Inbox || !item.TitleOK || len(item.Diagnostics) != 1 {
			t.Errorf("bad state %q: item = %+v", state, item)
		}
	}
}

func TestViewFormsAndLineNumbers(t *testing.T) {
	view, err := ParseView([]byte("# ranked\n- first\n\n- second # hold\n"))
	want := []model.ViewEntry{{Slug: "first", Line: 2}, {Slug: "second", Line: 4}}
	if err != nil || view.Flow || !slices.Equal(view.Entries, want) {
		t.Fatalf("block view = %+v, error = %v", view, err)
	}
	view, err = ParseView([]byte("# ranked\n[first, second]\n"))
	want = []model.ViewEntry{{Slug: "first", Line: 2}, {Slug: "second", Line: 2}}
	if err != nil || !view.Flow || !slices.Equal(view.Entries, want) {
		t.Errorf("flow view = %+v, error = %v", view, err)
	}
	for _, raw := range []string{"", "\n# empty\n"} {
		if view, err := ParseView([]byte(raw)); err != nil || len(view.Entries) > 0 {
			t.Errorf("empty view = %+v, error = %v", view, err)
		}
	}
	for _, raw := range []string{"key: value", "-\n", "- [nested]\n", "- folded\n  item\n"} {
		if _, err := ParseView([]byte(raw)); err == nil {
			t.Errorf("parsed invalid view %q", raw)
		}
	}
}

func TestPlanAndStageValidation(t *testing.T) {
	plan, err := ParsePlan([]byte("# Plan without frontmatter\n"))
	if err != nil || plan.Mercurius != "" || len(plan.Stages) > 0 {
		t.Fatalf("body-only plan = %+v, error = %v", plan, err)
	}
	if _, err := ParsePlan([]byte("---\nstages: [\n---\n")); err == nil {
		t.Error("invalid plan parsed")
	}
	for _, facts := range []string{"status: derived", "accepted: 2026-11-31", "status: started\nstatus: built"} {
		if _, err := ParseStage([]byte("---\n" + facts + "\n---\n")); err == nil {
			t.Errorf("invalid stage parsed: %q", facts)
		}
	}
	stage, err := ParseStage([]byte("---\ntitle: Stage\nstatus: built\naccepted: 2026-11-24\n---\n"))
	if err != nil || stage.Title != "Stage" || stage.Status != model.Built || stage.Accepted != "2026-11-24" {
		t.Errorf("stage = %+v, error = %v", stage, err)
	}
}

func TestBodyUsesRelativeInlineLinks(t *testing.T) {
	body := strings.Join([]string{
		"1. [ ] open", "+ [x] done",
		"~~~", "## ignored heading", "- [ ] ignored", "[ignored](../fenced/index.md)", "~~~",
		"3) [ ] still in prompt",
		"`[code](../inline-code/index.md)` and `prose`",
		"[label with `code`](../first/index.md#heading) and [second](../second/design.md)",
		"![image](../image/file.png) [reference][ref] [ref]: ../ref/index.md",
		"[absolute](/docs/future/abs/index.md) [web](https://example.com/docs/future/web/index.md)",
		"[outside](../../outside/index.md) [same](../current/index.md) [hidden](../.ranger/queue.yaml)",
		"## supporting notes", "- [ ] ignored checkbox, [included link](../third/)",
	}, "\n")
	count, targets := scanBody(body, "current")
	if count != (model.Checklist{Open: 2, Total: 3}) || !slices.Equal(targets, []string{"first", "second", "third"}) {
		t.Errorf("checklist = %+v, targets = %v", count, targets)
	}
}
