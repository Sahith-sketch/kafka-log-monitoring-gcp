FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o audit-logs-monitoring ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/audit-logs-monitoring .

# Environment variables for feature toggles
ENV IS_KAFKA_ENABLED=true
ENV IS_GCS_ENABLED=true
ENV IS_NOTIFICATIONS_ENABLED=true

# Default configuration
ENV PORT=8080
ENV BATCH_SIZE=100
ENV WORKER_COUNT=10
ENV KAFKA_BROKERS=localhost:9092
ENV KAFKA_TOPIC=audit-logs
ENV GCS_BUCKET=audit-logs-storage

# TLS Certificate Secret Names
ENV TLS_CA_CERT_SECRET=audit-processor-tls-ca-cert
ENV TLS_CLIENT_CERT_SECRET=audit-processor-tls-client-cert
ENV TLS_CLIENT_KEY_SECRET=audit-processor-tls-client-key

EXPOSE 8080
CMD ["./audit-logs-monitoring"]