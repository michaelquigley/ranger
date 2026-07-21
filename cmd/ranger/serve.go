package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/spf13/cobra"
	"github.com/michaelquigley/ranger/internal/api"
	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/server"
	"github.com/michaelquigley/ranger/ui"
)

// newServeCmd presents the localhost board over the discovered root — the
// ad-hoc, zero-config path. fail-fast is reserved for repository-level
// failures: the roadmap directory missing or unreadable, an unreadable
// order.yaml — checked once at startup, and again per request, because the
// working tree never stops changing.
func newServeCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "serve the localhost board",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			w, err := discovered()
			if err != nil {
				return err
			}
			cfg, mux, err := serveBootstrap(w.Root(), port)
			if err != nil {
				return err
			}
			return listen(cfg.Port, mux, fmt.Sprintf("serving %s", w.Root()))
		},
	}
	cmd.Flags().IntVar(&port, "port", config.DefaultPort, "listen port on 127.0.0.1")
	return cmd
}

// serveBootstrap synthesizes the one-project config over root and builds
// the shared mux, failing fast on a bad repository — one root is the whole
// point of the process; degradation has nothing to degrade to.
func serveBootstrap(root string, port int) (*config.Config, http.Handler, error) {
	cfg, err := config.New(root, port)
	if err != nil {
		return nil, nil, err
	}
	projects := server.NewProjects(func() (*config.Config, error) { return cfg, nil })
	w, err := projects.Resolve(cfg.Default)
	if err != nil {
		return nil, nil, err
	}
	if _, err := w.Load(); err != nil {
		return nil, nil, err
	}
	mux, err := buildMux(projects)
	if err != nil {
		return nil, nil, err
	}
	return cfg, mux, nil
}

// buildMux assembles the one server surface serve and the daemon share:
// the ogen API at /api/v1, project-scoped roadmap assets at
// /roadmap/{project}/, and the embedded SPA with index fallback.
func buildMux(projects *server.Projects) (http.Handler, error) {
	apiServer, err := api.NewServer(server.New(projects), api.WithPathPrefix("/api/v1"))
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/roadmap/", http.StripPrefix("/roadmap/", server.ProjectAssets(projects)))
	mux.Handle("/", ui.Middleware(apiServer))
	return mux, nil
}

// listen binds 127.0.0.1:port and serves until a signal lands; the bind
// error surfaces plainly — nobody retries or scans.
func listen(port int, handler http.Handler, banner string) error {
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	httpServer := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		dl.Infof("%s at http://%s", banner, addr)
		errCh <- httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		dl.Infof("%v; shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}
}
