package usecase

import (
	"ap2/payment-service/internal/domain"
	"time"
)

type PaymentRepository interface {
	Save(payment *domain.Payment) error
	FindByOrderID(orderID string) (*domain.Payment, error)
}

type EventPublisher interface {
	PublishPaymentCompleted(event PaymentCompletedEvent) error
	Close() error
}

type AuthorizeRequest struct {
	OrderID       string
	Amount        int64
	CustomerEmail string
}

type AuthorizeResponse struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}
