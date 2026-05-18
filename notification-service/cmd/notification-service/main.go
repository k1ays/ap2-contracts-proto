package main

import (
	"ap2/notification-service/internal/app"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	amqpURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	providerMode := getEnv("PROVIDER_MODE", "SIMULATED")

	maxRetries := getEnvInt("MAX_RETRIES", 5)
	initialBackoffSec := getEnvInt("INITIAL_BACKOFF", 2)
	initialBackoff := time.Duration(initialBackoffSec) * time.Second
	workerCount := getEnvInt("WORKER_COUNT", 5)

	a, err := app.New(amqpURL, redisAddr, providerMode, maxRetries, initialBackoff, workerCount)
	if err != nil {
		log.Fatalf("failed to init notification service: %v", err)
	}
	defer a.Close()

	log.Println("Notification service started")
	if err := a.Run(); err != nil {
		log.Fatalf("notification service error: %v", err)
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
