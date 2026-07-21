package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
