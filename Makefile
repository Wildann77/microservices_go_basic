# Microservices Go - Makefile
# ===========================

.PHONY: help build up down restart logs clean test lint proto

# Default target
help:
	@echo "Microservices Go - Available Commands:"
	@echo ""
	@echo "  make build          - Build all Docker images"
	@echo "  make up             - Start all services with Docker Compose"
	@echo "  make up-d           - Start all services in detached mode"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo "  make logs           - View logs from all services"
	@echo "  make logs-f         - Follow logs from all services"
	@echo "  make clean          - Remove all containers and volumes"
	@echo "  make clean-all      - Remove everything including images"
	@echo ""
	@echo "  make run-gateway    - Run GraphQL Gateway locally"
	@echo "  make run-user       - Run User Service locally"
	@echo "  make run-order      - Run Order Service locally"
	@echo "  make run-payment    - Run Payment Service locally"
	@echo ""
	@echo "  make test           - Run all tests"
	@echo "  make test-user      - Run User Service tests"
	@echo "  make test-order     - Run Order Service tests"
	@echo "  make test-payment   - Run Payment Service tests"
	@echo ""
	@echo "  make migrate-up     - Run all database migrations"
	@echo "  make migrate-down   - Rollback all database migrations"
	@echo ""
	@echo "  make deps           - Download all Go dependencies"
	@echo "  make tidy           - Tidy all Go modules"
	@echo "  make lint           - Run linter on all services"
	@echo ""
	@echo "  make health         - Check health of all services"
	@echo "  make seed           - Seed database with test data"

# Build commands
build:
	docker-compose build

build-no-cache:
	docker-compose build --no-cache

# Docker Compose commands
up:
	docker-compose up

up-d:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

logs:
	docker-compose logs

logs-f:
	docker-compose logs -f

logs-user:
	docker-compose logs -f user-service

logs-order:
	docker-compose logs -f order-service

logs-payment:
	docker-compose logs -f payment-service

# Cleanup commands
clean:
	docker-compose down -v

docker system prune -f

clean-all:
	docker-compose down -v --rmi all
	docker system prune -af

# Local development commands
run-gateway:
	@echo "Starting GraphQL Gateway..."
	cd gateway && go run cmd/main.go

run-user:
	@echo "Starting User Service..."
	cd services/user && USER_PORT=8081 go run cmd/main.go

run-order:
	@echo "Starting Order Service..."
	cd services/order && ORDER_PORT=8082 go run cmd/main.go

run-payment:
	@echo "Starting Payment Service..."
	cd services/payment && PAYMENT_PORT=8083 go run cmd/main.go

# Run all services locally (requires databases)
run-all-local:
	@echo "Make sure databases are running: make up-d"
	@echo "Then run each service in separate terminals:"
	@echo "  make run-user"
	@echo "  make run-order"
	@echo "  make run-payment"
	@echo "  make run-gateway"

# Test commands
test:
	@echo "Running all tests..."
	cd shared && go test ./...
	cd services/user && go test ./...
	cd services/order && go test ./...
	cd services/payment && go test ./...

test-user:
	cd services/user && go test -v ./...

test-order:
	cd services/order && go test -v ./...

test-payment:
	cd services/payment && go test -v ./...

test-coverage:
	cd services/user && go test -cover ./...
	cd services/order && go test -cover ./...
	cd services/payment && go test -cover ./...

# Database migration commands
migrate-up:
	@echo "Running database migrations..."
	@echo "User Service migrations:"
	docker-compose exec -T postgres-user psql -U postgres -d user -f /docker-entrypoint-initdb.d/001_create_users.up.sql || true
	@echo "Order Service migrations:"
	docker-compose exec -T postgres-order psql -U postgres -d order -f /docker-entrypoint-initdb.d/001_create_orders.up.sql || true
	@echo "Payment Service migrations:"
	docker-compose exec -T postgres-payment psql -U postgres -d payment -f /docker-entrypoint-initdb.d/001_create_payments.up.sql || true

migrate-down:
	@echo "Rolling back database migrations..."
	docker-compose exec -T postgres-user psql -U postgres -d user -f /docker-entrypoint-initdb.d/001_create_users.down.sql || true
	docker-compose exec -T postgres-order psql -U postgres -d order -f /docker-entrypoint-initdb.d/001_create_orders.down.sql || true
	docker-compose exec -T postgres-payment psql -U postgres -d payment -f /docker-entrypoint-initdb.d/001_create_payments.down.sql || true

# Dependency management
deps:
	cd shared && go mod download
	cd gateway && go mod download
	cd services/user && go mod download
	cd services/order && go mod download
	cd services/payment && go mod download

tidy:
	cd shared && go mod tidy
	cd gateway && go mod tidy
	cd services/user && go mod tidy
	cd services/order && go mod tidy
	cd services/payment && go mod tidy

update-deps:
	cd shared && go get -u ./...
	cd gateway && go get -u ./...
	cd services/user && go get -u ./...
	cd services/order && go get -u ./...
	cd services/payment && go get -u ./...

# Linting
lint:
	@echo "Running linter..."
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	cd shared && golangci-lint run ./...
	cd gateway && golangci-lint run ./...
	cd services/user && golangci-lint run ./...
	cd services/order && golangci-lint run ./...
	cd services/payment && golangci-lint run ./...

# Code formatting
fmt:
	cd shared && go fmt ./...
	cd gateway && go fmt ./...
	cd services/user && go fmt ./...
	cd services/order && go fmt ./...
	cd services/payment && go fmt ./...

# Health checks
health:
	@echo "Checking service health..."
	@echo "User Service:"
	curl -s http://localhost:8081/health | jq . || echo "Not healthy"
	@echo "Order Service:"
	curl -s http://localhost:8082/health | jq . || echo "Not healthy"
	@echo "Payment Service:"
	curl -s http://localhost:8083/health | jq . || echo "Not healthy"

# Database seeding
seed:
	@echo "Seeding database with test data..."
	@echo "Use GraphQL mutations or REST API to create test data"
	@echo "Example:"
	@echo '  curl -X POST http://localhost:8081/api/v1/users/register \\'
	@echo '    -H "Content-Type: application/json" \\'
	@echo '    -d '\''{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}'\'''

# GraphQL playground
playground:
	@echo "Opening GraphQL Playground..."
	@echo "Visit: http://localhost:4000/"
	@echo "Query endpoint: http://localhost:4000/query"

# API documentation
api-docs:
	@echo "API Endpoints:"
	@echo ""
	@echo "GraphQL Gateway:"
	@echo "  Playground: http://localhost:4000/"
	@echo "  Endpoint:   http://localhost:4000/query"
	@echo ""
	@echo "REST Services:"
	@echo "  User Service:    http://localhost:8081/api/v1/users"
	@echo "  Order Service:   http://localhost:8082/api/v1/orders"
	@echo "  Payment Service: http://localhost:8083/api/v1/payments"
	@echo ""
	@echo "Health Checks:"
	@echo "  User Service:    http://localhost:8081/health"
	@echo "  Order Service:   http://localhost:8082/health"
	@echo "  Payment Service: http://localhost:8083/health"
	@echo ""
	@echo "Infrastructure:"
	@echo "  RabbitMQ Management: http://localhost:15672 (guest/guest)"
	@echo "  Jaeger UI:          http://localhost:16686"
	@echo "  Prometheus:         http://localhost:9090"

# Generate GraphQL code (requires gqlgen)
generate:
	cd gateway && go run github.com/99designs/gqlgen generate

# Initialize project
init:
	@echo "Initializing project..."
	make deps
	@echo "Project initialized. Run 'make up-d' to start services."

# Development mode - start infrastructure only
infra:
	docker-compose up -d postgres-user postgres-order postgres-payment rabbitmq jaeger prometheus
	@echo "Infrastructure started. Waiting for services to be ready..."
	@sleep 5
	@echo "Run individual services locally with:"
	@echo "  make run-user"
	@echo "  make run-order"
	@echo "  make run-payment"
	@echo "  make run-gateway"

# Stop infrastructure
infra-down:
	docker-compose stop postgres-user postgres-order postgres-payment rabbitmq jaeger prometheus

# Full reset
reset: clean-all
	@echo "Full reset complete. Run 'make up-d' to start fresh."