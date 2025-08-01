package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerRepo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *WorkerRepo {
	return &WorkerRepo{
		db: db,
	}
}

func (r *WorkerRepo) Get() error {
	return nil
}
