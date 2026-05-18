package app

import (
	orderv1 "ap2/contracts-generated/order/v1"
	rediscache "ap2/order-service/internal/infrastructure/redis"
	"ap2/order-service/internal/repository"
	transportgrpc "ap2/order-service/internal/transport/grpc"
	transporthttp "ap2/order-service/internal/transport/http"
	"ap2/order-service/internal/usecase"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	db          *sql.DB
	router      *gin.Engine
	httpServer  *http.Server
	grpcServer  *grpcpkg.Server
	paymentConn *grpcpkg.ClientConn
	updates     *transportgrpc.OrderUpdateBroker
	cache       *rediscache.OrderCache
}

func New(dsn, paymentGRPCAddr, redisAddr string, cacheTTL time.Duration, rateLimitMax int, rateLimitWindow time.Duration) (*App, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Println("Order DB connected")

	// Initialize Redis cache.
	orderCache, err := rediscache.NewOrderCache(redisAddr, cacheTTL)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init redis cache: %w", err)
	}

	paymentConn, err := grpcpkg.Dial(paymentGRPCAddr, grpcpkg.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = orderCache.Close()
		_ = db.Close()
		return nil, fmt.Errorf("dial payment grpc: %w", err)
	}

	updates, err := transportgrpc.NewOrderUpdateBroker(dsn)
	if err != nil {
		_ = paymentConn.Close()
		_ = orderCache.Close()
		_ = db.Close()
		return nil, fmt.Errorf("create order update broker: %w", err)
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	paymentClient := transportgrpc.NewPaymentGRPCClient(paymentConn)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient, orderCache)

	// Initialize rate limiter middleware.
	rateLimiter, err := rediscache.NewRateLimiterFromAddr(redisAddr, rateLimitMax, rateLimitWindow)
	if err != nil {
		log.Printf("Warning: rate limiter init failed: %v (proceeding without it)", err)
	}

	httpHandler := transporthttp.NewOrderHandler(orderUC)
	router := gin.Default()
	if rateLimiter != nil {
		router.Use(rateLimiter.Middleware())
		log.Printf("Rate limiter enabled: %d requests per %v", rateLimitMax, rateLimitWindow)
	}
	httpHandler.RegisterRoutes(router)

	grpcServer := grpcpkg.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, transportgrpc.NewOrderServer(orderUC, updates))

	return &App{
		db:          db,
		router:      router,
		grpcServer:  grpcServer,
		paymentConn: paymentConn,
		updates:     updates,
		cache:       orderCache,
	}, nil
}

func (a *App) Run(httpAddr, grpcAddr string) error {
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: a.router,
	}

	errCh := make(chan error, 2)

	go func() {
		log.Printf("Order gRPC server listening on %s", grpcAddr)
		errCh <- a.grpcServer.Serve(grpcListener)
	}()

	go func() {
		log.Printf("Order REST server listening on %s", httpAddr)
		errCh <- a.httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Order service shutting down gracefully...")
		a.Close()
		return nil
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, grpcpkg.ErrServerStopped) {
			return nil
		}
		return err
	}
}

func (a *App) Close() {
	if a.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpServer.Shutdown(ctx)
		cancel()
	}
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.paymentConn != nil {
		_ = a.paymentConn.Close()
	}
	if a.updates != nil {
		a.updates.Close()
	}
	if a.cache != nil {
		_ = a.cache.Close()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
}
