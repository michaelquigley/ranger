//go:build !no_ui

// Package ui embeds the built frontend and routes requests between it and
// the API. build with -tags no_ui for a headless binary.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Middleware routes /api/* to the API handler and serves the embedded SPA
// for everything else, falling back to index.html for client-side routes.
func Middleware(apiHandler http.Handler) http.Handler {
	dist, _ := fs.Sub(distFS, "dist")
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			apiHandler.ServeHTTP(w, r)
			return
		}
		if path := strings.TrimPrefix(r.URL.Path, "/"); path != "" {
			if f, err := dist.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
