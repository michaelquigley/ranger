package document

import (
	"reflect"
	"testing"
)

func TestParseFilters(t *testing.T) {
	raw := []byte(`filters:
  - name: active work
    tags: [feature]
    not_tags: [spike]
    milestone: v0.1.x
  - name: quiet
    not_milestone: v0.1.x
    not_subsystems: [flo]
`)
	filters, err := ParseFilters(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := []SavedFilter{
		{Name: "active work", Tags: []string{"feature"}, NotTags: []string{"spike"}, Milestone: "v0.1.x"},
		{Name: "quiet", NotSubsystems: []string{"flo"}, NotMilestone: "v0.1.x"},
	}
	if !reflect.DeepEqual(filters, want) {
		t.Fatalf("got %+v, want %+v", filters, want)
	}
}

func TestParseFiltersEmptyList(t *testing.T) {
	filters, err := ParseFilters([]byte("filters: []\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(filters) != 0 {
		t.Fatalf("got %+v, want empty", filters)
	}
}

func TestParseFiltersRefusals(t *testing.T) {
	cases := map[string]string{
		"syntax":         "filters: [\n",
		"empty name":     "filters:\n  - tags: [feature]\n",
		"duplicate name": "filters:\n  - name: a\n  - name: a\n",
	}
	for label, raw := range cases {
		if _, err := ParseFilters([]byte(raw)); err == nil {
			t.Errorf("%s: expected error", label)
		}
	}
}

func TestRenderFiltersRoundTrip(t *testing.T) {
	filters := []SavedFilter{
		{Name: "active work", Tags: []string{"feature", "story"}, NotTags: []string{"spike"}, Milestone: "v0.1.x"},
		{Name: "needs: quoting", Subsystems: []string{"flo"}, NotMilestone: "v0.2.x"},
	}
	parsed, err := ParseFilters(RenderFilters(filters))
	if err != nil {
		t.Fatalf("round trip parse: %v", err)
	}
	if !reflect.DeepEqual(parsed, filters) {
		t.Fatalf("got %+v, want %+v", parsed, filters)
	}
}

func TestRenderFiltersEmpty(t *testing.T) {
	got := string(RenderFilters(nil))
	if got != "filters: []\n" {
		t.Fatalf("got %q", got)
	}
	parsed, err := ParseFilters(RenderFilters(nil))
	if err != nil || len(parsed) != 0 {
		t.Fatalf("round trip: %+v, %v", parsed, err)
	}
}

func TestRenderFiltersDeterministic(t *testing.T) {
	filters := []SavedFilter{{Name: "a", Tags: []string{"feature"}}}
	if string(RenderFilters(filters)) != string(RenderFilters(filters)) {
		t.Fatal("render is not deterministic")
	}
}
