package postgres

import (
	"context"
	"fmt"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type workerRepository struct {
	pool *pgxpool.Pool
}

func NewWorkerRepo(pool *pgxpool.Pool) *workerRepository {
	return &workerRepository{
		pool: pool,
	}
}

func (repo *workerRepository) List(ctx context.Context) ([]models.Worker, error) {
	const op = "workerRepository.List"

	query := `
	SELECT 
		name,
		status,
		orders_processed,
		last_seen
	FROM 
		workers;`

	rows, err := repo.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	workers, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Worker, error) {
		var worker models.Worker
		if err := row.Scan(&worker.Name, &worker.Status, &worker.ProcessedOrders, &worker.LastSeen); err != nil {
			return models.Worker{}, err
		}
		return worker, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	if len(workers) == 0 {
		return nil, models.ErrWorkerNotFound
	}
	return workers, nil
}

func (repo *workerRepository) MarkOnline(ctx context.Context, name, orderType string) error {
	return nil
}
