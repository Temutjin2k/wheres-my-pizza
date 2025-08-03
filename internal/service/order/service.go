package order

import (
	"context"
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
	today := todayDate()
	number, err := s.orderRepo.GetAndIncrementSequence(ctx, today)
	if err != nil {
		return nil, err
	}

	req.SetNumber(today, number)
	req.CalucalteTotalAmount()
	req.CalculatePriority()
	req.Status = types.StatusOrderReceived

	order, err := s.orderRepo.Create(ctx, req)
	if err != nil {
		s.log.Error(ctx, types.ActionOrderReceived, "failed to create new order", err)
		return nil, err
	}

	return &models.OrderCreatedInfo{
		Number:      order.Number,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}, nil
}

func todayDate() string {
	return time.Now().UTC().Format("20060102") // Go's reference time format
}
