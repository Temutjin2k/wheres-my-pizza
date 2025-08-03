package services

import (
	"context"
	"errors"
)

type NotificationSubsriber struct {
}

func NewNotificationSubscriber() *NotificationSubsriber {
	return &NotificationSubsriber{}
}

func (s *NotificationSubsriber) Start(ctx context.Context) error {
	return errors.ErrUnsupported
}
