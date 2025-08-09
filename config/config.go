package config

import (
	"errors"
	"flag"
	"os"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/configparser"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
)

const (
	DefaultOrderServicePort    = 3000
	DefaultTrackingServicePort = 3002
)

var (
	// General
	modeFlag = flag.String("mode", "", "application mode")
	helpFlag = flag.Bool("help", false, "Show help message")
	portFlag = flag.Int("port", -1, "The HTTP port for the API")

	//Order service
	maxConcurrent = flag.Int("max-concurrent", 50, "Maximum number of concurrent orders to process.")

	// Kitchen service
	workerName   = flag.String("worker-name", "", "unique name for the worker (e.g., chef_mario) (required)")
	orderTypes   = flag.String("order-types", "all", "comma-separated list of order types the worker can handle (e.g., dine_in,takeout)")
	heartbeatInt = flag.Int("heartbeat-interval", 30, "interval (seconds) between heartbeats")
	prefetch     = flag.Int("prefetch", 1, "RabbitMQ prefetch count")
)

var (
	ErrModeNotProvided = errors.New("mode flag not provided")
	ErrInvalidModeFlag = errors.New("invalid mode flag")
)

type (
	// Config
	Config struct {
		Flags    Flags
		Services Services
		Server   Server
		Postgres postgres.Config
		RabbitMQ RabbitMQ
	}

	Services struct {
		Order    OrderService
		Kitchen  KitchenService
		Tracking TrackingService
	}

	// Servers config
	Server struct {
		HTTPServer HTTPServer
	}

	// HTTP service
	HTTPServer struct {
		Port int
	}

	Flags struct {
		Help bool
		Mode types.ServiceMode
	}

	OrderService struct {
		MaxConcurrent int
	}

	TrackingService struct {
		HeartbeatInterval int
	}

	KitchenService struct {
		WorkerName        string
		OrderType         string
		Prefetch          int
		HeartbeatInterval int
	}

	RabbitMQ struct {
		Conn                  rabbit.Config
		OrderExchange         string `env:"RABBITMQ_ORDER_EXCHANGE" default:"orders_topic"`
		OrderQueue            string `env:"RABBITMQ_ORDER_QUEUE" default:"orders_topic"`
		NotificationsExchange string `env:"RABBITMQ_NOTIFICATIONS_EXCHANGE" default:"notifications_fanout"`
	}
)

func New(filepath string) (*Config, error) {
	cfg := &Config{}

	if err := configparser.LoadAndParseYaml(filepath, cfg); err != nil {
		return nil, err
	}

	if err := parseFlags(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseFlags(cfg *Config) error {
	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if modeFlag == nil {
		return ErrModeNotProvided
	}

	cfg.Flags.Mode = types.ServiceMode(*modeFlag)
	cfg.Flags.Help = *helpFlag

	switch cfg.Flags.Mode {
	case types.ModeOrder:
		if portFlag == nil || *portFlag < 1024 || *portFlag > 65535 {
			cfg.Server.HTTPServer.Port = DefaultOrderServicePort
		} else {
			cfg.Server.HTTPServer.Port = *portFlag
		}

		if maxConcurrent != nil && !(*maxConcurrent >= 0 && *maxConcurrent <= 1000) {
			return errors.New("--max-concurrent flag must be between 1 and 1000")
		}
		cfg.Services.Order.MaxConcurrent = *maxConcurrent
	case types.ModeKitchenWorker:
		if workerName == nil || *workerName == "" {
			return errors.New("missing required flag: --worker-name")
		}

		cfg.Services.Kitchen.WorkerName = *workerName
		cfg.Services.Kitchen.OrderType = *orderTypes
		cfg.Services.Kitchen.HeartbeatInterval = *heartbeatInt
		cfg.Services.Kitchen.Prefetch = *prefetch
	case types.ModeTracking:
		if portFlag != nil {
			cfg.Server.HTTPServer.Port = *portFlag
		} else {
			cfg.Server.HTTPServer.Port = DefaultTrackingServicePort
		}
		cfg.Services.Tracking.HeartbeatInterval = *heartbeatInt
	case types.ModeNotificationSubscriber:
	default:
		return ErrInvalidModeFlag
	}

	return nil
}
