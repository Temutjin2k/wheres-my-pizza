package tracking

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

// Repository contracts

type StatusRepo interface {
	GetCurrent(ctx context.Context, orderNumber string) (models.OrderStatus, error)
	ListOrderHistory(ctx context.Context, orderNumber string) ([]models.OrderHistory, error)
}

type WorkerRepo interface {
	List(ctx context.Context) ([]models.Worker, error)
}
