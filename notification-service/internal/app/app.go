package app

import (
	"ap2/notification-service/internal/consumer"
	"ap2/notification-service/internal/provider"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	consumer *consumer.Consumer
}

func New(amqpURL, redisAddr, providerMode string, maxRetries int, initialBackoff time.Duration, workerCount int) (*App, error) {
	var notifProvider provider.NotificationProvider

	switch providerMode {
	case "REAL":
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := os.Getenv("SMTP_PORT")
		smtpUser := os.Getenv("SMTP_USER")
		smtpPass := os.Getenv("SMTP_PASS")
		smtpFrom := os.Getenv("SMTP_FROM")
		if smtpHost == "" {
			smtpHost = "localhost"
		}
		if smtpPort == "" {
			smtpPort = "587"
		}
		notifProvider = provider.NewSMTPProvider(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
		log.Println("Using REAL (SMTP) notification provider")
	default:
		notifProvider = provider.NewMockProvider()
		log.Println("Using SIMULATED notification provider")
	}

	c, err := consumer.New(amqpURL, redisAddr, notifProvider, maxRetries, initialBackoff, workerCount)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}
	return &App{consumer: c}, nil
}

func (a *App) Run() error {
	done := make(chan struct{})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Notification service shutting down gracefully...")
		close(done)
	}()

	return a.consumer.Start(done)
}

func (a *App) Close() {
	if a.consumer != nil {
		a.consumer.Close()
	}
}
