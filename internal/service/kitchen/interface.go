package kitchen

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

// Repository contract
type WorkerRepository interface {
	// MarkOnline marks worker by inserting (or updating) a record in the
	// workers table with its unique name and type, marking it online.
	MarkOnline(ctx context.Context, name, orderTypes string) error

	// UpdateLastSeen updates last seen timestamp
	UpdateLastSeen(ctx context.Context, name string) error
}

type OrderRepository interface {
	SetStatus(ctx context.Context, orderID, workerName, status, notes string) error
}

type OrderConsumer interface {
	Consume(ctx context.Context, orderType string, handler func(req *models.CreateOrder) error) error
}
