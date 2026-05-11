package main

import (
	orderv1 "ap2/contracts-generated/order/v1"
	"context"
	"io"
	"log"
	"os"

	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: order-updates-client <order-id>")
	}

	addr := getEnv("ORDER_GRPC_ADDR", "localhost:9090")
	orderID := os.Args[1]

	conn, err := grpcpkg.Dial(addr, grpcpkg.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial order grpc: %v", err)
	}
	defer conn.Close()

	client := orderv1.NewOrderServiceClient(conn)
	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderv1.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatalf("subscribe to order updates: %v", err)
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("receive stream event: %v", err)
		}

		log.Printf(
			"order_id=%s status=%s amount=%d published_at=%s",
			update.GetOrderId(),
			update.GetStatus(),
			update.GetAmount(),
			update.GetPublishedAt().AsTime().Format("2006-01-02 15:04:05"),
		)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
