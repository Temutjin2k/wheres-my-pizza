package notification

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

var NotificationStopped = errors.New("notification subscriber stopped")

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

	defer func() {
		s.reader.Close()
		errCh <- NotificationStopped
	}()

	notifHost := "notification-sub "
	if count, err := s.reader.GetListenerCount(); err != nil {
		s.log.Warn(ctx, "rabbit_queue_listeners", "failed to get listeners count", "error", err)
	} else {
		notifHost += strconv.Itoa(count)
	}

	for update := range updateCh {
		s.log.Info(ctx, types.ActionNotificationReceived,
			fmt.Sprintf("Notification for order %s: Status changed from '%s' to '%s' by %s",
				update.OrderNumber,
				update.OldStatus,
				update.NewStatus,
				update.ChangedBy,
			),
		)

		s.log.Info(ctx, types.ActionNotificationReceived,
			fmt.Sprintf("Received status update for order %s", update.OrderNumber),
			"notification_host", notifHost,
			"order_timestamp", update.Timestamp,
			"request_id", update.RequestID,
			"details", fmt.Sprintf(`{ "order_number": "%s", "new_status": "%s" }`, update.OrderNumber, update.NewStatus),
		)
	}
}

func (s *Service) Close() error {
	return s.reader.Close()
}
