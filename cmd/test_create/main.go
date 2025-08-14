package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/order"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/semaphore"
)

func main() {
	wg := sync.WaitGroup{}
	count := 100

	sem := semaphore.NewSemaphore(1)
	l := logger.InitLogger("test", logger.LevelDebug)

	// Init config
	cfg, err := config.New("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	svc := order.NewService(*cfg, &HappyOrderRepository{}, &HappyMessageBroker{}, sem, 0, l)

	ctx := context.Background()
	errManyReqs := 0
	mu := sync.Mutex{}

	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateOrder(ctx, &models.CreateOrder{})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if errors.Is(err, order.ErrTooManyRequest) {
					errManyReqs++
				}
				l.Error(ctx, "test", "", err, "i", i)
			} else {
				l.Info(ctx, "test", "message", "i", i)
			}
		}()
	}

	wg.Wait()
	l.Info(ctx, "test", "results", "total", count, "error-too-many-request-count", errManyReqs)
}

type HappyOrderRepository struct{}

func (h *HappyOrderRepository) Create(ctx context.Context, req *models.CreateOrder, changedBy, notes string) (*models.Order, error) {
	time.Sleep(time.Millisecond)
	return &models.Order{
		ID:     1,
		Number: "ORD_00001",
		Status: "received",
		// ... other minimal fields
	}, nil
}

func (h *HappyOrderRepository) GetAndIncrementSequence(ctx context.Context, date string) (int, error) {
	time.Sleep(time.Millisecond * 100) // simulate
	return 1, nil
}

type HappyMessageBroker struct{}

func (h *HappyMessageBroker) PublishCreateOrder(ctx context.Context, order *models.CreateOrder) error {
	time.Sleep(time.Millisecond * 100) // simulate
	return nil
}

type HappySemaphore struct{}

func (h *HappySemaphore) TryAcquire(timeout time.Duration) bool {
	time.Sleep(time.Millisecond * 100) // simulate
	return true
}
