package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	queueName    string
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
	s.log.Info(ctx, "subscriber_start", "Starting to listen for notifications")

	if err := s.declareAndBindQueue(); err != nil {
		s.log.Error(ctx, "rabbit_init_queue", "Failed to declare/bind queue", err)
		return nil, err
	}

	s.log.Info(ctx, "rabbit_queue_ready", fmt.Sprintf("Queue %s bound to exchange %s", s.queueName, s.exchangeName))

	updateCh := make(chan models.StatusUpdate, 1)

	go s.startConsuming(ctx, updateCh)

	return updateCh, nil
}

func (s *NotificationSubscriber) declareAndBindQueue() error {
	if err := s.reader.Channel.ExchangeDeclare(
		s.exchangeName,
		"fanout",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := s.reader.Channel.QueueDeclare(
		"", true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := s.reader.Channel.QueueBind(
		q.Name, "", s.exchangeName, false, nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	s.queueName = q.Name

	return nil
}

func (s *NotificationSubscriber) startConsuming(ctx context.Context, outCh chan models.StatusUpdate) {
	defer close(outCh)

	for {
		msgs, err := s.reader.Channel.Consume(
			s.queueName, "", false, false, false, false, nil,
		)
		if err != nil {
			s.log.Error(ctx, "rabbit_consume_start", "failed to consume messages", err)
			return
		}

		s.log.Info(ctx, "rabbit_listening", "Started listening for notifications")

		connClose := make(chan struct{}, 1)
		go s.isAlive(ctx, connClose)

	consumeLoop:
		for {
			select {
			case msg := <-msgs:
				update, err := decodeStatusUpdate(msg.Body)
				if len(update.RequestID) != 0 {
					ctx = logger.WithRequestID(ctx, update.RequestID) // request_id logging
				}

				if err != nil {
					s.log.Error(ctx, "notification_decode", "Failed to decode status update", err)
					if err := msg.Nack(false, false); err != nil {
						s.log.Error(ctx, "rabbit_ack", "Failed to ack message", err)
					}
					continue
				}

				if err := msg.Ack(false); err != nil {
					s.log.Error(ctx, "rabbit_ack", "Failed to ack message", err)
				}

				outCh <- update
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
			}
		}
	}
}

func (s *NotificationSubscriber) isAlive(ctx context.Context, connClose chan struct{}) {
	t := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-t.C:
			if s.reader.Conn.IsClosed() {
				s.log.Warn(ctx, "rabbit_connection_dead", "Detected closed connection")
				connClose <- struct{}{}
				return
			}
		case <-s.stop:
			s.log.Debug(ctx, "is_alive_stop", "Stopped connection health check")
			return
		}
	}
}

func (s *NotificationSubscriber) reconnect(ctx context.Context) error {
	var lastErr error

	for attempt := 1; attempt <= 5; attempt++ {
		s.log.Info(ctx, "rabbit_reconnect_attempt", fmt.Sprintf("Attempt %d to reconnect", attempt))

		conn, err := rabbit.New(ctx, s.cfg.Conn, s.log)
		if err == nil {
			s.reader = conn
			s.log.Info(ctx, "rabbit_reconnect_success", "Successfully reconnected to RabbitMQ")
			return nil
		}

		lastErr = err
		s.log.Warn(ctx, "rabbit_reconnect_failed", fmt.Sprintf("Attempt %d failed", attempt), err)

		select {
		case <-s.stop:
			s.log.Info(ctx, "rabbit_reconnect_stopped", "Reconnect was stopped externally")
			return fmt.Errorf("reconnect stopped externally")
		case <-time.After(time.Duration(attempt) * time.Second):
			time.Sleep(2 * time.Second)
		}
	}

	s.log.Error(ctx, "rabbit_reconnect_max_failed", "Failed to reconnect after max attempts", lastErr)
	return fmt.Errorf("failed to reconnect after 5 attempts: %w", lastErr)
}

func (s *NotificationSubscriber) Close() error {
	select {
	case s.stop <- struct{}{}:
	default:
	}

	if _, err := s.reader.Channel.QueueDelete(s.queueName, false, false, true); err != nil {
		s.log.Warn(context.Background(), "rabbitMQ_closing", "failed to close queue", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.reader.Close(ctx)
}

func decodeStatusUpdate(body []byte) (models.StatusUpdate, error) {
	var update models.StatusUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return models.StatusUpdate{}, err
	}
	return update, nil
}

func (s *NotificationSubscriber) GetListenerCount() (int, error) {
	vhost := url.PathEscape("/")
	exchange := url.PathEscape(s.exchangeName)

	url := fmt.Sprintf(
		"http://%s:%s@%s:15672/api/exchanges/%s/%s/bindings/source",
		s.cfg.Conn.User,
		s.cfg.Conn.Password,
		s.cfg.Conn.Host,
		vhost,
		exchange,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	var bindings []struct {
		Destination     string `json:"destination"`
		DestinationType string `json:"destination_type"`
	}
	if err := json.Unmarshal(body, &bindings); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	// Фильтруем только очереди
	count := 0
	for _, b := range bindings {
		if b.DestinationType == "queue" {
			count++
		}
	}

	return count, nil
}
