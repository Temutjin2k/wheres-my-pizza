package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/kitchen"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	postgresclient "github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
)

type KitchenWorker interface {
	Work(ctx context.Context, errCh chan<- error)
}

type KitchenService struct {
	postgresDB    *postgresclient.PostgreDB
	kitchenWorker KitchenWorker

	cfg config.Config
	log logger.Logger
}

func NewKitchen(ctx context.Context, cfg config.Config, log logger.Logger) (*KitchenService, error) {
	// Postgres database
	db, err := postgresclient.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error(ctx, "db_connect", "failed to connect postgres", err)
		return nil, fmt.Errorf("failed to connect postgres: %v", err)
	}
	log.Info(ctx, types.ActionDBConnected, "connected to the database")

	repo := postgres.NewWorkerRepo(db.Pool)

	kitchenWorker, err := kitchen.NewService(repo, cfg.Services.Kitchen.WorkerName, cfg.Services.Kitchen.OrderType)
	if err != nil {
		return nil, err
	}

	return &KitchenService{
		postgresDB:    db,
		kitchenWorker: kitchenWorker,

		cfg: cfg,
		log: log,
	}, nil
}

func (s *KitchenService) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	// kitchen worker starts to work in goroutine
	s.kitchenWorker.Work(ctx, errCh)

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		s.log.Info(ctx, types.ActionGracefulShutdown, "shuting down application", "signal", sig.String())

		s.close()
		s.log.Info(ctx, types.ActionGracefulShutdown, "graceful shutdown completed!")
	}

	return nil
}

func (s *KitchenService) close() {
	s.postgresDB.Pool.Close()
}
