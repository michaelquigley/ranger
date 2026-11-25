package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/server"
)

// TestServeBootstrapFailFast keeps serve's error tier visibly distinct
// from the daemon's per-project degradation: one root is the whole point
// of the process, so a bad repository refuses at startup.
func TestServeBootstrapFailFast(t *testing.T) {
	root := t.TempDir() // no roadmap directory
	if _, _, err := serveBootstrap(root, 4114); err == nil || !strings.Contains(err.Error(), "roadmap directory") {
		t.Fatalf("err = %v", err)
	}
}

func TestServeBootstrapServesTheSharedMux(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "future", "roadmap"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, mux, err := serveBootstrap(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 4114 {
		t.Errorf("port must default: %d", cfg.Port)
	}
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("project index over the assembled mux: %d", resp.StatusCode)
	}
}

func TestServeBootstrapV2(t *testing.T) {
	root := t.TempDir()
	future := filepath.Join(root, "docs", "future")
	ranger := filepath.Join(future, ".ranger")
	if err := os.MkdirAll(ranger, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := serveBootstrap(root, 0); err == nil || !strings.Contains(err.Error(), "version.yaml") {
		t.Fatalf("missing marker error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(ranger, "version.yaml"), []byte("version: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		filepath.Join(future, "attachment.txt"): "attachment",
		filepath.Join(ranger, "queue.yaml"):     "invalid view YAML",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, mux, err := serveBootstrap(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]int{
		"/api/v1/projects/" + cfg.Default + "/v2/project":         http.StatusOK,
		"/future/" + cfg.Default + "/attachment.txt":              http.StatusOK,
		"/future/" + cfg.Default + "/.ranger/version.yaml":        http.StatusNotFound,
		"/future/another/attachment.txt":                          http.StatusNotFound,
		"/future/" + cfg.Default + "/..%2fanother/attachment.txt": http.StatusNotFound,
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rec.Code, want)
		}
	}
}

func TestFutureAssetsStayInTheirProject(t *testing.T) {
	cfg := &config.Config{Default: "a"}
	for _, name := range []string{"a", "b"} {
		root := t.TempDir()
		future := filepath.Join(root, "docs", "future")
		if err := os.MkdirAll(future, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(future, "attachment.txt"), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg.Projects = append(cfg.Projects, config.ProjectRef{Name: name, Root: root})
	}
	mux, err := buildMux(server.NewProjects(func() (*config.Config, error) { return cfg, nil }))
	if err != nil {
		t.Fatal(err)
	}
	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	for _, name := range []string{"a", "b"} {
		if rec := get("/future/" + name + "/attachment.txt"); rec.Code != http.StatusOK || rec.Body.String() != name {
			t.Errorf("GET project %s = %d %q", name, rec.Code, rec.Body.String())
		}
	}
	if rec := get("/future/a/..%2fb/attachment.txt"); rec.Code != http.StatusNotFound {
		t.Errorf("cross-project path = %d %q", rec.Code, rec.Body.String())
	}
}
