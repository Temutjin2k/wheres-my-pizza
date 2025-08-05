package notification

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

type MessageReceiver interface {
	StartListening(ctx context.Context) (chan models.StatusUpdate, error)
	Close() error
}
