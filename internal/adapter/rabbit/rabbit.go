package rabbit

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
	"github.com/rabbitmq/amqp091-go"
)

type Client interface {
	Close() error
	Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
	DeclareQueue(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	Publish(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp091.Publishing) error
}

type RabbitClient struct {
	client Client
}

func NewClient(ctx context.Context, cfg rabbit.Config) (*RabbitClient, error) {
	client, err := rabbit.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &RabbitClient{
		client: client,
	}, nil
}

func (r *RabbitClient) Close() error {
	return r.client.Close()
}
