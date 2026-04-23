# AP2 Assignment 2 - gRPC Migration & Contract-First Development

This project migrates the internal communication between `order-service` and `payment-service` from REST to gRPC while preserving the Clean Architecture structure from Assignment 1.

## What Changed

- External API is still REST on `order-service` (`POST /orders`, `GET /orders/:id`, `PATCH /orders/:id/cancel`).
- Internal call `order-service -> payment-service` is now gRPC.
- `payment-service` is now a gRPC server.
- `order-service` is also a gRPC server for server-side streaming:
  - `SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate)`
- Order updates are tied to real database changes through PostgreSQL `LISTEN/NOTIFY`.
- Contracts are managed in a separate local workspace:
  - `contracts-proto/` - source `.proto` files
  - `contracts-generated/` - generated Go code consumed by both services

## Architecture

```text
Client (curl/Postman)
        |
        | REST :8080
        v
Order Service
  - Gin REST API
  - gRPC client -> Payment Service :9091
  - gRPC streaming server :9090
  - PostgreSQL order_db
  - LISTEN/NOTIFY subscriber for order status changes
        |
        | gRPC
        v
Payment Service
  - gRPC server :9091
  - Unary interceptor (method + duration logging)
  - PostgreSQL payment_db
```

## Contract-First Workspace

`contracts-proto/` simulates the dedicated proto repository required by the assignment.

- [contracts-proto/payment/v1/payment.proto](/C:/Users/tasim/Documents/New%20project/ap2_assignment1/contracts-proto/payment/v1/payment.proto)
- [contracts-proto/order/v1/order.proto](/C:/Users/tasim/Documents/New%20project/ap2_assignment1/contracts-proto/order/v1/order.proto)
- [contracts-proto/.github/workflows/generate-go.yml](/C:/Users/tasim/Documents/New%20project/ap2_assignment1/contracts-proto/.github/workflows/generate-go.yml)

`contracts-generated/` simulates the generated-code repository:

- [contracts-generated/go.mod](/C:/Users/tasim/Documents/New%20project/ap2_assignment1/contracts-generated/go.mod)
- generated `*.pb.go` and `*_grpc.pb.go` files are consumed by both services through local `replace` directives.

For final submission, push these two folders into separate GitHub repositories and replace the placeholders below with real URLs.

## Repository Links

- Proto Repository: push `contracts-proto/` and add the GitHub URL here.
- Generated Code Repository: push `contracts-generated/` and add the GitHub URL here.

## Project Structure

```text
ap2_assignment1/
  contracts-proto/
  contracts-generated/
  order-service/
    cmd/order-service/
    cmd/order-updates-client/
    internal/domain/
    internal/usecase/
    internal/repository/
    internal/transport/http/
    internal/transport/grpc/
    migrations/
  payment-service/
    cmd/payment-service/
    internal/domain/
    internal/usecase/
    internal/repository/
    internal/transport/grpc/
    migrations/
```

## Environment Variables

### Order Service

- `ORDER_DB_DSN` - PostgreSQL DSN for `order_db`
- `ORDER_ADDR` - REST address, default `:8080`
- `ORDER_GRPC_ADDR` - gRPC streaming server address, default `:9090`
- `PAYMENT_GRPC_ADDR` - payment gRPC server address, default `localhost:9091`

### Payment Service

- `PAYMENT_DB_DSN` - PostgreSQL DSN for `payment_db`
- `PAYMENT_GRPC_ADDR` - gRPC server address, default `:9091`

## Running With Docker

Use a fresh database volume at least once, because Assignment 2 adds a PostgreSQL trigger for order update notifications.

```bash
docker compose down -v
docker compose up --build
```

Services:

- `order-db` -> `localhost:5432`
- `payment-db` -> `localhost:5433`
- `order-service` REST -> `localhost:8080`
- `order-service` gRPC stream -> `localhost:9090`
- `payment-service` gRPC -> `localhost:9091`

## Running Without Docker

If Go is installed locally:

```bash
# Terminal 1
cd payment-service
PAYMENT_DB_DSN="postgres://postgres:postgres@localhost:5433/payment_db?sslmode=disable" \
PAYMENT_GRPC_ADDR=":9091" \
go run ./cmd/payment-service

# Terminal 2
cd order-service
ORDER_DB_DSN="postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable" \
ORDER_ADDR=":8080" \
ORDER_GRPC_ADDR=":9090" \
PAYMENT_GRPC_ADDR="localhost:9091" \
go run ./cmd/order-service
```

## REST API Examples

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_001","item_name":"Laptop","amount":50000}'
```

Example response:

```json
{
  "id": "9db4f0f3-2b0f-4cb9-a02e-f4d1cc0d3f5f",
  "customer_id": "cust_001",
  "item_name": "Laptop",
  "amount": 50000,
  "status": "Paid",
  "created_at": "2026-04-16T10:00:00Z"
}
```

### Create Declined Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_002","item_name":"Supercar","amount":200000}'
```

The order is still created, but the final status becomes `Failed`.

### Get Order

```bash
curl http://localhost:8080/orders/<order-id>
```

### Cancel Order

```bash
curl -X PATCH http://localhost:8080/orders/<order-id>/cancel
```

## gRPC Streaming Demo

The streaming RPC is backed by PostgreSQL notifications. The stream reacts to actual database changes, not to artificial `sleep` loops.

### Option A: demo from Docker

1. Create an order via REST and copy the returned `id`.
2. Start the subscriber:

```bash
docker compose exec order-service ./order-updates-client <order-id>
```

3. Trigger a real DB-backed change:

```bash
docker compose exec order-db \
  psql -U postgres -d order_db \
  -c "UPDATE orders SET status='Cancelled' WHERE id='<order-id>';"
```

4. The subscriber immediately prints the new status.

This SQL update is intended only for streaming demonstration. Business rules are still enforced in the application layer for normal API calls.

### Option B: local Go client

```bash
cd order-service
ORDER_GRPC_ADDR="localhost:9090" go run ./cmd/order-updates-client <order-id>
```

## gRPC Contracts

### Payment Service

- `ProcessPayment(PaymentRequest) returns (PaymentResponse)`
- `GetPaymentByOrderID(GetPaymentByOrderIDRequest) returns (PaymentResponse)`

### Order Service

- `SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate)`

Both contracts use proper package names and `go_package` options. Timestamp fields use `google.protobuf.Timestamp`.

## Error Handling

- Invalid payment amount -> gRPC `InvalidArgument`
- Payment not found -> gRPC `NotFound`
- Order not found in stream subscription -> gRPC `NotFound`
- Payment transport failure from order service -> REST `503 Service Unavailable`
- Payment decline due to amount limit -> successful gRPC response with status `Declined`

## Bonus

The payment service includes a unary gRPC interceptor that logs:

- full gRPC method name
- request duration

## Verification

The updated code was verified with:

```bash
docker run --rm -v "<project-root>:/src" -w /src/order-service golang:1.22 bash -lc "/usr/local/go/bin/go test ./..."
docker run --rm -v "<project-root>:/src" -w /src/payment-service golang:1.22 bash -lc "/usr/local/go/bin/go test ./..."
```

Both modules compile successfully.
