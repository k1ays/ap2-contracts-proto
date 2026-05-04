package grpc

import (
	"ap2/contracts-generated/payment/v1"
	"ap2/payment-service/internal/domain"
	"ap2/payment-service/internal/usecase"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{uc: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	resp, err := s.uc.Authorize(usecase.AuthorizeRequest{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	})
	if err == domain.ErrInvalidAmount {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &paymentv1.PaymentResponse{
		Id:            resp.ID,
		OrderId:       resp.OrderID,
		TransactionId: resp.TransactionID,
		Amount:        resp.Amount,
		Status:        resp.Status,
		CreatedAt:     timestamppb.New(resp.CreatedAt),
	}, nil
}

func (s *PaymentServer) GetPaymentByOrderID(ctx context.Context, req *paymentv1.GetPaymentByOrderIDRequest) (*paymentv1.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	payment, err := s.uc.GetByOrderID(req.GetOrderId())
	if err == domain.ErrPaymentNotFound {
		return nil, status.Error(codes.NotFound, "payment not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &paymentv1.PaymentResponse{
		Id:            payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     timestamppb.New(payment.CreatedAt),
	}, nil
}

func UnaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("grpc method=%s duration=%s", info.FullMethod, time.Since(start))
	return resp, err
}
