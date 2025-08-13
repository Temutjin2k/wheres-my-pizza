package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	httpserver "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/http/server"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/order"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	postgresclient "github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/semaphore"
)

// Feature: Order Service
// The Order Service is the public-facing entry point of the restaurant system. Its primary responsibility is
// to receive new orders from customers via an HTTP API, validate them, store them in the database, and publish
// them to a message queue for the kitchen staff to process. It acts as the gatekeeper, ensuring all incoming
// data is correct and formatted before entering the system.
type Order struct {
	postgresDB *postgresclient.PostgreDB
	httpServer *httpserver.API
	producer   *rabbit.OrderProducer

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

	// RabbitMQ connection
	orderRepo := postgres.NewOrderRepo(db.Pool)

	producer, err := rabbit.NewOrderProducer(ctx, cfg.RabbitMQ, log)
	if err != nil {
		log.Error(ctx, types.ActionRabbitConnectionFailed, "failed to connect rabbitmq", err)
		return nil, fmt.Errorf("failed to connect rabbitmq: %v", err)
	}

	// Semaphore to control maximum number of concurrent orders to process.
	sem := semaphore.NewSemaphore(cfg.Services.Order.MaxConcurrent)

	orderService := order.NewService(orderRepo, producer, sem, log)

	api := httpserver.New(cfg, orderService, nil, log)
	return &Order{
		postgresDB: db,
		httpServer: api,
		producer:   producer,

		cfg: cfg,
		log: log,
	}, nil
}

func (s *Order) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	s.httpServer.Run(ctx, errCh)

	defer func() {
		s.close(ctx)
		s.log.Info(ctx, types.ActionGracefulShutdown, "order service closed")
	}()

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		s.log.Info(ctx, types.ActionGracefulShutdown, "shuting down application", "signal", sig.String())
		return nil
	}
}

func (s *Order) close(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := s.httpServer.Stop(ctx); err != nil {
		s.log.Warn(ctx, types.ActionGracefulShutdown, "failed to shutdown HTTP server")
	}

	if err := s.producer.Close(ctx); err != nil {
		s.log.Warn(ctx, types.ActionGracefulShutdown, "failed to close rabbitMQ order client connection")
	}

	s.postgresDB.Pool.Close()
}
