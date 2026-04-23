package usecase

import (
	"ap2/payment-service/internal/domain"
	"time"
)

type PaymentRepository interface {
	Save(payment *domain.Payment) error
	FindByOrderID(orderID string) (*domain.Payment, error)
}

type AuthorizeRequest struct {
	OrderID string
	Amount  int64
}

type AuthorizeResponse struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}
