package order

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

// Repository contract
type OrderRepository interface {
	Create(ctx context.Context, req *models.CreateOrder) (*models.Order, error)
	GetAndIncrementSequence(ctx context.Context, date string) (int, error)
}
