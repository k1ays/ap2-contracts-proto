package app

import (
	"ap2/contracts-generated/payment/v1"
	"ap2/payment-service/internal/repository"
	transportgrpc "ap2/payment-service/internal/transport/grpc"
	"ap2/payment-service/internal/usecase"
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

type App struct {
	db         *sql.DB
	grpcServer *grpc.Server
}

func New(dsn string) (*App, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Println("Payment DB connected")

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := transportgrpc.NewPaymentServer(uc)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportgrpc.UnaryLoggingInterceptor),
	)
	paymentv1.RegisterPaymentServiceServer(grpcServer, handler)

	return &App{
		db:         db,
		grpcServer: grpcServer,
	}, nil
}

func (a *App) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	return a.grpcServer.Serve(lis)
}

func (a *App) Close() {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.db != nil {
		a.db.Close()
	}
}
