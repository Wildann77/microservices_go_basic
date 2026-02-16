# Multi-stage Dockerfile for Order Service
# Build context: project root

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy shared module
COPY shared/ ./shared/

# Copy order service
COPY services/order/ ./services/order/

# Build the application
WORKDIR /build/services/order
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main \
    ./cmd/main.go

# Stage 2: Runtime (Distroless)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/services/order/main /app/order-service

# Expose port
EXPOSE 8082

# Use non-root user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/app/order-service"]
