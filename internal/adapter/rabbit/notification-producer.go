package rabbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrEmptyExchangeName = errors.New("empty exchange name")

// notificationProducer
type notificationProducer struct {
	client *rabbit.RabbitMQ

	exchangeName string
	log          logger.Logger
}

func NewProducerNotify(client *rabbit.RabbitMQ, exchange string, log logger.Logger) (*notificationProducer, error) {
	if len(exchange) == 0 {
		return nil, ErrEmptyExchangeName
	}

	// declaring notification exchange
	if err := client.Channel.ExchangeDeclare(
		exchange,
		"fanout",
		true, false, false, false, nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
	}

	return &notificationProducer{
		client:       client,
		exchangeName: exchange,
		log:          log,
	}, nil
}

// StatusUpdate publishes event about status change.
func (p *notificationProducer) StatusUpdate(ctx context.Context, req *models.StatusUpdate) error {
	// Marshal the struct to JSON
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal StatusUpdate: %w", err)
	}

	// Prepare the message
	msg := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent, // Persistent message (2). Means rabbitmq will store message in the disc.
		Timestamp:    time.Now(),
	}

	// Publish to the exchange with empty routing key (fanout ignores it)
	if err := p.client.Channel.PublishWithContext(
		ctx,
		p.exchangeName,
		"",    // routing key is ignored for fanout
		false, // mandatory
		false, // immediate
		msg,
	); err != nil {
		return fmt.Errorf("failed to publish StatusUpdate: %w", err)
	}

	return nil
}
