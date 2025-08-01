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
	Host     string `yaml:"host" env:"RABBITMQ_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"RABBITMQ_PORT" env-default:"5672"`
	User     string `yaml:"user" env:"RABBITMQ_USER" env-default:"guest"`
	Password string `yaml:"password" env:"RABBITMQ_PASSWORD" env-default:"guest"`
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
		Locale:    "en_US",
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

// DeclareQueue is a helper to declare a queue
func (r *RabbitMQ) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,
		durable,
		autoDelete,
		exclusive,
		noWait,
		args,
	)
}

// Publish is a helper to publish messages
func (r *RabbitMQ) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return r.Channel.PublishWithContext(
		ctx,
		exchange,
		key,
		mandatory,
		immediate,
		msg,
	)
}

// Consume is a helper to consume messages
func (r *RabbitMQ) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return r.Channel.Consume(
		queue,
		consumer,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
}
