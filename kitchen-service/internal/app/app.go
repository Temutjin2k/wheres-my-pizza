package app

import (
	"kitchen-service/config"
	postgresRepo "kitchen-service/internal/adapter/postgres"
	"kitchen-service/pkg/logger"
	"kitchen-service/pkg/postgres"

	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	postgresDB *postgres.PostgreDB

	log logger.Logger
}

func NewApplication(ctx context.Context, cfg *config.Config, logger logger.Logger) (*App, error) {
	// Database
	db, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %v", err)
	}

	// Repo instance
	_ = postgresRepo.NewRepo(db.Pool)

	app := &App{
		postgresDB: db,

		log: logger,
	}
	return app, nil
}

func (app *App) Close(ctx context.Context) {
	// Closing database connection
	app.postgresDB.Pool.Close()
}

func (app *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()

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
