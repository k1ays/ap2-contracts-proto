package main

import (
	"ap2/order-service/internal/app"
	"log"
	"os"
)

func main() {
	dsn := getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable")
	httpAddr := getEnv("ORDER_ADDR", ":8080")
	grpcAddr := getEnv("ORDER_GRPC_ADDR", ":9090")
	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:9091")

	a, err := app.New(dsn, paymentGRPCAddr)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer a.Close()

	if err := a.Run(httpAddr, grpcAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
