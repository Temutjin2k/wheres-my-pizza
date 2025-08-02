package app

import (
	"github.com/Temutjin2k/wheres-my-pizza/config"
	httpserver "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/http/server"
	postgresRepo "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"

	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	httpServer *httpserver.API
	postgresDB *postgres.PostgreDB

	log logger.Logger
}

// NewApplication parses flag and choose service to start.
func NewApplication(ctx context.Context, cfg *config.Config, logger logger.Logger) (*App, error) {
	// Database
	db, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %v", err)
	}

	// Repo instance
	_ = postgresRepo.NewRepo(db.Pool)

	httpServer := httpserver.New(cfg, nil, logger)

	app := &App{
		httpServer: httpServer,
		postgresDB: db,

		log: logger,
	}
	return app, nil
}

func (app *App) Close(ctx context.Context) {
	// Closing database connection
	app.postgresDB.Pool.Close()

	// Closing http server
	if err := app.httpServer.Stop(ctx); err != nil {
		app.log.Error(ctx, "app_close", "failed to shutdown HTTP service", err)
	}
}

func (app *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()

	// Running http server
	app.httpServer.Run(ctx, errCh)

	app.log.Info(ctx, "app_run", "application started")

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun
	case s := <-shutdownCh:
		app.log.Info(ctx, "app_run", "shuting down application", "signal", s.String())

		app.Close(ctx)
		app.log.Info(ctx, "app_run", "graceful shutdown completed!")
	}

	return nil
}
