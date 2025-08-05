package notification

import (
	"context"
	"fmt"

	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

type Service struct {
	reader MessageReceiver
	log    logger.Logger
}

func NewService(reader MessageReceiver, log logger.Logger) *Service {
	return &Service{
		reader: reader,
		log:    log,
	}
}

func (s *Service) Notify(ctx context.Context, errCh chan error) {
	updateCh, err := s.reader.StartListening(ctx)
	if err != nil {
		s.log.Error(ctx, "rabbit_queue_listening", "failed to start listening", err)
		errCh <- err
		return
	}

	for update := range updateCh {
		s.log.Info(ctx, "notification_received",
			fmt.Sprintf("Order %s: %s → %s by %s",
				update.OrderNumber,
				update.OldStatus,
				update.NewStatus,
				update.ChangedBy,
			),
		)
	}
}

func (s *Service) Close() error {
	return s.reader.Close()
}
