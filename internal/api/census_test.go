package api

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestContractCensus proves the project scoping over the parsed spec
// rather than assuming it: every operation except GET /projects lives
// under /projects/{project}, exactly once — no unscoped or double-scoped
// survivor.
func TestContractCensus(t *testing.T) {
	raw, err := os.ReadFile("specs/ranger.yml")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Paths) < 2 {
		t.Fatalf("suspiciously few paths parsed: %d", len(spec.Paths))
	}
	for p := range spec.Paths {
		if p == "/projects" {
			continue
		}
		if !strings.HasPrefix(p, "/projects/{project}/") {
			t.Errorf("unscoped path survived the restructure: %s", p)
		}
		if strings.Count(p, "{project}") != 1 {
			t.Errorf("path does not carry exactly one project segment: %s", p)
		}
	}
}
