//go:build no_ui

// Package ui embeds the built frontend and routes requests between it and
// the API. this is the headless stub: the API serves, the board doesn't.
package ui

import "net/http"

// Middleware routes /api/* to the API handler; every other path reports
// that this binary was built without the UI.
func Middleware(apiHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			apiHandler.ServeHTTP(w, r)
			return
		}
		http.Error(w, "ranger was built without the ui (no_ui); the api is at /api/v1", http.StatusNotFound)
	})
}
