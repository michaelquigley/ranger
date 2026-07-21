package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFull(t *testing.T) {
	cfg, err := Load(write(t, `
projects:
  - root: /repos/ranger
  - root: /repos/other
    name: archive
default: archive
port: 4200
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 2 {
		t.Fatalf("projects: %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "ranger" || cfg.Projects[0].Root != "/repos/ranger" {
		t.Errorf("first entry: %+v", cfg.Projects[0])
	}
	if cfg.Projects[1].Name != "archive" || cfg.Projects[1].Root != "/repos/other" {
		t.Errorf("second entry: %+v", cfg.Projects[1])
	}
	if cfg.Default != "archive" {
		t.Errorf("default: %q", cfg.Default)
	}
	if cfg.Port != 4200 {
		t.Errorf("port: %d", cfg.Port)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(write(t, `
projects:
  - root: "/repos/My Repo"
  - root: /repos/other
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Projects[0].Name != "my-repo" {
		t.Errorf("basename slugification: %q", cfg.Projects[0].Name)
	}
	if cfg.Default != "my-repo" {
		t.Errorf("default should be the first entry: %q", cfg.Default)
	}
	if cfg.Port != 4114 {
		t.Errorf("port default: %d", cfg.Port)
	}
}

func TestLoadTildeExpansion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg, err := Load(write(t, `
projects:
  - root: ~/repos/ranger
`))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "repos", "ranger"); cfg.Projects[0].Root != want {
		t.Errorf("root: %q, want %q", cfg.Projects[0].Root, want)
	}
}

func TestLoadRejections(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"absent projects", "port: 4114\n", "at least one root"},
		{"empty projects", "projects: []\n", "at least one root"},
		{"relative root", "projects:\n  - root: repos/ranger\n", "not absolute"},
		{"empty-slug basename", "projects:\n  - root: /repos/☂☂☂\n", "slugifies to nothing"},
		{"non-slug name", "projects:\n  - root: /repos/ranger\n    name: My Name\n", "not slug-shaped"},
		{"duplicate names", "projects:\n  - root: /a/ranger\n  - root: /b/ranger\n", `share the name "ranger"`},
		{"post-slug collision", "projects:\n  - root: \"/a/My Repo\"\n  - root: /b/x\n    name: my-repo\n", `share the name "my-repo"`},
		{"unknown default", "projects:\n  - root: /repos/ranger\ndefault: missing\n", "names no configured project"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Load(write(t, c.content))
			if err == nil {
				t.Fatal("expected a load error")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
}

func TestLoadDuplicateErrorNamesBothRoots(t *testing.T) {
	_, err := Load(write(t, "projects:\n  - root: /a/ranger\n  - root: /b/ranger\n"))
	if err == nil {
		t.Fatal("expected a load error")
	}
	for _, root := range []string{"/a/ranger", "/b/ranger"} {
		if !strings.Contains(err.Error(), root) {
			t.Errorf("error %q does not name %s", err, root)
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "config.yaml")); err == nil {
		t.Fatal("expected a load error")
	}
}
