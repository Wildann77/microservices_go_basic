# Multi-stage Dockerfile for Gateway Service
# Build context: project root

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy shared module
COPY shared/ ./shared/

# Copy gateway service
COPY gateway/ ./gateway/

# Build the application
WORKDIR /build/gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main \
    ./cmd/main.go

# Stage 2: Runtime (Distroless)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gateway/main /app/gateway

# Expose port
EXPOSE 4000

# Use non-root user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/app/gateway"]
