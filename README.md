# AP2 Assignment 4 — Performance Optimization & External Integrations

This project extends Assignments 1–3 by adding **Redis Caching**, **Background Job reliability**, and **External Provider Adapters** to the microservices system.

## What Changed (Assignment 4)

- **Redis Cache** added to Order Service (cache-aside pattern for `GET /orders/:id`).
- **Cache Invalidation**: cache is deleted immediately after any DB status update (create, cancel).
- **Notification Provider Adapter**: `NotificationProvider` interface with two implementations — `MockProvider` (simulated) and `SMTPProvider` (real SMTP).
- **Provider selection via env**: `PROVIDER_MODE=SIMULATED` or `PROVIDER_MODE=REAL`.
- **Redis-based Idempotency**: replaced in-memory map with Redis keys (`notification:processed:<payment_id>`) with 24h TTL.
- **Exponential Backoff Retries**: configurable retry policy with formula `backoff = INITIAL_BACKOFF * 2^(attempt-1)`.
- **API Rate Limiter (Bonus)**: Redis-based per-IP rate limiting middleware returning HTTP 429.
- **`.env` file**: all configuration (TTL, retry counts, API keys) in a single `.env` file.

## Architecture

```text
Client (curl/Postman)
        |
        | REST :8080 (Rate Limiter middleware)
        v
Order Service
  - Gin REST API
  - Redis Cache (cache-aside, TTL=5min)
  - gRPC client -> Payment Service :50051
  - gRPC streaming server :9090
  - PostgreSQL order_db
  - Cache invalidation on status change
        |
        | gRPC
        v
Payment Service (Producer)
  - gRPC server :50051
  - PostgreSQL payment_db
  - Publishes payment.completed events to RabbitMQ
        |
        | AMQP (RabbitMQ)
        v
  [payment.events exchange] --> [payment.completed queue]
        |                              |
        |                   [payment.events.dlx] (DLX)
        |                              |
        |                   [payment.completed.dlq] (DLQ)
        v
Notification Service (Background Worker)
  - Listens to payment.completed queue
  - NotificationProvider adapter (Mock / SMTP)
  - Exponential backoff retries (2s, 4s, 8s, 16s...)
  - Redis-based idempotency (prevents duplicate sends)
  - Manual ACKs, DLQ on exhausted retries
```

## 1. Caching Strategy (Order Service)

### Cache-Aside Pattern

- **Read Path (`GET /orders/:id`):**
  1. Check Redis for key `order:<id>`.
  2. **Cache HIT** → return cached data (no DB query).
  3. **Cache MISS** → query PostgreSQL, store in Redis with TTL, return data.

- **TTL**: Configurable via `CACHE_TTL` env var (default: 300s = 5 minutes).

### Cache Invalidation

**Atomic invalidation** — cache key is deleted immediately after any DB update:

- After `CreateOrder`: when status changes to `Paid`/`Failed`, key `order:<id>` is deleted.
- After `CancelOrder`: key `order:<id>` is deleted.
- **Delete over update**: avoids race conditions; next read repopulates from DB.

This ensures stale data (e.g., "Pending" for a paid order) is never served.

## 2. External Provider Adapter (Notification Service)

### Adapter Pattern

```go
type NotificationProvider interface {
    Send(to, subject, body string) error
}
```

| Provider       | Description                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `MockProvider` | Simulates latency (100–500ms) and 30% random failures for testing retries. |
| `SMTPProvider` | Real SMTP email delivery (requires SMTP_HOST, SMTP_PORT, etc.).            |

### Configuration

`PROVIDER_MODE` env var:
- `SIMULATED` (default) — MockProvider
- `REAL` — SMTPProvider

## 3. Retry Logic & Idempotency

### Exponential Backoff

```
Attempt 1: immediate
Attempt 2: wait 2s
Attempt 3: wait 4s
Attempt 4: wait 8s
Attempt 5: wait 16s
```

Formula: `backoff = INITIAL_BACKOFF * 2^(attempt-1)`

If all retries exhausted → message NACKed → sent to DLQ.

### Redis Idempotency

Before processing, check Redis key `notification:processed:<payment_id>`:
1. Key **exists** → duplicate, ACK and skip.
2. Key **does not exist** → send notification.
3. After success → set key with 24h TTL.

Prevents duplicate emails on message retry.

## 4. API Rate Limiter (Bonus +10%)

Redis-based per-IP rate limiting:
- Key: `rate_limit:<ip>`, counter incremented per request.
- Configurable: `RATE_LIMIT_MAX` (default: 10), `RATE_LIMIT_WINDOW` (default: 60s).
- Returns HTTP 429 with `Retry-After` header when exceeded.
- Response headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`.

## Project Structure

```text
├── contracts-proto/
├── contracts-generated/
├── order-service/
│   ├── cmd/order-service/
│   ├── cmd/order-updates-client/
│   ├── internal/domain/
│   ├── internal/usecase/
│   ├── internal/repository/
│   ├── internal/transport/http/
│   ├── internal/transport/grpc/
│   ├── internal/infrastructure/redis/    ← NEW: cache + rate limiter
│   └── migrations/
├── payment-service/
│   ├── cmd/payment-service/
│   ├── internal/domain/
│   ├── internal/usecase/
│   ├── internal/repository/
│   ├── internal/transport/grpc/
│   ├── internal/infrastructure/rabbitmq/
│   └── migrations/
├── notification-service/
│   ├── cmd/notification-service/
│   ├── internal/app/
│   ├── internal/consumer/                ← UPDATED: Redis idempotency + backoff
│   ├── internal/domain/
│   └── internal/provider/               ← NEW: adapter pattern
├── docker-compose.yml
├── .env
└── README.md
```

## Environment Variables

All configuration is in `.env`:

### Order Service

| Variable           | Default              | Description                        |
|--------------------|----------------------|------------------------------------|
| `ORDER_DB_DSN`     | postgres://...       | PostgreSQL DSN for order_db        |
| `ORDER_ADDR`       | :8080                | REST address                       |
| `ORDER_GRPC_ADDR`  | :9090                | gRPC streaming server address      |
| `PAYMENT_GRPC_ADDR`| payment-service:50051| Payment gRPC address               |
| `REDIS_ADDR`       | redis:6379           | Redis address                      |
| `CACHE_TTL`        | 300                  | Cache TTL in seconds               |
| `RATE_LIMIT_MAX`   | 10                   | Max requests per window            |
| `RATE_LIMIT_WINDOW`| 60                   | Rate limit window in seconds       |

### Payment Service

| Variable           | Default              | Description                        |
|--------------------|----------------------|------------------------------------|
| `PAYMENT_DB_DSN`   | postgres://...       | PostgreSQL DSN for payment_db      |
| `PAYMENT_GRPC_ADDR`| :50051               | gRPC server address                |
| `RABBITMQ_URL`     | amqp://...           | RabbitMQ connection URL            |

### Notification Service

| Variable           | Default              | Description                        |
|--------------------|----------------------|------------------------------------|
| `RABBITMQ_URL`     | amqp://...           | RabbitMQ connection URL            |
| `REDIS_ADDR`       | redis:6379           | Redis address for idempotency      |
| `PROVIDER_MODE`    | SIMULATED            | Provider mode (SIMULATED/REAL)     |
| `MAX_RETRIES`      | 5                    | Max retry attempts                 |
| `INITIAL_BACKOFF`  | 2                    | Initial backoff in seconds         |

## Running With Docker

```bash
docker compose down -v
docker compose up --build
```

Services:

| Service                | Port(s)       | Description                          |
|------------------------|---------------|--------------------------------------|
| `order-db`             | 5432          | PostgreSQL for orders                |
| `payment-db`           | 5433          | PostgreSQL for payments              |
| `redis`                | 6379          | Cache, rate limiting, idempotency    |
| `rabbitmq`             | 5672, 15672   | Message broker                       |
| `order-service`        | 8080, 9090    | REST + gRPC API                      |
| `payment-service`      | 50051         | gRPC payment processing              |
| `notification-service` | —             | Background worker                    |

## REST API Examples

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_001","item_name":"Laptop","amount":50000}'
```

### Get Order (with caching)

```bash
curl http://localhost:8080/orders/<id>
# First call: Cache MISS → DB query → cached
# Second call: Cache HIT → returned from Redis
```

### Cancel Order

```bash
curl -X PATCH http://localhost:8080/orders/<id>/cancel
# Cache invalidated after cancellation
```

### List Orders

```bash
curl http://localhost:8080/orders
curl "http://localhost:8080/orders?status=Paid"
```

## Graceful Shutdown

All services intercept `SIGINT`/`SIGTERM`:

- **Order Service**: shuts down HTTP server, stops gRPC, closes Redis + DB connections.
- **Payment Service**: stops gRPC, closes RabbitMQ publisher, closes DB.
- **Notification Service**: stops consumer loop, closes RabbitMQ + Redis connections.

## Contract-First Workspace

- `contracts-proto/payment/v1/payment.proto`
- `contracts-proto/order/v1/order.proto`
- `contracts-generated/` — generated Go code via local `replace` directives.

## Repository Links

- Proto Repository: https://github.com/k1ays/ap2-contracts-proto
- Generated Code Repository: https://github.com/k1ays/ap2-contracts-generated
