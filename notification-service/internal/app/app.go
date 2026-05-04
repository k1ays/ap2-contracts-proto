package app

import (
	"ap2/notification-service/internal/consumer"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	consumer *consumer.Consumer
}

func New(amqpURL string) (*App, error) {
	c, err := consumer.New(amqpURL)
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
