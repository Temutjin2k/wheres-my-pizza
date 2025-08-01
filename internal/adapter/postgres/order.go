package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *orderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Get() error {
	return nil
}
