package grpc

import (
	orderv1 "ap2/contracts-generated/order/v1"
	"ap2/order-service/internal/domain"
	"ap2/order-service/internal/usecase"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	uc      *usecase.OrderUseCase
	updates *OrderUpdateBroker
}

func NewOrderServer(uc *usecase.OrderUseCase, updates *OrderUpdateBroker) *OrderServer {
	return &OrderServer{
		uc:      uc,
		updates: updates,
	}
}

func (s *OrderServer) SubscribeToOrderUpdates(
	req *orderv1.OrderRequest,
	stream orderv1.OrderService_SubscribeToOrderUpdatesServer,
) error {
	if req.GetOrderId() == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.uc.GetOrder(req.GetOrderId())
	if err == domain.ErrOrderNotFound {
		return status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	lastStatus := order.Status
	if err := stream.Send(orderToProto(order, time.Now())); err != nil {
		return err
	}

	updates, unsubscribe := s.updates.Subscribe(req.GetOrderId())
	defer unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update := <-updates:
			if update.Status == lastStatus {
				continue
			}
			lastStatus = update.Status
			if err := stream.Send(updateToProto(update, time.Now())); err != nil {
				return err
			}
		}
	}
}

func orderToProto(order *domain.Order, publishedAt time.Time) *orderv1.OrderStatusUpdate {
	return &orderv1.OrderStatusUpdate{
		OrderId:     order.ID,
		CustomerId:  order.CustomerID,
		ItemName:    order.ItemName,
		Amount:      order.Amount,
		Status:      order.Status,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		PublishedAt: timestamppb.New(publishedAt),
	}
}

func updateToProto(update OrderUpdate, publishedAt time.Time) *orderv1.OrderStatusUpdate {
	return &orderv1.OrderStatusUpdate{
		OrderId:     update.ID,
		CustomerId:  update.CustomerID,
		ItemName:    update.ItemName,
		Amount:      update.Amount,
		Status:      update.Status,
		CreatedAt:   timestamppb.New(update.CreatedAt),
		PublishedAt: timestamppb.New(publishedAt),
	}
}
