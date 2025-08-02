package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type workerRepository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *workerRepository {
	return &workerRepository{
		db: db,
	}
}

func (r *workerRepository) Get() error {
	return nil
}
