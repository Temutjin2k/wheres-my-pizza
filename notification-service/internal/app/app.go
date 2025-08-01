package app

import (
	"github.com/Temutjin2k/wheres-my-pizza/notification-service/config"

	"github.com/Temutjin2k/wheres-my-pizza/notification-service/pkg/logger"

	"context"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "go-template"

type App struct {
	log logger.Logger
}

func NewApplication(ctx context.Context, cfg *config.Config, logger logger.Logger) (*App, error) {
	app := &App{
		log: logger,
	}
	return app, nil
}

func (app *App) Close(ctx context.Context) {
}

func (app *App) Run() error {
	errCh := make(chan error, 1)
	ctx := context.Background()

	app.log.Info(ctx, "application started", "name", serviceName)

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case errRun := <-errCh:
		return errRun
	case s := <-shutdownCh:
		app.log.Info(ctx, "shuting down application", "signal", s.String())

		app.Close(ctx)
		app.log.Info(ctx, "graceful shutdown completed!")
	}

	return nil
}
