package usecase

import (
	"ap2/payment-service/internal/domain"
	"fmt"
	"log"
	"time"
)

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) Authorize(req AuthorizeRequest) (*AuthorizeResponse, error) {
	payment, err := domain.NewPayment(req.OrderID, req.Amount)
	if err == domain.ErrInvalidAmount {
		return nil, err
	}
	if err == domain.ErrAmountExceedsLimit {
		declined := &domain.Payment{
			ID:            generateID(),
			OrderID:       req.OrderID,
			TransactionID: "",
			Amount:        req.Amount,
			Status:        domain.StatusDeclined,
			CreatedAt:     time.Now(),
		}
		if saveErr := uc.repo.Save(declined); saveErr != nil {
			log.Printf("failed to save declined payment for order %s: %v", req.OrderID, saveErr)
		}
		return &AuthorizeResponse{
			ID:            declined.ID,
			OrderID:       declined.OrderID,
			TransactionID: declined.TransactionID,
			Amount:        declined.Amount,
			Status:        declined.Status,
			CreatedAt:     declined.CreatedAt,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	payment.ID = generateID()
	payment.TransactionID = generateTransactionID()

	if err := uc.repo.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &AuthorizeResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     payment.CreatedAt,
	}, nil
}

func (uc *PaymentUseCase) GetByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.FindByOrderID(orderID)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateTransactionID() string {
	return fmt.Sprintf("txn_%d", time.Now().UnixNano())
}
