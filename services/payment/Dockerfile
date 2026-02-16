# Multi-stage Dockerfile for Payment Service
# Build context: project root

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy shared module
COPY shared/ ./shared/

# Copy payment service
COPY services/payment/ ./services/payment/

# Build the application
WORKDIR /build/services/payment
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main \
    ./cmd/main.go

# Stage 2: Runtime (Distroless)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/services/payment/main /app/payment-service

# Copy migrations
COPY --from=builder /build/services/payment/migrations /app/migrations

# Expose port
EXPOSE 8083

# Use non-root user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/app/payment-service"]
