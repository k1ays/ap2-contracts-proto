FROM golang:1.24-alpine AS builder
WORKDIR /workspace

COPY contracts-generated /workspace/contracts-generated
COPY payment-service /workspace/payment-service

WORKDIR /workspace/payment-service
RUN go mod download
RUN go build -o /payment-service ./cmd/payment-service

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /payment-service .

EXPOSE 9091

CMD ["./payment-service"]
