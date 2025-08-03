package order

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

type OrderRepository interface {
	Create(ctx context.Context, req *models.CreateOrder) (*models.Order, error)
	GetAndIncrementSequence(ctx context.Context, date string) (int, error)
}

type MessageBroker interface {
	Publish(ctx context.Context) error
}
