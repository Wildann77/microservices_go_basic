# Microservices Go - Makefile
# ===========================

.PHONY: help

# =============================================================================
# HELP
# =============================================================================

help:
	@echo "╔══════════════════════════════════════════════════════════════════╗"
	@echo "║          Microservices Go - Available Commands                   ║"
	@echo "╚══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🚀 QUICK START:"
	@echo "   make setup        - Full k3s setup (build + import + deploy)"
	@echo "   make up           - Start with Docker Compose"
	@echo ""
	@echo "📦 K3S DEPLOYMENT:"
	@echo "   make k3s-setup    - Setup k3s environment"
	@echo "   make k3s-build    - Build binaries & images"
	@echo "   make k3s-import   - Import images to k3s (sudo)"
	@echo "   make k3s-deploy   - Deploy to Kubernetes"
	@echo "   make k3s-status   - Check k3s status"
	@echo "   make k3s-logs     - View k3s logs (SERVICE=name)"
	@echo "   make k3s-clean    - Remove k3s resources"
	@echo ""
	@echo "🐳 DOCKER COMPOSE:"
	@echo "   make up           - Start all services"
	@echo "   make up-d         - Start in detached mode"
	@echo "   make down         - Stop all services"
	@echo "   make restart      - Restart services"
	@echo "   make logs         - View logs"
	@echo ""
	@echo "💻 LOCAL DEVELOPMENT:"
	@echo "   make run-user     - Run User Service locally"
	@echo "   make run-order    - Run Order Service locally"
	@echo "   make run-payment  - Run Payment Service locally"
	@echo "   make run-gateway  - Run Gateway locally"
	@echo "   make dev-all      - Run all with hot reload"
	@echo ""
	@echo "🧪 TESTING & MIGRATIONS:"
	@echo "   make test         - Run all tests"
	@echo "   make migrate-up   - Run migrations"
	@echo "   make health       - Check health"
	@echo ""
	@echo "📊 EXAMPLES:"
	@echo "   make k3s-logs SERVICE=gateway    # View gateway logs"
	@echo "   make k3s-scale SERVICE=user REPLICAS=3"

# =============================================================================
# K3S DEPLOYMENT
# =============================================================================

# Full automated setup
k3s-setup:
	@chmod +x setup.sh
	@./setup.sh

# Setup environment only
k3s-setup-env:
	@chmod +x scripts/k3s/setup.sh
	@./scripts/k3s/setup.sh

# Build binaries and images
k3s-build:
	@chmod +x scripts/k3s/build.sh
	@./scripts/k3s/build.sh

# Import images to k3s (requires sudo)
k3s-import:
	@chmod +x scripts/k3s/import.sh
	@sudo ./scripts/k3s/import.sh

# Deploy to Kubernetes
k3s-deploy:
	@chmod +x scripts/k3s/deploy.sh
	@./scripts/k3s/deploy.sh

# Check status
k3s-status:
	@chmod +x scripts/k3s/status.sh
	@./scripts/k3s/status.sh

# View logs (usage: make k3s-logs SERVICE=gateway)
k3s-logs:
ifndef SERVICE
	@echo "Usage: make k3s-logs SERVICE=<service>"
	@echo "Services: gateway, user, order, payment, all"
	@exit 1
endif
	@chmod +x scripts/k3s/logs.sh
	@./scripts/k3s/logs.sh $(SERVICE)

# Clean up all resources
k3s-clean:
	@chmod +x scripts/k3s/cleanup.sh
	@./scripts/k3s/cleanup.sh

# Restart deployments
k3s-restart:
	@export KUBECONFIG=$$HOME/.kube/config && \
	kubectl rollout restart deployment/gateway -n microservices && \
	kubectl rollout restart deployment/user-service -n microservices && \
	kubectl rollout restart deployment/order-service -n microservices && \
	kubectl rollout restart deployment/payment-service -n microservices

# Scale services
k3s-scale:
ifndef SERVICE
	@echo "Usage: make k3s-scale SERVICE=<name> REPLICAS=<number>"
	@exit 1
endif
ifndef REPLICAS
	@echo "Usage: make k3s-scale SERVICE=<name> REPLICAS=<number>"
	@exit 1
endif
	@export KUBECONFIG=$$HOME/.kube/config && \
	kubectl scale deployment $(SERVICE)-service --replicas=$(REPLICAS) -n microservices

# =============================================================================
# DOCKER COMPOSE
# =============================================================================

build:
	docker compose build

build-no-cache:
	docker compose build --no-cache

up:
	docker compose up

up-d:
	docker compose up -d

down:
	docker compose down

down-v:
	docker compose down -v

restart:
	docker compose restart

logs:
	docker compose logs

logs-f:
	docker compose logs -f

logs-user:
	docker compose logs -f user-service

logs-order:
	docker compose logs -f order-service

logs-payment:
	docker compose logs -f payment-service

clean:
	docker compose down -v
	docker system prune -f

clean-all:
	docker compose down -v --rmi all
	docker system prune -af

# =============================================================================
# LOCAL DEVELOPMENT
# =============================================================================

run-gateway:
	@cd gateway && go run cmd/main.go

run-user:
	@cd services/user && USER_PORT=8081 go run cmd/main.go

run-order:
	@cd services/order && ORDER_PORT=8082 go run cmd/main.go

run-payment:
	@cd services/payment && PAYMENT_PORT=8083 go run cmd/main.go

run-all-local:
	@echo "Run each in separate terminals:"
	@echo "  make run-user"
	@echo "  make run-order"
	@echo "  make run-payment"
	@echo "  make run-gateway"

dev-gateway:
	@cd gateway && air

dev-user:
	@cd services/user && air

dev-order:
	@cd services/order && air

dev-payment:
	@cd services/payment && air

dev-all:
	@(cd services/user && air) & \
	(cd services/order && air) & \
	(cd services/payment && air) & \
	(cd gateway && air) & \
	wait

install-air:
	@go install github.com/air-verse/air@latest

# =============================================================================
# TESTING & MIGRATIONS
# =============================================================================

test:
	@cd shared && go test ./...
	@cd services/user && go test ./...
	@cd services/order && go test ./...
	@cd services/payment && go test ./...

test-user:
	@cd services/user && go test -v ./...

test-order:
	@cd services/order && go test -v ./...

test-payment:
	@cd services/payment && go test -v ./...

test-coverage:
	@cd services/user && go test -cover ./...
	@cd services/order && go test -cover ./...
	@cd services/payment && go test -cover ./...

migrate-up:
	@cd services/user && go run cmd/migrate/main.go -action=up
	@cd services/order && go run cmd/migrate/main.go -action=up
	@cd services/payment && go run cmd/migrate/main.go -action=up

migrate-down:
	@cd services/user && go run cmd/migrate/main.go -action=down
	@cd services/order && go run cmd/migrate/main.go -action=down
	@cd services/payment && go run cmd/migrate/main.go -action=down

migrate-status:
	@cd services/user && go run cmd/migrate/main.go -action=version
	@cd services/order && go run cmd/migrate/main.go -action=version
	@cd services/payment && go run cmd/migrate/main.go -action=version

# =============================================================================
# INFRASTRUCTURE & UTILITIES
# =============================================================================

infra:
	@docker compose up -d postgres-user postgres-order postgres-payment rabbitmq redis
	@sleep 5
	@echo "Infrastructure ready!"

infra-down:
	@docker compose stop postgres-user postgres-order postgres-payment rabbitmq redis

health:
	@echo "User Service:"
	@curl -s http://localhost:8081/health || echo "Not running"
	@echo "Order Service:"
	@curl -s http://localhost:8082/health || echo "Not running"
	@echo "Payment Service:"
	@curl -s http://localhost:8083/health || echo "Not running"

deps:
	@cd shared && go mod download
	@cd gateway && go mod download
	@cd services/user && go mod download
	@cd services/order && go mod download
	@cd services/payment && go mod download

tidy:
	@cd shared && go mod tidy
	@cd gateway && go mod tidy
	@cd services/user && go mod tidy
	@cd services/order && go mod tidy
	@cd services/payment && go mod tidy

fmt:
	@cd shared && go fmt ./...
	@cd gateway && go fmt ./...
	@cd services/user && go fmt ./...
	@cd services/order && go fmt ./...
	@cd services/payment && go fmt ./...

lint:
	@which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@cd shared && golangci-lint run ./...
	@cd gateway && golangci-lint run ./...
	@cd services/user && golangci-lint run ./...
	@cd services/order && golangci-lint run ./...
	@cd services/payment && golangci-lint run ./...

reset: clean-all
	@echo "Full reset complete"

# =============================================================================
# NGINX
# =============================================================================

nginx-up:
	@docker compose up -d nginx

nginx-down:
	@docker compose stop nginx

nginx-restart: nginx-down nginx-up

nginx-logs:
	@docker compose logs -f nginx

nginx-reload:
	@docker compose exec nginx nginx -s reload

# =============================================================================
# PERFORMANCE
# =============================================================================

test-performance:
	@./test-performance.sh

test-performance-quick:
	@./test-performance-quick.sh

load-test:
ifndef URL
	@echo "Usage: make load-test URL=http://localhost/ CONCURRENCY=10 REQUESTS=100"
	@exit 1
endif
	@ab -n $(REQUESTS) -c $(CONCURRENCY) "$(URL)"

# =============================================================================
# ALIASES (for convenience)
# =============================================================================

setup: k3s-setup
status: k3s-status
logs: k3s-logs
