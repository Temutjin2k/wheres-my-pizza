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

var (
	ErrWorkerStopped = errors.New("worker stopped")
	ErrNilOrder      = errors.New("nil order")
)

type (
	KitchenWorker struct {
		workerRepo WorkerRepository
		orderRepo  OrderRepository
		consumer   Consumer
		producer   Producer

		mu        sync.Mutex
		isWorking bool
		worker    *worker
		log       logger.Logger

		cancel func()
	}

	worker struct {
		name       string
		orderTypes []string // Comma-separated list of order types the worker can handle (e.g., `dine_in,takeout`). If omitted, handles all.
		heartbeat  time.Duration
	}
)

// NewWorker creates new instance of kitchen-worker service
func NewWorker(
	workerRepo WorkerRepository,
	orderRepo OrderRepository,
	consumer Consumer,
	producer Producer,
	workerName string,
	orderTypes []string,
	heartbeat time.Duration,
	log logger.Logger,
) *KitchenWorker {
	return &KitchenWorker{
		workerRepo: workerRepo,
		orderRepo:  orderRepo,
		consumer:   consumer,
		producer:   producer,
		mu:         sync.Mutex{},

		worker: &worker{
			name:       workerName,
			orderTypes: orderTypes,
			heartbeat:  heartbeat,
		},

		log: log,
	}
}

// Work starts consuming orders and proccesses them
func (s *KitchenWorker) Work(ctx context.Context, errCh chan<- error) {
	s.mu.Lock()
	// check if it's already working.
	if s.isWorking {
		errCh <- errors.New("worker is already working")
		return
	}
	s.isWorking = true
	s.mu.Unlock()

	defer func() {
		// stop the worker from consuming and updating database.
		s.Stop(ctx)
		errCh <- ErrWorkerStopped
	}()

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// turning all order types that worker can handle into string to store in database.
	workerOrderTypes := strings.Join(s.worker.orderTypes, ",")

	// Marking worker as online
	if err := s.workerRepo.MarkOnline(ctx, s.worker.name, workerOrderTypes); err != nil {
		s.log.Error(ctx, types.ActionWorkerRegistrationFailed, "failed to mark online worker", err, "worker-name", s.worker.name)
		errCh <- err
		return
	}
	s.log.Info(ctx,
		types.ActionWorkerRegistered,
		"worker was successfully registered",
		"worker-name", s.worker.name,
		"order-types", workerOrderTypes,
		"heartbeat-interval", s.worker.heartbeat,
	)

	wg := sync.WaitGroup{}
	for _, ot := range s.worker.orderTypes {
		wg.Add(1)
		go func(ot string) {
			defer func() {
				wg.Done()
				s.log.Info(ctx, "kitchen_worker_stop_consume", "stopped consuming orders", "order-type", ot)
			}()

			// Start consumining
			err := s.consumer.Consume(ctx, ot, s.proccesOrder)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to start consuming: %w", err):
				default:
					s.log.Error(ctx, "error_channel_full", "failed to send error to channel", err)
				}
			}
		}(ot)
	}

	go func() {
		s.heartbeatLoop(ctx, s.worker.heartbeat)
	}()

	wg.Wait()
}

// Stop worker from work
func (s *KitchenWorker) Stop(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isWorking {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}

	// Mark worker offline
	if err := s.workerRepo.MarkOffline(ctx, s.worker.name); err != nil {
		s.log.Error(ctx, types.ActionDBQueryFailed, "failed to mark worker offline", err, "worker-name", s.worker.name)
		return
	}
	s.isWorking = false

	s.log.Info(ctx, "worker_stop", "stopping worker", "worker-name", s.worker.name)
}

// proccesOrder processes created order
func (s *KitchenWorker) proccesOrder(ctx context.Context, req *models.CreateOrder) error {
	if req == nil {
		s.log.Error(ctx, types.ActionValidationFailed, "nil order to proccess", ErrNilOrder, "worker-name", s.worker.name)
		return ErrNilOrder
	}

	cookingTime := types.GetSimulateDuration(req.Type) // Simulated time

	s.log.Debug(
		ctx,
		types.ActionOrderProcessingStarted,
		"kitchen worker started proccessing order",
		"worker-name", s.worker.name,
		"order-number", req.Number,
		"cooking-time", cookingTime)

	// Set status cooking
	oldStatus, err := s.orderRepo.SetStatus(ctx, req.Number, s.worker.name, types.StatusOrderCooking, "")
	if err != nil {
		s.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to set cooking status for order", err, "worker-name", s.worker.name)
		return fmt.Errorf("failed to set cooking status for order : %w", err)
	}

	timestamp := time.Now()
	completion := timestamp.Add(cookingTime)

	// RequestID
	var requestID string
	if reqID, ok := ctx.Value(models.GetRequestIDKey()).(string); ok {
		requestID = reqID
	}

	// Publish status update message
	if err := s.producer.StatusUpdate(ctx, &models.StatusUpdate{
		OrderNumber: req.Number,
		OldStatus:   oldStatus,
		NewStatus:   types.StatusOrderCooking,
		ChangedBy:   s.worker.name,
		Timestamp:   timestamp,
		Completion:  completion,
		RequestID:   requestID,
	}); err != nil {
		s.log.Error(ctx, types.ActionRabbitMQPublishFailed, "failed to publish status update", err)
		s.log.Warn(ctx, types.ActionMessageProcessingFailed, "order status changed to cooking, but could not increment number of proccessed order for worker in the database")
	}

	// Simulating working proccess.
	time.Sleep(cookingTime)

	// Set status ready
	oldStatus, err = s.orderRepo.SetStatus(ctx, req.Number, s.worker.name, types.StatusOrderReady, "")
	if err != nil {
		s.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to set ready status for order", err, "worker-name", s.worker.name)
		return fmt.Errorf("failed to set ready status for order: %w", err)
	}

	// Publish status update message
	timestamp = time.Now()
	if err := s.producer.StatusUpdate(ctx, &models.StatusUpdate{
		OrderNumber: req.Number,
		OldStatus:   oldStatus,
		NewStatus:   types.StatusOrderReady,
		ChangedBy:   s.worker.name,
		Timestamp:   timestamp,
		Completion:  completion,
		RequestID:   requestID,
	}); err != nil {
		s.log.Error(ctx, types.ActionRabbitMQPublishFailed, "failed to publish status update", err, "worker-name", s.worker.name)
		s.log.Warn(ctx, types.ActionRabbitMQPublishFailed, "order has been proccessed, but failed to publish status update")
	}

	// Increment number of proccessed orders by the worker.
	if err := s.workerRepo.IncrOrdersProcessed(ctx, s.worker.name); err != nil {
		s.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to increment number of ordered", err, "worker-name", s.worker.name)
		s.log.Warn(ctx, types.ActionMessageProcessingFailed, "order has been proccessed, but could not increment number of proccessed order for worker in the database")
	}

	s.log.Debug(ctx, types.ActionOrderCompleted, "order proccess finished", "worker-name", s.worker.name)
	return nil
}

// heartbeatLoop tries to update last seen field in database each heartbeat interval.
func (s *KitchenWorker) heartbeatLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.log.Info(ctx, "worker_hearbeat_stop", "stopped hearbeat loop")
			return
		case <-ticker.C:
			if err := s.workerRepo.UpdateLastSeen(ctx, s.worker.name); err != nil {
				s.log.Error(ctx, types.ActionDBQueryFailed, "failed to update last seen on worker", err, "worker-name", s.worker.name)
				continue
			}
			s.log.Debug(ctx, types.ActionHeartbeatSent, "heartbeat was sent", "worker-name", s.worker.name)
		}
	}
}
