# AP2 Assignment 3 - Event-Driven Architecture with Message Queues

This project extends Assignments 1 & 2 by introducing an **Event-Driven Architecture (EDA)** using **RabbitMQ** as the message broker. A new **Notification Service** consumes payment events asynchronously, fully decoupled from the Payment and Order services.

## What Changed (Assignment 3)

- **RabbitMQ** added as the message broker.
- **Payment Service** now acts as a **Producer**: after a successful DB transaction, it publishes a `payment.completed` event to RabbitMQ.
- **Notification Service** (new) acts as a **Consumer**: listens to the `payment.completed` queue, simulates sending email notifications.
- **Manual ACKs**: auto-ack is disabled; messages are acknowledged only after successful processing.
- **Durable Queues**: queues and messages are persistent and survive broker restarts.
- **Idempotency**: in-memory store tracks processed event IDs to prevent duplicate notifications.
- **Graceful Shutdown**: all three services handle `SIGINT`/`SIGTERM` via `os/signal` for clean resource cleanup.
- **Dead Letter Queue (DLQ)**: messages that fail processing after 3 retries or cannot be parsed are routed to a Dead Letter Queue (`payment.completed.dlq`).

## Architecture

```text
Client (curl/Postman)
        |
        | REST :8080
        v
Order Service
  - Gin REST API
  - gRPC client -> Payment Service :50051
  - gRPC streaming server :9090
  - PostgreSQL order_db
  - LISTEN/NOTIFY subscriber for order status changes
        |
        | gRPC
        v
Payment Service (Producer)
  - gRPC server :50051
  - Unary interceptor (method + duration logging)
  - PostgreSQL payment_db
  - Publishes payment.completed events to RabbitMQ
        |
        | AMQP (RabbitMQ)
        v
  [payment.events exchange] --routing_key:payment.completed--> [payment.completed queue]
        |                                                             |
        |                                           [payment.events.dlx] (Dead Letter Exchange)
        |                                                             |
        |                                           [payment.completed.dlq] (Dead Letter Queue)
        v
Notification Service (Consumer)
  - Listens to payment.completed queue
  - Manual ACKs (no auto-ack)
  - Idempotent consumer (in-memory event ID tracking)
  - Logs simulated email notifications
  - Graceful shutdown via os/signal
```

## Event Payload (JSON)

```json
{
  "event_id": "evt_order123_1714500000000000000",
  "order_id": "order123",
  "amount": 50000,
  "customer_email": "customer@example.com",
  "status": "Authorized"
}
```

## Idempotency Strategy

The Notification Service uses an **in-memory map** (`map[string]struct{}`) protected by a `sync.RWMutex` to track processed `event_id` values. Before processing any message:

1. The consumer checks if the `event_id` has already been processed.
2. If yes, it immediately ACKs the message without re-sending the notification (duplicate suppression).
3. If no, it processes the message, marks the `event_id` as processed, then ACKs.

This ensures that even if the same message is delivered multiple times (at-least-once delivery), the notification log is printed only once per unique event.

## ACK Logic Implementation

- **Auto-ACK is disabled**: the consumer is created with `autoAck: false`.
- Messages are acknowledged (`msg.Ack(false)`) **only after** the notification log is successfully printed and the event is marked as processed.
- If a message cannot be parsed (invalid JSON), it is **rejected** (`msg.Nack(false, false)`) and routed to the DLQ.
- If the consumer crashes mid-processing, the unacknowledged message remains in the queue and is redelivered.

## Dead Letter Queue (DLQ) - Bonus

The system implements a Dead Letter Queue for advanced failure handling:

- **DLX Exchange**: `payment.events.dlx` (direct exchange)
- **DLQ**: `payment.completed.dlq`
- Messages are sent to the DLQ when:
  - JSON parsing fails (malformed messages)
  - Processing fails after **3 retry attempts** (checked via `x-death` headers)
- The main queue (`payment.completed`) is configured with `x-dead-letter-exchange` and `x-dead-letter-routing-key` arguments.

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
    internal/infrastructure/rabbitmq/   <- NEW: RabbitMQ publisher
    migrations/
  notification-service/                  <- NEW SERVICE
    cmd/notification-service/
    internal/app/
    internal/consumer/
    internal/domain/
```

## Environment Variables

### Order Service

- `ORDER_DB_DSN` - PostgreSQL DSN for `order_db`
- `ORDER_ADDR` - REST address, default `:8080`
- `ORDER_GRPC_ADDR` - gRPC streaming server address, default `:9090`
- `PAYMENT_GRPC_ADDR` - payment gRPC server address, default `localhost:50051`

### Payment Service

- `PAYMENT_DB_DSN` - PostgreSQL DSN for `payment_db`
- `PAYMENT_GRPC_ADDR` - gRPC server address, default `:50051`
- `RABBITMQ_URL` - RabbitMQ connection URL, default `amqp://guest:guest@localhost:5672/`

### Notification Service

- `RABBITMQ_URL` - RabbitMQ connection URL, default `amqp://guest:guest@localhost:5672/`

## Running With Docker

```bash
docker compose down -v
docker compose up --build
```

Services:

- `order-db` -> `localhost:5432`
- `payment-db` -> `localhost:5433`
- `rabbitmq` -> `localhost:5672` (AMQP), `localhost:15672` (Management UI, guest/guest)
- `order-service` REST -> `localhost:8080`
- `order-service` gRPC stream -> `localhost:9090`
- `payment-service` gRPC -> `localhost:50051`
- `notification-service` (no exposed ports, consumes from RabbitMQ)

## Running Without Docker

If Go is installed locally:

```bash
# Terminal 1 - Start RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management-alpine

# Terminal 2
cd payment-service
PAYMENT_DB_DSN="postgres://postgres:postgres@localhost:5433/payment_db?sslmode=disable" \
PAYMENT_GRPC_ADDR=":50051" \
RABBITMQ_URL="amqp://guest:guest@localhost:5672/" \
go run ./cmd/payment-service

# Terminal 3
cd order-service
ORDER_DB_DSN="postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable" \
ORDER_ADDR=":8080" \
ORDER_GRPC_ADDR=":9090" \
PAYMENT_GRPC_ADDR="localhost:50051" \
go run ./cmd/order-service

# Terminal 4
cd notification-service
RABBITMQ_URL="amqp://guest:guest@localhost:5672/" \
go run ./cmd/notification-service
```

## REST API Examples

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_001","item_name":"Laptop","amount":50000}'
```

Expected console output from Notification Service:
```
[Notification] Sent email to customer@example.com for Order #<order_id>. Amount: $500.00. Status: Authorized
```

### Create Declined Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_002","item_name":"Supercar","amount":200000}'
```

The order is created but declined (amount exceeds limit). A notification event is still published with status `Declined`.

### Get Order

```bash
curl http://localhost:8080/orders/<id>
```

### Cancel Order

```bash
curl -X PATCH http://localhost:8080/orders/<id>/cancel
```

### List Orders

```bash
curl http://localhost:8080/orders
curl "http://localhost:8080/orders?status=Paid"
```

## Graceful Shutdown

All three services intercept `SIGINT` and `SIGTERM` signals using `os/signal`:

- **Order Service**: shuts down HTTP server with timeout, stops gRPC server gracefully, closes DB and gRPC connections.
- **Payment Service**: stops gRPC server gracefully, closes RabbitMQ publisher connection, closes DB.
- **Notification Service**: signals the consumer loop to stop, closes RabbitMQ connection.

## Contract-First Workspace

`contracts-proto/` simulates the dedicated proto repository required by the assignment.

- `contracts-proto/payment/v1/payment.proto`
- `contracts-proto/order/v1/order.proto`

`contracts-generated/` contains the generated Go code consumed by both services through local `replace` directives.

## Repository Links

- Proto Repository: https://github.com/k1ays/ap2-contracts-proto
- Generated Code Repository: https://github.com/k1ays/ap2-contracts-generated
