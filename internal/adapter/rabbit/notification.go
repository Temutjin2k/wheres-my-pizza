package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
)

type NotificationSubscriber struct {
	reader *rabbit.RabbitMQ
	cfg    config.RabbitMQ

	exchangeName string
	stop         chan struct{}
	log          logger.Logger
}

func NewNotificationSubscriber(client *rabbit.RabbitMQ, cfg config.RabbitMQ, log logger.Logger) *NotificationSubscriber {
	return &NotificationSubscriber{
		reader:       client,
		cfg:          cfg,
		exchangeName: cfg.NotificationsExchange,
		stop:         make(chan struct{}, 1),
		log:          log,
	}
}

func (s *NotificationSubscriber) StartListening(ctx context.Context) (chan models.StatusUpdate, error) {
	queueName, err := s.declareAndBindQueue()
	if err != nil {
		s.log.Error(ctx, "rabbit_init_queue", "failed to declare/bind queue", err)
		return nil, err
	}

	s.log.Info(ctx, "rabbit_queue_ready", fmt.Sprintf("Queue %s bound to exchange %s", queueName, s.exchangeName))

	updateCh := make(chan models.StatusUpdate, 1)

	go s.startConsuming(ctx, queueName, updateCh)

	return updateCh, nil
}

func (s *NotificationSubscriber) declareAndBindQueue() (string, error) {
	if err := s.reader.Channel.ExchangeDeclare(
		s.exchangeName,
		"fanout",
		true, false, false, false, nil,
	); err != nil {
		return "", fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := s.reader.Channel.QueueDeclare(
		"", true, false, false, false, nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := s.reader.Channel.QueueBind(
		q.Name, "", s.exchangeName, false, nil,
	); err != nil {
		return "", fmt.Errorf("failed to bind queue: %w", err)
	}

	return q.Name, nil
}

func (s *NotificationSubscriber) startConsuming(ctx context.Context, queueName string, outCh chan models.StatusUpdate) {
	defer close(outCh)

	for {
		msgs, err := s.reader.Channel.Consume(
			queueName, "", false, false, false, false, nil,
		)
		if err != nil {
			s.log.Error(ctx, "rabbit_consume_start", "failed to consume messages", err)
			return
		}

		s.log.Info(ctx, "rabbit_listening", "Started listening for notifications")

		connClose := make(chan struct{}, 1)
		go s.isAlive(connClose)
	consumeLoop:
		for {
			select {
			case <-connClose:
				s.log.Warn(ctx, "rabbit_channel_closed", "Channel closed by broker, attempting to reconnect", "error", err)

				if err := s.reconnect(ctx); err != nil {
					s.log.Error(ctx, "rabbit_reconnect_failed", "Reconnection failed", err)
					return
				}

				break consumeLoop
			case <-s.stop:
				s.log.Info(ctx, "rabbit_consume_stop", "Stopped listening to notifications")
				return

			case msg := <-msgs:

				update, err := decodeStatusUpdate(msg.Body)
				if err != nil {
					s.log.Error(ctx, "notification_decode", "Failed to decode status update", err)
					_ = msg.Nack(false, false)
					continue
				}

				if err := msg.Ack(false); err != nil {
					s.log.Error(ctx, "rabbit_ack", "Failed to ack message", err)
				}

				outCh <- update
			}
		}
	}
}

func (s *NotificationSubscriber) isAlive(connClose chan struct{}) {
	t := time.NewTicker(time.Second * 5)

	select {
	case <-t.C:
		if s.reader.Conn.IsClosed() {
			connClose <- struct{}{}
			return
		}

		if err := s.reader.Channel.Flow(false); err != nil {
			connClose <- struct{}{}
			return
		}
	case <-s.stop:
		return
	}
}

func (s *NotificationSubscriber) reconnect(ctx context.Context) error {
	var lastErr error

	for attempt := 1; attempt <= 5; attempt++ {
		s.log.Info(ctx, "rabbit_reconnect_attempt", fmt.Sprintf("Attempt %d to reconnect", attempt))

		conn, err := rabbit.New(ctx, s.cfg.Conn)
		if err == nil {
			s.reader = conn
			s.log.Info(ctx, "rabbit_reconnect_success", "Successfully reconnected to RabbitMQ")
			return nil
		}

		lastErr = err
		s.log.Warn(ctx, "rabbit_reconnect_failed", fmt.Sprintf("Attempt %d failed: %v", attempt, err))

		select {
		case <-s.stop:
			return fmt.Errorf("reconnect stopped externally")
		case <-time.After(time.Duration(attempt) * time.Second):
			// Exponential-ish backoff
		}
	}

	return fmt.Errorf("failed to reconnect after 5 attempts: %w", lastErr)
}

func (s *NotificationSubscriber) Close() error {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	return s.reader.Close()
}

func decodeStatusUpdate(body []byte) (models.StatusUpdate, error) {
	var update models.StatusUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return models.StatusUpdate{}, err
	}
	return update, nil
}
