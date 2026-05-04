package consumer

import (
	"ap2/notification-service/internal/domain"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "payment.events"
	exchangeType = "direct"
	routingKey   = "payment.completed"
	queueName    = "payment.completed"

	dlxExchange = "payment.events.dlx"
	dlqName     = "payment.completed.dlq"

	maxRetries = 3
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	mu        sync.RWMutex
	processed map[string]struct{} // in-memory idempotency store
}

func New(amqpURL string) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Declare exchange
	if err := ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	// Declare DLX exchange and DLQ
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

	// Declare main queue with DLX
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

	// Prefetch 1 message at a time for fair dispatch
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Println("RabbitMQ consumer connected")
	return &Consumer{
		conn:      conn,
		ch:        ch,
		processed: make(map[string]struct{}),
	}, nil
}

func (c *Consumer) Start(done <-chan struct{}) error {
	msgs, err := c.ch.Consume(
		queueName,
		"notification-consumer",
		false, // auto-ack disabled (manual ACK)
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
		msg.Nack(false, false) // send to DLQ
		return
	}

	// Idempotency check
	if c.isProcessed(event.EventID) {
		log.Printf("[Notification] Duplicate event %s, skipping", event.EventID)
		msg.Ack(false)
		return
	}

	// Check retry count; reject to DLQ after max retries
	retryCount := getRetryCount(msg)
	if retryCount >= maxRetries {
		log.Printf("[Notification] Message %s exceeded max retries (%d), sending to DLQ",
			event.EventID, maxRetries)
		msg.Nack(false, false)
		return
	}

	// Simulate sending notification (log to console as per assignment)
	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f. Status: %s",
		event.CustomerEmail, event.OrderID, float64(event.Amount)/100.0, event.Status)

	// Mark as processed for idempotency
	c.markProcessed(event.EventID)

	// Manual ACK after successful processing
	if err := msg.Ack(false); err != nil {
		log.Printf("[Notification] Failed to ACK message: %v", err)
	}
}

func (c *Consumer) isProcessed(eventID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.processed[eventID]
	return exists
}

func (c *Consumer) markProcessed(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processed[eventID] = struct{}{}
}

func getRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	deaths, ok := msg.Headers["x-death"]
	if !ok {
		return 0
	}
	deathList, ok := deaths.([]interface{})
	if !ok || len(deathList) == 0 {
		return 0
	}
	for _, d := range deathList {
		death, ok := d.(amqp.Table)
		if !ok {
			continue
		}
		if count, ok := death["count"]; ok {
			switch v := count.(type) {
			case int64:
				return int(v)
			case int32:
				return int(v)
			}
		}
	}
	return 0
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
