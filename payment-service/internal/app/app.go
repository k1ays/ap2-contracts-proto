package app

import (
	"ap2/contracts-generated/payment/v1"
	"ap2/payment-service/internal/infrastructure/rabbitmq"
	"ap2/payment-service/internal/repository"
	transportgrpc "ap2/payment-service/internal/transport/grpc"
	"ap2/payment-service/internal/usecase"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

type App struct {
	db         *sql.DB
	grpcServer *grpc.Server
	publisher  *rabbitmq.Publisher
}

func New(dsn, amqpURL string) (*App, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Println("Payment DB connected")

	publisher, err := rabbitmq.NewPublisher(amqpURL)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("rabbitmq publisher: %w", err)
	}

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo, publisher)
	handler := transportgrpc.NewPaymentServer(uc)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportgrpc.UnaryLoggingInterceptor),
	)
	paymentv1.RegisterPaymentServiceServer(grpcServer, handler)

	return &App{
		db:         db,
		grpcServer: grpcServer,
		publisher:  publisher,
	}, nil
}

func (a *App) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Payment service shutting down gracefully...")
		a.Close()
	}()

	return a.grpcServer.Serve(lis)
}

func (a *App) Close() {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.publisher != nil {
		a.publisher.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
}
