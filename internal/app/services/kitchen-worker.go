package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/internal/service/kitchen"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	postgresclient "github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
)

var (
	ErrEmptyOrderTypes    = fmt.Errorf("order types cannot be empty")
	ErrDuplicateOrderType = fmt.Errorf("duplicate order type found")
	ErrInvalidOrderType   = errors.New("invalid order type. must be Comma-separated list of order types the worker can handle (e.g., dine_in,takeout)")

	ErrInvalidHeartbeatInterval = errors.New("heartbeat interval must be at least 5 secons")
)

type KitchenWorker interface {
	Work(ctx context.Context, errCh chan<- error)
}

type KitchenService struct {
	postgresDB    *postgresclient.PostgreDB
	kitchenWorker KitchenWorker

	cfg config.Config
	log logger.Logger
}

func NewKitchen(ctx context.Context, cfg config.Config, log logger.Logger) (*KitchenService, error) {
	// Validating worker name
	if err := validateWorkerName(cfg.Services.Kitchen.WorkerName); err != nil {
		log.Error(ctx, types.ActionValidationFailed, "failed to vailidate worker name", err)
		return nil, fmt.Errorf("failed to validate worker name: %w", err)
	}
	// Validating order-types to handle by worker.
	validOrderTypes, err := ValidateOrderTypes(cfg.Services.Kitchen.OrderTypes)
	if err != nil {
		log.Error(ctx, types.ActionValidationFailed, "failed to vailidate provided order types", err)
		return nil, fmt.Errorf("failed to vailidate provided order types: %w", err)
	}

	// validate heartbeat interval
	var heartbeatDuration = time.Duration(cfg.Services.Kitchen.HeartbeatInterval) * time.Second
	if heartbeatDuration <= time.Second*5 {
		return nil, ErrInvalidHeartbeatInterval
	}

	// Postgres database connection
	db, err := postgresclient.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error(ctx, "db_connect", "failed to connect postgres", err)
		return nil, fmt.Errorf("failed to connect postgres: %v", err)
	}
	log.Info(ctx, types.ActionDBConnected, "connected to the database")

	// Initialize consumer
	consumer, err := rabbit.NewOrderConsumer(ctx, cfg.RabbitMQ, cfg.Services.Kitchen.Prefetch, validOrderTypes, log)
	if err != nil {
		log.Error(ctx, "order_consumer_create", "failed to create order consumer", err)
		return nil, fmt.Errorf("failed to create order consumer: %w", err)
	}

	// Initialize repository
	repo := postgres.NewWorkerRepo(db.Pool)

	// Initialize kitchen-worker service
	kitchenWorker := kitchen.NewWorker(repo, consumer, cfg.Services.Kitchen.WorkerName, validOrderTypes, heartbeatDuration, log)

	return &KitchenService{
		postgresDB:    db,
		kitchenWorker: kitchenWorker,

		cfg: cfg,
		log: log,
	}, nil
}

func (s *KitchenService) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	// kitchen worker starts to work in goroutine
	go s.kitchenWorker.Work(ctx, errCh)

	// Waiting signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info(ctx, types.ActionServiceStarted, "service started")

	select {
	case errRun := <-errCh:
		return errRun
	case sig := <-shutdownCh:
		s.log.Info(ctx, types.ActionGracefulShutdown, "shuting down application", "signal", sig.String())

		s.close()
		s.log.Info(ctx, types.ActionGracefulShutdown, "graceful shutdown completed!")
	}

	return nil
}

func (s *KitchenService) close() {
	s.postgresDB.Pool.Close()
}

// ValidateOrderTypes handles all validation cases for the --order-types flag
// Input examples: "", "dine_in", "dine_in,takeout", "dine_in,takeout,dine_in" (invalid)
func ValidateOrderTypes(input string) ([]string, error) {
	// Handle empty input (means all types)
	if input == "" {
		return types.AllOrderTypes, nil
	}

	// Split and clean input
	rawTypes := strings.Split(input, ",")
	orderTypes := make([]string, 0, len(rawTypes))
	seen := make(map[string]struct{})

	for _, rawType := range rawTypes {
		trimmed := strings.TrimSpace(rawType)
		if trimmed == "" {
			continue // skip empty entries
		}

		// Check for duplicates
		if _, exists := seen[trimmed]; exists {
			return nil, fmt.Errorf("%q: %w", trimmed, ErrDuplicateOrderType)
		}
		seen[trimmed] = struct{}{}

		// Validate against available types
		if !types.IsValidOrderType(trimmed) {
			return nil, fmt.Errorf("%q: %w (available: %v)", trimmed, ErrInvalidOrderType, types.AllOrderTypes)
		}

		orderTypes = append(orderTypes, trimmed)
	}

	// After cleaning, check if we have any types left
	if len(orderTypes) == 0 {
		return nil, ErrEmptyOrderTypes
	}

	return orderTypes, nil
}

func validateWorkerName(name string) error {
	// Check length (1-100 characters)
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return errors.New("worker name length must be between 1-100 characters")
	}
	return nil
}
