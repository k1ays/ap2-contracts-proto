package consumer

import (
	"ap2/notification-service/internal/domain"
	"ap2/notification-service/internal/provider"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
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
	workerCount    int
}

func New(amqpURL, redisAddr string, notifProvider provider.NotificationProvider, maxRetries int, initialBackoff time.Duration, workerCount int) (*Consumer, error) {
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

	if workerCount < 1 {
		workerCount = 1
	}
	// Prefetch == workerCount so each worker can hold one unacked message.
	if err := ch.Qos(workerCount, 0, false); err != nil {
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
	log.Printf("Notification provider: %T, maxRetries=%d, initialBackoff=%v, workerCount=%d",
		notifProvider, maxRetries, initialBackoff, workerCount)

	return &Consumer{
		conn:           conn,
		ch:             ch,
		rdb:            rdb,
		provider:       notifProvider,
		maxRetries:     maxRetries,
		initialBackoff: initialBackoff,
		workerCount:    workerCount,
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

	log.Printf("Notification consumer started with %d workers, waiting for messages...", c.workerCount)

	var wg sync.WaitGroup
	for i := 1; i <= c.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[Worker %d] started", workerID)
			for {
				select {
				case <-done:
					log.Printf("[Worker %d] shutdown signal received", workerID)
					return
				case msg, ok := <-msgs:
					if !ok {
						log.Printf("[Worker %d] message channel closed", workerID)
						return
					}
					c.handleMessage(workerID, msg)
				}
			}
		}(i)
	}

	wg.Wait()
	log.Println("All notification workers stopped")
	return nil
}

func (c *Consumer) handleMessage(workerID int, msg amqp.Delivery) {
	var event domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Worker %d] [Notification] Failed to parse message: %v", workerID, err)
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
		log.Printf("[Worker %d] [Notification] Duplicate payment_id %s, skipping", workerID, event.PaymentID)
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
			log.Printf("[Worker %d] [Notification] Retry %d/%d for payment_id=%s, backoff=%v",
				workerID, attempt, c.maxRetries, event.PaymentID, backoff)
			time.Sleep(backoff)
		}

		sendErr = c.provider.Send(event.CustomerEmail, subject, body)
		if sendErr == nil {
			break
		}
		log.Printf("[Worker %d] [Notification] Provider error (attempt %d/%d): %v",
			workerID, attempt+1, c.maxRetries+1, sendErr)
	}

	if sendErr != nil {
		log.Printf("[Worker %d] [Notification] All retries exhausted for payment_id=%s, sending to DLQ: %v",
			workerID, event.PaymentID, sendErr)
		msg.Nack(false, false)
		return
	}

	// Mark as processed in Redis (with 24h TTL to prevent memory leaks).
	if err := c.rdb.Set(ctx, idempotencyKey, "sent", 24*time.Hour).Err(); err != nil {
		log.Printf("[Worker %d] [Notification] Failed to set idempotency key: %v", workerID, err)
	}

	log.Printf("[Worker %d] [Notification] Successfully sent notification for payment_id=%s order=%s to=%s",
		workerID, event.PaymentID, event.OrderID, event.CustomerEmail)

	if err := msg.Ack(false); err != nil {
		log.Printf("[Worker %d] [Notification] Failed to ACK message: %v", workerID, err)
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
