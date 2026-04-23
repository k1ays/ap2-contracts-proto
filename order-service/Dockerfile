FROM golang:1.22-alpine AS builder
WORKDIR /workspace

COPY contracts-generated /workspace/contracts-generated
COPY order-service /workspace/order-service

WORKDIR /workspace/order-service
RUN go mod download
RUN go build -o /order-service ./cmd/order-service
RUN go build -o /order-updates-client ./cmd/order-updates-client

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /order-service .
COPY --from=builder /order-updates-client .

EXPOSE 8080 9090

CMD ["./order-service"]
