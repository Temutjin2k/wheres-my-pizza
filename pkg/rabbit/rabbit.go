package rabbit

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

type Config struct {
	Host     string `env:"RABBITMQ_HOST" default:"localhost"`
	Port     string `env:"RABBITMQ_PORT" default:"5672"`
	User     string `env:"RABBITMQ_USER" default:"guest"`
	Password string `env:"RABBITMQ_PASSWORD" default:"guest"`
}

func (c Config) GetDSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.User,
		c.Password,
		c.Host,
		c.Port,
	)
}

func New(ctx context.Context, config Config) (*RabbitMQ, error) {
	conn, err := amqp.DialConfig(config.GetDSN(), amqp.Config{
		Heartbeat: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close() // Close connection if channel creation fails
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Verify the connection is alive
	select {
	case <-conn.NotifyClose(make(chan *amqp.Error)):
		return nil, fmt.Errorf("rabbitmq connection is closed")
	default:
		// Connection is good
	}

	return &RabbitMQ{
		Conn:    conn,
		Channel: channel,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.Conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
