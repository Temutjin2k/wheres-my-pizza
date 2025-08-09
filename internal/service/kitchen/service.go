package kitchen

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

type (
	KitchenWorker struct {
		repo     WorkerRepository
		consumer OrderConsumer

		isWorking bool
		worker    *worker
		log       logger.Logger
	}

	worker struct {
		name       string
		orderTypes []string // Comma-separated list of order types the worker can handle (e.g., `dine_in,takeout`). If omitted, handles all.
		heartbeat  time.Duration
	}
)

// NewWorker creates new instance of kitchen-worker service
func NewWorker(repo WorkerRepository, consumer OrderConsumer, workerName string, orderTypes []string, heatbeat time.Duration, log logger.Logger) *KitchenWorker {
	return &KitchenWorker{
		repo:     repo,
		consumer: consumer,
		worker: &worker{
			name:       workerName,
			orderTypes: orderTypes,
			heartbeat:  heatbeat,
		},

		log: log,
	}
}

// Work works
func (s *KitchenWorker) Work(ctx context.Context, errCh chan<- error) {
	// check if it's already working.
	if s.isWorking {
		errCh <- errors.New("worker is already working")
		return
	}
	s.isWorking = true
	defer func() {
		s.isWorking = false
	}()

	// turning all order types that worker can handle into string to store in database.
	workerOrderTypes := strings.Join(s.worker.orderTypes, ",")

	// Marking worker as online
	if err := s.repo.MarkOnline(ctx, s.worker.name, workerOrderTypes); err != nil {
		s.log.Error(ctx, types.ActionWorkerRegistrationFailed, "failed to mark online worker", err, "worker-name", s.worker.name)
		errCh <- err
		return
	}
	s.log.Info(ctx, types.ActionWorkerRegistered, "worker was successfully marked online", "worker-name", s.worker.name, "order-types", workerOrderTypes)

	// Start consumining
	wg := sync.WaitGroup{}
	for _, ot := range s.worker.orderTypes {
		wg.Add(1)
		go func(ot string) {
			defer func() {
				wg.Done()
				s.log.Info(ctx, "stop_consume", "stoped consume queue", "order-type", ot)
			}()

			if err := s.consumer.Consume(ctx, ot, s.proccesOrder); err != nil {
				errCh <- fmt.Errorf("failed to start conuming: %w", err)
				return
			}
		}(ot)
	}

	go s.heartbeatLoop(ctx, s.worker.heartbeat)

	wg.Wait()
}

// proccesOrder processes created order
func (s *KitchenWorker) proccesOrder(req *models.CreateOrder) error {
	fmt.Println("BEFORE", req)
	time.Sleep(types.GetSimulateDuration(req.Type))
	fmt.Println("AFTER", req)

	return nil
}

func (s *KitchenWorker) heartbeatLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.repo.UpdateLastSeen(ctx, s.worker.name); err != nil {
				s.log.Error(ctx, types.ActionDBQueryFailed, "failed to update last seen on worker", err, "worker-name", s.worker.name)
			}
			s.log.Debug(ctx, types.ActionHeartbeatSent, "heatbear was sent", "worker-name", s.worker.name)
		}
	}
}
