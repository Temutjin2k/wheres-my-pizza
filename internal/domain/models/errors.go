package models

import "errors"

var (
	ErrWorkerNotFound = errors.New("worker is not found")
	ErrOrderNotFound  = errors.New("order is not found")
)
