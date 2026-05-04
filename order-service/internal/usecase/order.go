package usecase

import (
	"ap2/order-service/internal/domain"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo    OrderRepository
	payment PaymentClient
}

func NewOrderUseCase(repo OrderRepository, payment PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, payment: payment}
}

// CreateOrder creates a Pending order, calls Payment Service, then updates status.
func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64) (*domain.Order, error) {
	order, err := domain.NewOrder(customerID, itemName, amount)
	if err != nil {
		return nil, err
	}
	order.ID = generateID()

	// Persist as Pending first.
	if err := uc.repo.Save(order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	// Call Payment Service (timeout is enforced by the HTTP client).
	_, status, err := uc.payment.Authorize(ctx, order.ID, order.Amount)
	if err != nil {
		// Payment Service unavailable — mark as Failed, return 503 to caller.
		order.MarkFailed()
		_ = uc.repo.Update(order)
		return nil, ErrPaymentServiceUnavailable
	}

	if status == "Authorized" {
		order.MarkPaid()
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("update order status: %w", err)
	}
	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.FindByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := order.Cancel(); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}
	return order, nil
}

func (uc *OrderUseCase) ListOrders(status string) ([]*domain.Order, error) {
	return uc.repo.FindByStatus(status)
}

// Sentinel error returned when Payment Service is unreachable.
var ErrPaymentServiceUnavailable = fmt.Errorf("payment service unavailable")

func generateID() string {
	return uuid.NewString()
}

func generateTransactionID() string {
	return "txn_" + uuid.NewString()
}
