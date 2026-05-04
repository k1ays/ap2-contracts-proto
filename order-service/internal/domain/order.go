package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrCannotCancelPaid  = errors.New("paid orders cannot be cancelled")
	ErrCannotCancelOrder = errors.New("only pending orders can be cancelled")
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64 // in cents
	Status     string
	CreatedAt  time.Time
}

func NewOrder(customerID, itemName string, amount int64) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}, nil
}

func (o *Order) MarkPaid() {
	o.Status = StatusPaid
}

func (o *Order) MarkFailed() {
	o.Status = StatusFailed
}

func (o *Order) Cancel() error {
	if o.Status == StatusPaid {
		return ErrCannotCancelPaid
	}
	if o.Status != StatusPending {
		return ErrCannotCancelOrder
	}
	o.Status = StatusCancelled
	return nil
}
