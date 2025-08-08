package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	httpserver "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/http/server"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/order"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	postgresclient "github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
)

type Order struct {
	postgresDB  *postgresclient.PostgreDB
	httpServer  *httpserver.API
	rabbitOrder *rabbit.OrderProducer

	cfg config.Config
	log logger.Logger
}

func NewOrder(ctx context.Context, cfg config.Config, log logger.Logger) (*Order, error) {
	// Postgres database
	db, err := postgresclient.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error(ctx, "db_connect", "failed to connect postgres", err)
		return nil, fmt.Errorf("failed to connect postgres: %v", err)
	}
	log.Info(ctx, types.ActionDBConnected, "connected to the database")

	orderRepo := postgres.NewOrderRepo(db.Pool)

	producer, err := rabbit.NewOrderProducer(ctx, cfg.RabbitMQ, log)
	if err != nil {
		log.Error(ctx, types.ActionRabbitMQConnected, "failed to connect rabbitmq", err)
		return nil, fmt.Errorf("failed to connect rabbitmq: %v", err)
	}

	orderService := order.NewService(orderRepo, producer, log)

	api := httpserver.New(cfg, orderService, nil, log)
	return &Order{
		postgresDB:  db,
		httpServer:  api,
		rabbitOrder: producer,

		cfg: cfg,
		log: log,
	}, nil
}

func (s *Order) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	s.httpServer.Run(ctx, errCh)

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		s.log.Info(ctx, types.ActionServiceStop, "shuting down application", "signal", sig.String())

		s.close(ctx)
		s.log.Info(ctx, types.ActionServiceStop, "graceful shutdown completed!")
	}

	return nil
}

func (s *Order) close(ctx context.Context) {
	if err := s.httpServer.Stop(ctx); err != nil {
		s.log.Warn(ctx, types.ActionServiceStop, "failed to shutdown HTTP server")
	}
	if err := s.rabbitOrder.Close(); err != nil {
		s.log.Warn(ctx, types.ActionServiceStop, "failed to close rabbitMQ order client connection")
	}

	s.postgresDB.Pool.Close()
}
