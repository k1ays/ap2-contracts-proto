package main

import (
	"ap2/order-service/internal/app"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	dsn := getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable")
	httpAddr := getEnv("ORDER_ADDR", ":8080")
	grpcAddr := getEnv("ORDER_GRPC_ADDR", ":9090")
	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:9091")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	cacheTTLSec := getEnvInt("CACHE_TTL", 300)
	cacheTTL := time.Duration(cacheTTLSec) * time.Second

	rateLimitMax := getEnvInt("RATE_LIMIT_MAX", 10)
	rateLimitWindowSec := getEnvInt("RATE_LIMIT_WINDOW", 60)
	rateLimitWindow := time.Duration(rateLimitWindowSec) * time.Second

	a, err := app.New(dsn, paymentGRPCAddr, redisAddr, cacheTTL, rateLimitMax, rateLimitWindow)
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

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
