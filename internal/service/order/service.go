package order

import (
	"context"
	"math/rand"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

type Service struct {
	orderRepo OrderRepository
	writer    MessageBroker

	log logger.Logger
}

func NewService(repo OrderRepository, writer MessageBroker, log logger.Logger) *Service {
	return &Service{
		orderRepo: repo,
		writer:    writer,
		log:       log,
	}
}

// CreateOrder creates new order
func (s *Service) CreateOrder(ctx context.Context, req *models.CreateOrder) (*models.OrderCreatedInfo, error) {
	s.log.Debug(ctx, types.ActionOrderReceived, "new order received", "")

	today := todayDate()
	number, err := s.orderRepo.GetAndIncrementSequence(ctx, today)
	if err != nil {
		s.log.Error(ctx, types.ActionDBQueryFailed, "failed to get next order sequence.", err)
		// fallback mechanism
		number = getRandomOrderNumber()
	}

	req.SetNumber(today, number)
	req.CalucalteTotalAmount()
	req.CalculatePriority()
	req.Status = types.StatusOrderReceived

	// Store order to database
	order, err := s.orderRepo.Create(ctx, req, "")
	if err != nil {
		s.log.Error(ctx, types.ActionDBTransactionFailed, "failed to create new order", err)
		return nil, err
	}

	// Send request info about publishing order with retry
	if err := retry(5, time.Second, func() error {
		return s.writer.PublishCreateOrder(ctx, req)
	}); err != nil {
		s.log.Error(ctx, types.ActionDBQueryFailed, "order stored to database, but not produced to writer", err)
		return nil, err
	}

	return &models.OrderCreatedInfo{
		Number:      order.Number,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}, nil
}

// Generate a random number between 10000 and 99999 (inclusive)
func getRandomOrderNumber() int {
	return rand.Intn(90000) + 10000
}

func todayDate() string {
	return time.Now().UTC().Format("20060102") // Go's reference time format
}

// retry funciton to attempt request multiple times
func retry(attempts int, delay time.Duration, fn func() error) error {
	var err error

	for i := range attempts {
		err = fn()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			time.Sleep(delay)
		}
	}

	return err
}
