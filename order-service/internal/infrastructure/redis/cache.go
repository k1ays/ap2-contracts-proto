package redis

import (
	"ap2/order-service/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewOrderCache(addr string, ttl time.Duration) (*OrderCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Printf("Redis connected at %s (TTL=%v)", addr, ttl)
	return &OrderCache{client: client, ttl: ttl}, nil
}

func (c *OrderCache) cacheKey(id string) string {
	return "order:" + id
}

func (c *OrderCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	data, err := c.client.Get(ctx, c.cacheKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("unmarshal cached order: %w", err)
	}
	return &order, nil
}

func (c *OrderCache) Set(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}

	if err := c.client.Set(ctx, c.cacheKey(order.ID), data, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *OrderCache) Invalidate(ctx context.Context, id string) error {
	if err := c.client.Del(ctx, c.cacheKey(id)).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (c *OrderCache) Close() error {
	return c.client.Close()
}
