package server

import (
	"context"
	"errors"
	"fmt"

	"net/http"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/config"
	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/pkg/logger"
)

const serverIPAddress = ":%d" // Change to 0.0.0.0 for external access

type API struct {
	cfg    config.HTTPServer
	router *http.ServeMux
	server *http.Server

	addr string

	log logger.Logger
}

func New(cfg *config.Config, logger logger.Logger) *API {
	addr := fmt.Sprintf(serverIPAddress, cfg.Server.HTTPServer.Port)

	// Setup routes
	mux := http.NewServeMux()

	api := &API{
		router: mux,

		addr: addr,
		cfg:  cfg.Server.HTTPServer,
		log:  logger,
	}

	api.server = &http.Server{
		Addr:    api.addr,
		Handler: api.router,
	}

	api.setupRoutes()

	return api
}

func (a *API) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	a.log.Info(ctx, "Shutting down HTTP server...", "Address", a.addr)
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	return nil
}

func (a *API) Run(ctx context.Context, errCh chan<- error) {
	go func() {
		a.log.Info(ctx, "Started http server", "Address", a.addr)
		if err := http.ListenAndServe(a.addr, a.router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start HTTP server: %w", err)
			return
		}
	}()
}
