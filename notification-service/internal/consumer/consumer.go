package consumer

import (
	"ap2/notification-service/internal/domain"
	"ap2/notification-service/internal/provider"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "payment.events"
	exchangeType = "direct"
	routingKey   = "payment.completed"
	queueName    = "payment.completed"

	dlxExchange = "payment.events.dlx"
	dlqName     = "payment.completed.dlq"
)

type Consumer struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	rdb      *redis.Client
	provider provider.NotificationProvider

	maxRetries     int
	initialBackoff time.Duration
}

func New(amqpURL, redisAddr string, notifProvider provider.NotificationProvider, maxRetries int, initialBackoff time.Duration) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(dlxExchange, "direct", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare dlx exchange: %w", err)
	}

	_, err = ch.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare dlq: %w", err)
	}

	if err := ch.QueueBind(dlqName, routingKey, dlxExchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("bind dlq: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    dlxExchange,
		"x-dead-letter-routing-key": routingKey,
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(queueName, routingKey, exchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	// Initialize Redis for idempotency.
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Println("RabbitMQ consumer connected")
	log.Printf("Notification provider: %T, maxRetries=%d, initialBackoff=%v", notifProvider, maxRetries, initialBackoff)

	return &Consumer{
		conn:           conn,
		ch:             ch,
		rdb:            rdb,
		provider:       notifProvider,
		maxRetries:     maxRetries,
		initialBackoff: initialBackoff,
	}, nil
}

func (c *Consumer) Start(done <-chan struct{}) error {
	msgs, err := c.ch.Consume(
		queueName,
		"notification-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Println("Notification consumer started, waiting for messages...")

	for {
		select {
		case <-done:
			log.Println("Consumer received shutdown signal")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return nil
			}
			c.handleMessage(msg)
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	var event domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Notification] Failed to parse message: %v", err)
		msg.Nack(false, false)
		return
	}

	ctx := context.Background()

	// Redis-based idempotency check.
	idempotencyKey := fmt.Sprintf("notification:processed:%s", event.PaymentID)
	exists, err := c.rdb.Exists(ctx, idempotencyKey).Result()
	if err != nil {
		log.Printf("[Notification] Redis idempotency check failed: %v", err)
	}
	if exists > 0 {
		log.Printf("[Notification] Duplicate payment_id %s, skipping", event.PaymentID)
		msg.Ack(false)
		return
	}

	// Send notification with exponential backoff retries.
	subject := fmt.Sprintf("Payment %s for Order #%s", event.Status, event.OrderID)
	body := fmt.Sprintf("Dear Customer,\n\nYour payment of $%.2f for order %s has been %s.\n\nThank you.",
		float64(event.Amount)/100.0, event.OrderID, event.Status)

	var sendErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.initialBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
			log.Printf("[Notification] Retry %d/%d for payment_id=%s, backoff=%v",
				attempt, c.maxRetries, event.PaymentID, backoff)
			time.Sleep(backoff)
		}

		sendErr = c.provider.Send(event.CustomerEmail, subject, body)
		if sendErr == nil {
			break
		}
		log.Printf("[Notification] Provider error (attempt %d/%d): %v",
			attempt+1, c.maxRetries+1, sendErr)
	}

	if sendErr != nil {
		log.Printf("[Notification] All retries exhausted for payment_id=%s, sending to DLQ: %v",
			event.PaymentID, sendErr)
		msg.Nack(false, false)
		return
	}

	// Mark as processed in Redis (with 24h TTL to prevent memory leaks).
	if err := c.rdb.Set(ctx, idempotencyKey, "sent", 24*time.Hour).Err(); err != nil {
		log.Printf("[Notification] Failed to set idempotency key: %v", err)
	}

	log.Printf("[Notification] Successfully sent notification for payment_id=%s order=%s to=%s",
		event.PaymentID, event.OrderID, event.CustomerEmail)

	if err := msg.Ack(false); err != nil {
		log.Printf("[Notification] Failed to ACK message: %v", err)
	}
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	if c.rdb != nil {
		c.rdb.Close()
	}
}
