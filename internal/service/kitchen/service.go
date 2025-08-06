package kitchen

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
)

type (
	Service struct {
		repo WorkerRepository

		worker *worker
	}

	worker struct {
		Name        string
		OrderType   string
		cookingTime time.Duration // Simulating time
	}
)

func NewService(repo WorkerRepository, workerName, orderType string) (*Service, error) {
	if err := validateWorkerName(workerName); err != nil {
		return nil, err
	}

	worker := &worker{}
	switch orderType {
	case types.OrderTypeDineIn:
		worker = newWorker(workerName, types.OrderTypeDineIn, types.CookingTimeDineIn)
	case types.OrderTypeTakeOut:
		worker = newWorker(workerName, types.OrderTypeTakeOut, types.CookingTimeDineIn)
	case types.OrderTypeDelivery:
		worker = newWorker(workerName, types.OrderTypeDelivery, types.CookingTimeDineIn)
	default:
		return nil, errors.New("invalid order type")
	}

	return &Service{
		repo:   repo,
		worker: worker,
	}, nil
}

func newWorker(name, OrderType string, cookingTime time.Duration) *worker {
	return &worker{
		Name:        name,
		OrderType:   OrderType,
		cookingTime: cookingTime,
	}
}

func validateWorkerName(name string) error {
	// Check length (1-100 characters)
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return errors.New("worker name length must be between 1-100 characters")
	}
	return nil
}

func (s *Service) Work(ctx context.Context, errCh chan<- error) {
	if err := s.repo.RegisterOnline(ctx, s.worker.Name, s.worker.OrderType); err != nil {
		errCh <- err
		return
	}

}
