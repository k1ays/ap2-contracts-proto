package http

import (
	"ap2/order-service/internal/domain"
	"ap2/order-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
	r.GET("/orders", h.ListOrders)
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.uc.CreateOrder(c.Request.Context(), req.CustomerID, req.ItemName, req.Amount)
	if err == usecase.ErrPaymentServiceUnavailable {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service unavailable"})
		return
	}
	if err == domain.ErrInvalidAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, orderResponse(order))
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	order, err := h.uc.GetOrder(c.Param("id"))
	if err == domain.ErrOrderNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orderResponse(order))
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	order, err := h.uc.CancelOrder(c.Param("id"))
	if err == domain.ErrOrderNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err == domain.ErrCannotCancelPaid || err == domain.ErrCannotCancelOrder {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orderResponse(order))
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	status := c.Query("status")

	validStatuses := map[string]bool{
		domain.StatusPending:   true,
		domain.StatusPaid:      true,
		domain.StatusFailed:    true,
		domain.StatusCancelled: true,
		"":                     true,
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status, use: Pending, Paid, Failed, Cancelled"})
		return
	}

	orders, err := h.uc.ListOrders(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		response = append(response, orderResponse(order))
	}
	c.JSON(http.StatusOK, response)
}

func orderResponse(o *domain.Order) gin.H {
	return gin.H{
		"id":          o.ID,
		"customer_id": o.CustomerID,
		"item_name":   o.ItemName,
		"amount":      o.Amount,
		"status":      o.Status,
		"created_at":  o.CreatedAt,
	}
}
