package usecase

import (
	"ap2/order-service/internal/domain"
	"context"
)

// OrderRepository is the port for persistence.
type OrderRepository interface {
	Save(order *domain.Order) error
	FindByID(id string) (*domain.Order, error)
	Update(order *domain.Order) error
	FindByStatus(status string) ([]*domain.Order, error)
}

// PaymentClient is the port for outbound payment authorization.
type PaymentClient interface {
	Authorize(ctx context.Context, orderID string, amount int64) (string, string, error)
}

// OrderCache is the port for caching order data.
type OrderCache interface {
	Get(ctx context.Context, id string) (*domain.Order, error)
	Set(ctx context.Context, order *domain.Order) error
	Invalidate(ctx context.Context, id string) error
}
