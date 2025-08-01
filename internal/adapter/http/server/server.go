package server

import (
	"context"
	"errors"
	"fmt"

	"net/http"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

const serverIPAddress = "%s:%d"

type API struct {
	cfg    config.HTTPServer
	router *http.ServeMux
	server *http.Server

	addr string

	log logger.Logger
}

func New(cfg *config.Config, logger logger.Logger) *API {
	addr := fmt.Sprintf(serverIPAddress, cfg.Server.HTTPServer.Host, cfg.Server.HTTPServer.Port)

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

	a.log.Info(ctx, "http_server_stop", "shutting down HTTP server...", "address", a.addr)
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	return nil
}

func (a *API) Run(ctx context.Context, errCh chan<- error) {
	go func() {
		a.log.Info(ctx, "http_server_run", "started http server", "address", a.addr)
		if err := http.ListenAndServe(a.addr, a.router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start HTTP server: %w", err)
			return
		}
	}()
}
