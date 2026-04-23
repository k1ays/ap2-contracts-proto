package grpc

import (
	paymentv1 "ap2/contracts-generated/payment/v1"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

type PaymentGRPCClient struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentGRPCClient(conn grpc.ClientConnInterface) *PaymentGRPCClient {
	return &PaymentGRPCClient{
		client: paymentv1.NewPaymentServiceClient(conn),
	}
}

func (c *PaymentGRPCClient) Authorize(ctx context.Context, orderID string, amount int64) (string, string, error) {
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.ProcessPayment(callCtx, &paymentv1.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", "", fmt.Errorf("process payment via grpc: %w", err)
	}

	return resp.GetTransactionId(), resp.GetStatus(), nil
}
