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
	"git.hq.quigley.com/products/vane/internal/api"
	"git.hq.quigley.com/products/vane/internal/server"
	"git.hq.quigley.com/products/vane/ui"
)

// newServeCmd presents the localhost board. fail-fast is reserved for
// repository-level failures: the roadmap directory missing or unreadable,
// an unreadable order.yaml — checked once at startup, and again per
// request, because the working tree never stops changing.
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
			if _, err := w.Load(); err != nil {
				return err
			}

			apiServer, err := api.NewServer(server.New(w), api.WithPathPrefix("/api/v1"))
			if err != nil {
				return err
			}
			addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
			httpServer := &http.Server{Addr: addr, Handler: ui.Middleware(apiServer)}

			errCh := make(chan error, 1)
			go func() {
				dl.Infof("serving %s at http://%s", w.Root(), addr)
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
		},
	}
	cmd.Flags().IntVar(&port, "port", 4114, "listen port on 127.0.0.1")
	return cmd
}
