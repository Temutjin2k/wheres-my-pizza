package app

import (
	"errors"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	svc "github.com/Temutjin2k/wheres-my-pizza/internal/app/service"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"

	"context"
)

var (
	ErrInvalidMode = errors.New("invalid mode")
)

type Service interface {
	Start(ctx context.Context) error
}

type App struct {
	mode    types.ServiceMode
	service Service

	cfg config.Config
	log logger.Logger
}

// NewApplication
func NewApplication(ctx context.Context, cfg config.Config, logger logger.Logger) (*App, error) {
	app := &App{
		mode: cfg.Flags.Mode,
		cfg:  cfg,
		log:  logger,
	}

	if err := app.initService(ctx, app.mode); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	if app.service == nil {
		if err := app.initService(ctx, app.mode); err != nil {
			return err
		}
	}

	if err := app.service.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (app *App) initService(ctx context.Context, mode types.ServiceMode) error {
	var service Service
	var err error
	switch mode {
	case types.ModeOrder:
		service, err = svc.NewOrder(ctx, app.cfg, app.log)
	case types.ModeKitchenWorker:
		service = svc.NewKitchen()
	case types.ModeTracking:
		service = svc.NewTracking()
	case types.ModeNotificationSubscriber:
		service = svc.NewTracking()
	default:
		return ErrInvalidMode
	}

	if err != nil {
		return err
	}

	app.service = service

	return nil
}
