package rabbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	client *rabbit.RabbitMQ

	exchangeOrder string
	kitchenQueue  string
}

func NewClient(ctx context.Context, cfg config.RabbitMQ) (*RabbitClient, error) {
	client, err := rabbit.New(ctx, cfg.Conn)
	if err != nil {
		return nil, err
	}

	rabbitClient := &RabbitClient{
		client: client,

		exchangeOrder: cfg.OrderExchange,
		kitchenQueue:  cfg.OrderQueue,
	}

	rabbitClient.InitKitchenQueue()

	return rabbitClient, err
}

func (r *RabbitClient) InitKitchenQueue() error {
	const bindingKey = "kitchen.*.*"

	// 1. Declare the topic exchange (idempotent)
	err := r.client.Channel.ExchangeDeclare(
		r.exchangeOrder, // "orders_topic"
		"topic",         // type
		true,            // durable
		false,           // autoDelete
		false,           // internal
		false,           // noWait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 2. Declare the durable queue
	_, err = r.client.Channel.QueueDeclare(
		r.kitchenQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 3. Bind the queue to the exchange with a wildcard routing key
	err = r.client.Channel.QueueBind(
		r.kitchenQueue,
		bindingKey,
		r.exchangeOrder,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

func (r *RabbitClient) Close() error {
	if r.client == nil {
		return nil
	}
	if err := r.client.Close(); err != nil {
		return err
	}
	r.client = nil
	return nil
}

// PublishCreateOrder publishes an order message to the orders_topic exchange
func (r *RabbitClient) PublishCreateOrder(ctx context.Context, order *models.CreateOrder) error {
	if order == nil {
		return errors.New("nil order")
	}

	// Ensure the exchange exists (declare if not)
	err := r.client.Channel.ExchangeDeclare(
		r.exchangeOrder, // exchange name
		"topic",         // type (topic exchange)
		true,            // durable
		false,           // auto-delete
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Marshal order to JSON
	body, err := json.Marshal(FromInternalToPublishOrder(order))
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	routingKey := createOrderPublishedKey(order)

	// Create the message with persistent delivery mode
	msg := amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent, // Persistent message (2)
		Priority:     uint8(order.Priority),
		Timestamp:    time.Now(),
		Body:         body,
	}

	// Publish to the orders_topic exchange
	err = r.client.Channel.PublishWithContext(
		ctx,
		r.exchangeOrder, // exchange name
		routingKey,
		true,  // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish order: %w", err)
	}

	return nil
}

// Create the routing key
func createOrderPublishedKey(order *models.CreateOrder) string {
	return fmt.Sprintf("kitchen.%s.%d", order.Type, order.Priority)
}
