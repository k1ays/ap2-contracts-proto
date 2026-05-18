package usecase

import (
	"ap2/order-service/internal/domain"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo    OrderRepository
	payment PaymentClient
	cache   OrderCache
}

func NewOrderUseCase(repo OrderRepository, payment PaymentClient, cache OrderCache) *OrderUseCase {
	return &OrderUseCase{repo: repo, payment: payment, cache: cache}
}

// CreateOrder creates a Pending order, calls Payment Service, then updates status.
func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64) (*domain.Order, error) {
	order, err := domain.NewOrder(customerID, itemName, amount)
	if err != nil {
		return nil, err
	}
	order.ID = generateID()

	if err := uc.repo.Save(order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	_, status, err := uc.payment.Authorize(ctx, order.ID, order.Amount)
	if err != nil {
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

	// Invalidate cache after status change.
	if uc.cache != nil {
		if err := uc.cache.Invalidate(ctx, order.ID); err != nil {
			log.Printf("[Cache] failed to invalidate order %s: %v", order.ID, err)
		}
	}

	return order, nil
}

// GetOrder uses cache-aside: check cache first, fallback to DB, then populate cache.
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if uc.cache != nil {
		cached, err := uc.cache.Get(ctx, id)
		if err != nil {
			log.Printf("[Cache] failed to get order %s: %v", id, err)
		}
		if cached != nil {
			log.Printf("[Cache] HIT for order %s", id)
			return cached, nil
		}
		log.Printf("[Cache] MISS for order %s", id)
	}

	order, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if uc.cache != nil {
		if err := uc.cache.Set(ctx, order); err != nil {
			log.Printf("[Cache] failed to set order %s: %v", id, err)
		}
	}

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
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

	// Invalidate cache after status change.
	if uc.cache != nil {
		if err := uc.cache.Invalidate(ctx, order.ID); err != nil {
			log.Printf("[Cache] failed to invalidate order %s: %v", order.ID, err)
		}
	}

	return order, nil
}

func (uc *OrderUseCase) ListOrders(status string) ([]*domain.Order, error) {
	return uc.repo.FindByStatus(status)
}

var ErrPaymentServiceUnavailable = fmt.Errorf("payment service unavailable")

func generateID() string {
	return uuid.NewString()
}

func generateTransactionID() string {
	return "txn_" + uuid.NewString()
}
