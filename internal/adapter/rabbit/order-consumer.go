package rabbit

import (
	"context"
	"errors"
	"fmt"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
)

type OrderConsumer struct {
	client *rabbit.RabbitMQ

	prefetchCount int
	exchangeOrder string
	orderTypes    []string

	log logger.Logger
}

func NewOrderConsumer(ctx context.Context, exchange string, client *rabbit.RabbitMQ, prefetchCount int, orderTypes []string, log logger.Logger) (*OrderConsumer, error) {
	if len(orderTypes) == 0 {
		return nil, errors.New("orderTypes not provided, slice len 0")
	}

	// Creating exchange.
	if err := client.Channel.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	if err := InitQueuesForOrderTypes(client, exchange, orderTypes); err != nil {
		return nil, err
	}

	return &OrderConsumer{
		client:        client,
		prefetchCount: prefetchCount,
		exchangeOrder: exchange,
		orderTypes:    orderTypes,

		log: log,
	}, nil
}

// Consumes consumes created order messages.
func (c *OrderConsumer) Consume(
	ctx context.Context,
	orderType string,
	handler func(ctx context.Context, req *models.CreateOrder) error,
) error {
	// In RabbitMQ, basic.qos is a method used to configure the quality of service for consumers,
	// specifically by controlling how many messages a consumer can receive without acknowledging them.
	// Using basic.qos is critical to prevent worker overload and distribute the load evenly between workers.
	if err := c.client.Channel.Qos(c.prefetchCount, 0, false); err != nil {
		c.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to set QoS", err, "prefetchCount", c.prefetchCount)
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	queueName := getQueueByOrderType(orderType)

	msgs, err := c.client.Channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to consume queue", err, "queue", queueName)
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			req, err := ToInternalOrder(msg.Body)
			if err != nil {
				msg.Nack(false, false)
				c.log.Error(ctx, types.ActionValidationFailed, "failed to validate message", err)
				continue
			}

			// request_id logging
			if len(req.RequestID) != 0 {
				ctx = logger.WithRequestID(ctx, req.RequestID)
			}

			order := FromPublishToInternalOrder(req)
			if order == nil {
				msg.Nack(false, false)
				c.log.Error(ctx, types.ActionValidationFailed, "failed to validate message", err)
				continue
			}

			if err := handler(ctx, order); err != nil {
				if isRecoverableError(err) {
					msg.Nack(false, true) // Requeue
				} else {
					msg.Nack(false, false) // Sending to DLQ
				}

				c.log.Error(ctx, types.ActionMessageProcessingFailed, "failed to handle message", err, "requeue", isRecoverableError(err))
				continue
			}
			msg.Ack(false)
		}
	}
}

// - If any processing step fails (e.g., database unavailable), the worker must negatively acknowledge the
// message (`basic.nack`) with `requeue=true` so it can be re-processed later.
// - If there are data validation errors in the message that will never be corrected, the message should be
// sent to a Dead-Letter Queue (`DLQ`) to allow for manual analysis and to prevent the queue from being blocked.

// isRecoverableError returns true if the provided error must be requeued
func isRecoverableError(_ error) bool {
	return true
}
