FROM golang:1.24-alpine AS builder
WORKDIR /workspace

COPY notification-service /workspace/notification-service

WORKDIR /workspace/notification-service
RUN go mod download
RUN go build -o /notification-service ./cmd/notification-service

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /notification-service .

CMD ["./notification-service"]
