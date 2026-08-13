package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// newServer creates the HTTP server with the configured address and routes.
func newServer(address string, handler http.Handler) *http.Server {
	return &http.Server{Addr: address, Handler: handler}
}

// serve runs the HTTP server until it fails or a termination signal arrives.
func serve(ctx context.Context, server *http.Server, timeout time.Duration) error {
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	select {
	case err := <-serverErrors:
		return serverError(err)
	case <-ctx.Done():
		return shutdown(server, timeout, serverErrors)
	}
}

// shutdown stops accepting requests and waits for in-flight handlers to finish.
func shutdown(server *http.Server, timeout time.Duration, serverErrors <-chan error) error {
	slog.Info("api stopping", "timeout", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down api server: %w", err)
	}
	if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
		return serverError(err)
	}
	slog.Info("api stopped")
	return nil
}

func serverError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("running api server: %w", err)
}
