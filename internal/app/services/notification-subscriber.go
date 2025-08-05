package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/notification"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	pkg "github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
)

type NotificationSubsriber struct {
	service Service

	cfg config.Config
	log logger.Logger
}

type Service interface {
	Notify(ctx context.Context, errCh chan error)
	Close() error
}

func NewNotificationSubscriber(ctx context.Context, cfg config.Config, log logger.Logger) (*NotificationSubsriber, error) {
	client, err := pkg.New(ctx, cfg.RabbitMQ.Conn)
	if err != nil {
		log.Error(ctx, "rabbit_connect", "failed to connect rabbitmq", err)
		return nil, fmt.Errorf("failed to connect rabbitmq: %v", err)
	}

	reader := rabbit.NewNotificationSubscriber(client, cfg.RabbitMQ, log)
	service := notification.NewService(reader, log)

	return &NotificationSubsriber{
		service: service,
		cfg:     cfg,
		log:     log,
	}, nil
}

func (s *NotificationSubsriber) Start(ctx context.Context) error {
	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	errCh := make(chan error, 1)
	go s.service.Notify(ctx, errCh)

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		s.log.Info(ctx, types.ActionServiceStop, "shuting down application", "signal", sig.String())

		if err := s.close(); err != nil {
			s.log.Error(ctx, types.ActionServiceStop, "failed to close service", err)
		}

		s.log.Info(ctx, types.ActionServiceStop, "graceful shutdown completed!")
	}

	return nil
}

func (s *NotificationSubsriber) close() error {
	return s.service.Close()
}
