package domain

import (
	"errors"
	"time"
)

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"

	MaxPaymentAmount int64 = 100000 // 1000 units in cents
)

var (
	ErrAmountExceedsLimit = errors.New("amount exceeds payment limit")
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
	ErrPaymentNotFound    = errors.New("payment not found")
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}

func NewPayment(orderID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if amount > MaxPaymentAmount {
		return nil, ErrAmountExceedsLimit
	}
	return &Payment{
		OrderID:   orderID,
		Amount:    amount,
		Status:    StatusAuthorized,
		CreatedAt: time.Now(),
	}, nil
}
