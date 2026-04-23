package http

import (
	"ap2/payment-service/internal/domain"
	"ap2/payment-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/payments", h.Authorize)
	r.GET("/payments/:order_id", h.GetByOrderID)
}

type authorizeRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required"`
}

type authorizeResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.uc.Authorize(usecase.AuthorizeRequest{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err == domain.ErrInvalidAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusOK
	if resp.Status == domain.StatusDeclined {
		statusCode = http.StatusUnprocessableEntity
	}

	c.JSON(statusCode, authorizeResponse{
		ID:            resp.ID,
		OrderID:       resp.OrderID,
		TransactionID: resp.TransactionID,
		Amount:        resp.Amount,
		Status:        resp.Status,
		CreatedAt:     resp.CreatedAt.Format(http.TimeFormat),
	})
}

func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetByOrderID(orderID)
	if err == domain.ErrPaymentNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             payment.ID,
		"order_id":       payment.OrderID,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"status":         payment.Status,
		"created_at":     payment.CreatedAt,
	})
}
