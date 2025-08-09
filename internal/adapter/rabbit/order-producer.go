package rabbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
	"github.com/rabbitmq/amqp091-go"
)

type OrderProducer struct {
	client *rabbit.RabbitMQ

	exchangeOrder string
	log           logger.Logger
}

func NewOrderProducer(ctx context.Context, cfg config.RabbitMQ, log logger.Logger) (*OrderProducer, error) {
	client, err := rabbit.New(ctx, cfg.Conn)
	if err != nil {
		return nil, err
	}

	// Creating exchange.
	if err := client.Channel.ExchangeDeclare(
		cfg.OrderExchange,
		"topic", // exchange type of 'topic'
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// creating all possible queues and binding them to order exchange.
	if err := InitQueuesForOrderTypes(client, cfg.OrderExchange, types.AllOrderTypes); err != nil {
		return nil, err
	}

	// logging connection
	log.Info(ctx, types.ActionRabbitMQConnected, "connected to rabbitMQ")

	return &OrderProducer{
		client:        client,
		exchangeOrder: cfg.OrderExchange,
		log:           log,
	}, err
}

func (r *OrderProducer) Close() error {
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
func (r *OrderProducer) PublishCreateOrder(ctx context.Context, order *models.CreateOrder) error {
	if order == nil {
		return errors.New("nil order")
	}

	// Marshal order to JSON
	body, err := json.Marshal(FromInternalToPublishOrder(ctx, order))
	if err != nil {
		r.log.Error(ctx, types.ActionValidationFailed, "failed to marshal", err)
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	routingKey := createOrderPublishedKey(order)

	// Create the message with persistent delivery mode
	msg := amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent, // Persistent message (2). Means rabbitmq will store message in the disc.
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
		r.log.Error(ctx, types.ActionRabbitMQPublishFailed, "failed to publish order", err)
		return fmt.Errorf("failed to publish order: %w", err)
	}

	return nil
}
