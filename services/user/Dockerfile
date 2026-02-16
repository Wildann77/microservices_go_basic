# Multi-stage Dockerfile for User Service
# Build context: project root

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy shared module
COPY shared/ ./shared/

# Copy user service
COPY services/user/ ./services/user/

# Build the application
WORKDIR /build/services/user
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main \
    ./cmd/main.go

# Stage 2: Runtime (Distroless)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/services/user/main /app/user-service

# Copy migrations
COPY --from=builder /build/services/user/migrations /app/migrations

# Expose port
EXPOSE 8081

# Use non-root user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/app/user-service"]
