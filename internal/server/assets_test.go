package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/ranger/internal/config"
)

func TestAssets(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "roadmap")
	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "images", "pic.png"), []byte("png-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(parent, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, ".git", "config"), []byte("git-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(parent, ".git", "config"), filepath.Join(dir, "escape.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("nested-git"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".git/config", filepath.Join(dir, "inroot-link")); err != nil {
		t.Fatal(err)
	}
	h := Assets(dir)

	get := func(target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		return rec
	}

	if rec := get("/images/pic.png"); rec.Code != http.StatusOK || rec.Body.String() != "png-bytes" {
		t.Fatalf("file: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/../outside.txt"); rec.Code == http.StatusOK {
		t.Fatalf("traversal: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/images"); rec.Code != http.StatusNotFound {
		t.Fatalf("directory: got %d", rec.Code)
	}
	if rec := get("/escape.txt"); rec.Code != http.StatusNotFound {
		t.Fatalf("symlink escape: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/.git/config"); rec.Code != http.StatusNotFound {
		t.Fatalf("git component: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/inroot-link"); rec.Code != http.StatusNotFound {
		t.Fatalf("in-root symlink: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/images/missing.png"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing: got %d", rec.Code)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/images/pic.png", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post: got %d", rec.Code)
	}

	linkRoot := filepath.Join(parent, "roadmap-link")
	if err := os.Symlink(filepath.Join(parent, ".git"), linkRoot); err != nil {
		t.Fatal(err)
	}
	linked := Assets(linkRoot)
	rec = httptest.NewRecorder()
	linked.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("symlinked root: got %d %q", rec.Code, rec.Body.String())
	}
}

// TestProjectAssets pins the cross-project properties of the scoped mount:
// /roadmap/a/… can only ever serve from project a's root, an unknown
// project 404s, and no traversal spelling reaching the route can surface
// another project's bytes.
func TestProjectAssets(t *testing.T) {
	roots := map[string]string{}
	for _, name := range []string{"a", "b"} {
		root := t.TempDir()
		dir := filepath.Join(root, "docs", "future", "roadmap", "images")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "x.png"), []byte(name+"-bytes"), 0o644); err != nil {
			t.Fatal(err)
		}
		roots[name] = root
	}
	cfg := &config.Config{
		Projects: []config.ProjectRef{
			{Root: roots["a"], Name: "a"},
			{Root: roots["b"], Name: "b"},
		},
		Default: "a",
		Port:    config.DefaultPort,
	}
	mux := http.NewServeMux()
	mux.Handle("/roadmap/", http.StripPrefix("/roadmap/", ProjectAssets(NewProjects(constantSource(cfg)))))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	get := func(target string) (int, string) {
		t.Helper()
		resp, err := client.Get(srv.URL + target)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, string(body)
	}

	if code, body := get("/roadmap/a/images/x.png"); code != http.StatusOK || body != "a-bytes" {
		t.Fatalf("project a: got %d %q", code, body)
	}
	if code, body := get("/roadmap/b/images/x.png"); code != http.StatusOK || body != "b-bytes" {
		t.Fatalf("project b: got %d %q", code, body)
	}
	if code, _ := get("/roadmap/unknown/images/x.png"); code != http.StatusNotFound {
		t.Fatalf("unknown project: got %d", code)
	}
	// traversal spellings addressed under project a must never answer with
	// project b's bytes — the route-side half of the containment story;
	// the client-side half (browser normalization before the request ever
	// leaves) is pinned by the transform's vitest cases.
	for _, target := range []string{
		"/roadmap/a/../b/images/x.png",
		"/roadmap/a/%2e%2e/b/images/x.png",
		"/roadmap/a/..%2fb/images/x.png",
		"/roadmap/a/..%5cb%5cimages%5cx.png",
	} {
		if code, body := get(target); code == http.StatusOK && body == "b-bytes" {
			t.Errorf("%s served project b's bytes under project a's name", target)
		}
	}
}
