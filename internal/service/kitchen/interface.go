package kitchen

import "context"

// Repository contract
type WorkerRepository interface {
	// RegisterOnline register worker by inserting (or updating) a record in the
	// workers table with its unique name and type, marking it online.
	RegisterOnline(ctx context.Context, name, orderType string) error
}
