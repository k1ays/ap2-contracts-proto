package main

import (
	"ap2/notification-service/internal/app"
	"log"
	"os"
)

func main() {
	amqpURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	a, err := app.New(amqpURL)
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
