package main

import (
	"ap2/payment-service/internal/app"
	"log"
	"os"
)

func main() {
	dsn := getEnv("PAYMENT_DB_DSN", "postgres://postgres:postgres@localhost:5433/payment_db?sslmode=disable")
	addr := getEnv("PAYMENT_GRPC_ADDR", ":9091")

	a, err := app.New(dsn)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer a.Close()

	log.Printf("Payment gRPC server listening on %s", addr)
	if err := a.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
