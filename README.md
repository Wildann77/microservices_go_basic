# Microservices Go

A production-ready microservices architecture built with Go, featuring REST APIs, GraphQL Gateway, RabbitMQ for async communication, PostgreSQL databases, and comprehensive observability.

## 🏗️ Architecture

```
┌─────────────────┐
│   Client Apps   │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│  Nginx Reverse      │  ← Port 80, Caching Layer
│  Proxy + Cache      │     (5 min TTL, 100MB)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  GraphQL Gateway    │  ← Port 4000
│  (Authentication,   │
│   Rate Limiting,    │
│   DataLoader)       │
└────────┬────────────┘
         │
    ┌────┴────┬────────┐
    ▼         ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐
│  User  │ │ Order  │ │ Payment│  ← Services
│ Service│ │ Service│ │ Service│
│:8081   │ │:8082   │ │:8083   │
└───┬────┘ └───┬────┘ └───┬────┘
    │          │          │
    ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐
│Postgres│ │Postgres│ │Postgres│
│:5432   │ │:5433   │ │:5434   │
└────────┘ └────────┘ └────────┘
         \        |        /
          \       |       /
           ▼      ▼      ▼
        ┌─────────────────┐
        │    RabbitMQ     │  ← Message Broker
        │   :5672/:15672  │
        └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Make

### 1. Clone and Setup

```bash
git clone <repository-url>
cd microservices-go
cp .env.example .env
```

### 2. Start All Services

```bash
# Build and start all services
make up-d

# Or with logs
make up
```

### 3. Run GraphQL Gateway (Local)

```bash
make run-gateway
```

### 4. Access Services

| Service | URL | Notes |
|---------|-----|-------|
| **Nginx (Recommended)** | http://localhost | Reverse proxy + caching |
| GraphQL Playground (Direct) | http://localhost:4000 | Via Gateway |
| User Service REST | http://localhost:8081/api/v1/users | Direct access |
| Order Service REST | http://localhost:8082/api/v1/orders | Direct access |
| Payment Service REST | http://localhost:8083/api/v1/payments | Direct access |
| Adminer (DB UI) | http://localhost:8080 | Database management |
| RabbitMQ Management | http://localhost:15672 | guest/guest |

| Prometheus | http://localhost:9090 | Metrics |

**💡 Recommendation:** Use Nginx (port 80) for best performance with caching enabled.

## 📁 Project Structure

```
microservices-go/
├── gateway/                    # GraphQL Gateway (runs locally)
│   ├── cmd/main.go
│   ├── graph/
│   │   ├── schema.graphqls    # GraphQL schema definition
│   │   ├── resolver.go        # GraphQL resolvers
│   │   ├── dataloader.go      # N+1 query fix
│   │   └── model/
│   └── middleware/
│
├── services/
│   ├── user/                   # User Service
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── user/          # Feature-based: handler, service, repository, model
│   │   │   └── rabbit/        # Event publisher
│   │   ├── migrations/        # Database migrations
│   │   └── Dockerfile
│   │
│   ├── order/                  # Order Service
│   │   └── ... (same structure)
│   │
│   └── payment/                # Payment Service
│       └── ... (same structure)
│
├── shared/                     # Shared packages
│   ├── logger/                # Zerolog logging
│   ├── config/                # Environment configuration
│   ├── errors/                # Structured error handling
│   ├── middleware/            # Auth, rate limiting, logging
│   ├── validator/             # Input validation
│   ├── tracing/               # OpenTelemetry tracing
│   └── rabbitmq/              # RabbitMQ client
│
├── docker-compose.yml          # Infrastructure & services
├── Makefile                    # Development commands
├── prometheus.yml              # Prometheus configuration
└── README.md
```

## 🔧 Available Commands

```bash
# Build and run
make build          # Build all Docker images
make up             # Start all services
make up-d           # Start all services (detached)
make down           # Stop all services
make restart        # Restart all services
make logs           # View logs
make logs-f         # Follow logs

# Local development
make run-gateway    # Run GraphQL Gateway locally
make run-user       # Run User Service locally
make run-order      # Run Order Service locally
make run-payment    # Run Payment Service locally
make infra          # Start only infrastructure (DB, RabbitMQ)

# Nginx
make nginx-up          # Start Nginx reverse proxy
make nginx-down        # Stop Nginx
make nginx-restart     # Restart Nginx
make nginx-logs        # View Nginx logs
make nginx-clear-cache # Clear Nginx cache
make nginx-status      # Check Nginx status

# Testing
make test           # Run all tests
make test-user      # Run User Service tests
make test-coverage  # Run tests with coverage

# Database
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations

# Code quality
make lint           # Run linter
make fmt            # Format code
make tidy           # Tidy Go modules

# Utilities
make health         # Check service health
make api-docs       # Show API documentation
make clean          # Clean containers
make clean-all      # Clean everything
```

## 📡 GraphQL API Examples

### Register User

```graphql
mutation {
  register(input: {
    email: "user@example.com"
    password: "password123"
    firstName: "John"
    lastName: "Doe"
  }) {
    token
    user {
      id
      email
      fullName
    }
  }
}
```

### Login

```graphql
mutation {
  login(input: {
    email: "user@example.com"
    password: "password123"
  }) {
    token
    user {
      id
      email
    }
  }
}
```

### Create Order (Authenticated)

```graphql
mutation {
  createOrder(input: {
    shippingAddress: "123 Main St, City, Country"
    notes: "Please handle with care"
    items: [
      {
        productId: "prod-1"
        productName: "Laptop"
        quantity: 1
        unitPrice: 999.99
      }
    ]
  }) {
    id
    status
    totalAmount
    items {
      productName
      quantity
      unitPrice
    }
  }
}
```

### Get Orders with User and Payment (N+1 Fixed)

```graphql
query {
  orders(limit: 10) {
    data {
      id
      status
      totalAmount
      user {
        id
        email
        fullName
      }
      payment {
        id
        status
        amount
      }
    }
    pageInfo {
      total
      hasMore
    }
  }
}
```

### Get Current User with Orders

```graphql
query {
  me {
    id
    email
    firstName
    lastName
    orders {
      id
      status
      totalAmount
      items {
        productName
        quantity
      }
    }
  }
}
```

## 🔌 REST API Endpoints

### User Service (Port 8081)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/users/register` | Register new user | No |
| POST | `/api/v1/users/login` | Login user | No |
| GET | `/api/v1/users` | List users | Yes |
| GET | `/api/v1/users/:id` | Get user by ID | Yes |
| GET | `/api/v1/users/me` | Get current user | Yes |
| PUT | `/api/v1/users/:id` | Update user | Yes |
| DELETE | `/api/v1/users/:id` | Delete user | Yes |
| GET | `/health` | Health check | No |

### Order Service (Port 8082)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/orders` | Create order | Yes |
| GET | `/api/v1/orders` | List orders | Yes |
| GET | `/api/v1/orders/my-orders` | Get my orders | Yes |
| GET | `/api/v1/orders/:id` | Get order by ID | Yes |
| PATCH | `/api/v1/orders/:id/status` | Update order status | Yes |
| GET | `/health` | Health check | No |

### Payment Service (Port 8083)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/payments` | Create payment | Yes |
| GET | `/api/v1/payments` | List payments | Yes |
| GET | `/api/v1/payments/my-payments` | Get my payments | Yes |
| GET | `/api/v1/payments/:id` | Get payment by ID | Yes |
| GET | `/api/v1/payments/order/:orderId` | Get payment by order | Yes |
| POST | `/api/v1/payments/:id/process` | Process payment | Yes |
| POST | `/api/v1/payments/:id/refund` | Refund payment | Yes |
| GET | `/health` | Health check | No |

## 🏛️ Architecture Patterns

### 1. Feature-Based Structure

Each service follows feature-based organization:

```
service/
├── cmd/main.go              # Entry point
├── internal/
│   ├── feature/             # Feature package
│   │   ├── handler.go       # HTTP handlers
│   │   ├── service.go       # Business logic
│   │   ├── repository.go    # Data access
│   │   ├── model.go         # Domain models
│   │   └── validator.go     # Input validation
│   └── rabbit/              # Event handling
├── migrations/              # Database migrations
└── Dockerfile
```

### 2. Database Per Service

- **User Service**: PostgreSQL on port 5432
- **Order Service**: PostgreSQL on port 5433
- **Payment Service**: PostgreSQL on port 5434

### 3. Async Communication via RabbitMQ

Events published:
- `user.created` - When a new user registers
- `order.created` - When a new order is placed
- `order.status_changed` - When order status updates
- `payment.success` - When payment is successful
- `payment.failed` - When payment fails

### 4. ACID Transactions

Each service handles its own transactions:

```go
// Repository pattern with transactions
func (r *Repository) Create(ctx context.Context, order *Order) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Insert order
    // Insert order items
    // Commit transaction
    return tx.Commit()
}
```

### 5. Structured Error Handling

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Validation failed",
    "details": "email: Invalid email format"
  }
}
```

### 6. Security

- JWT authentication at Gateway
- Rate limiting (100 req/s default)
- Security headers (CSP, HSTS, X-Frame-Options)
- CORS configuration
- SQL injection safe (parameterized queries)

### 7. Standardized Responses

All services use a standardized JSON response format via the `shared/response` package:

```go
// Success Response
return response.Success(w, data, "Operation successful")

// Error Response
return response.Error(w, errors.ErrNotFound, "Resource not found")
```

Structure:
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "meta": { ... }
}
```

### 8. Soft Delete

Implemented using GORM's `DeletedAt`:

- **User & Order Services**: Records are marked as deleted (timestamped) rather than physically removed.
- **Data Safety**: Prevents accidental data loss and maintains audit history.
- ** Automatic Filtering**: GORM automatically excludes soft-deleted records from standard queries.

### 9. Observability

- **Logging**: Zerolog with structured JSON
- **Health Checks**: `/health` endpoint on each service
- **Tracing**: OpenTelemetry (prepared)

**Note**: Prometheus container running (port 9090) but not actively collecting metrics - to be implemented later.

### 10. Nginx Reverse Proxy with Caching

Nginx serves as the entry point for all client requests, providing:

- **Reverse Proxy**: Routes requests to GraphQL Gateway
- **Caching Layer**: 5-minute cache TTL for GraphQL queries
- **Rate Limiting**: Additional protection (1000 req/min backup)
- **Static Assets**: 1-year cache for CSS/JS files
- **Security Headers**: X-Frame-Options, X-Content-Type-Options

**Cache Headers:**
```
X-Cache-Status: HIT  ← Served from cache (< 1ms)
X-Cache-Status: MISS ← Fetched from gateway
```

**Performance Improvement:**
- Without Nginx: ~15ms response time, 3,240 RPS
- With Nginx: ~4ms response time, 11,878 RPS
- **3.7x faster with caching!** 🚀

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific service tests
make test-user
make test-order
make test-payment

# Run with coverage
make test-coverage
```

## 🚀 Performance Testing

Test your API performance with built-in benchmarking tools.

### Quick Performance Report

```bash
make test-performance-quick
```

**Sample Output:**
```
1. SERVICE STATUS
-----------------
User Service: healthy
Order Service: healthy
Payment Service: healthy

2. RESPONSE TIME TEST
---------------------
User Health:     18ms average
Order Health:    18ms average
Payment Health:  20ms average
Gateway:         15ms average
Nginx:           16ms average

3. NGINX CACHING PERFORMANCE
----------------------------
First request (cache MISS):  0.81s
Second request (cache HIT):  1.10ms
Improvement: -35% faster
```

### Full Performance Test Suite

```bash
make test-performance
```

Runs comprehensive tests including:
- Single request latency
- Cache performance comparison
- Load tests (100-500 concurrent requests)
- Concurrent user simulation

### Load Test Specific Endpoint

```bash
# Test any endpoint
make load-test URL=http://localhost/ CONCURRENCY=50 REQUESTS=500
```

### Performance Results

See detailed results in `PERFORMANCE-REPORT.md`:

```bash
make test-performance-report
```

**Key Metrics:**

| Metric | Value | Status |
|--------|-------|--------|
| Response Time | 1-3ms average | ✅ Excellent |
| Throughput | 3,000-12,000 RPS | ✅ Excellent |
| Cache Hit Rate | ~90% | ✅ Good |
| Failed Requests | < 1% | ✅ Excellent |

**Performance by Service:**

| Service | RPS | Latency |
|---------|-----|---------|
| User Service | 5,252 | 1.90ms |
| Order Service | 3,159 | 3.17ms |
| Payment Service | ~3,159 | ~3.17ms |
| Gateway (Direct) | 3,240 | 15.43ms |
| Gateway + Nginx | 11,878 | 4.21ms |

**Architecture Comparison:**

```
Without Nginx (Direct Gateway):
  3,240 RPS, 15ms latency

With Nginx (Cached):
  11,878 RPS, 4ms latency
  🚀 3.7x improvement!
```

## 🚢 Deployment

### Docker Compose (Development)

```bash
make up-d
```

### Production Considerations

1. **Environment Variables**: Set production values in `.env`
2. **Secrets**: Use Docker secrets or external secret management
3. **SSL/TLS**: Enable HTTPS with reverse proxy (nginx/traefik)
4. **Scaling**: Use Docker Swarm or Kubernetes
5. **Monitoring**: Set up alerts in Prometheus/Grafana
6. **Backups**: Configure database backups

## ☸️ Kubernetes (k3s) Deployment

This project includes Kubernetes manifests for deploying to k3s, a lightweight Kubernetes distribution.

### Prerequisites

- k3s installed and running
- kubectl configured

### Quick Start

```bash
# Setup (verify k3s and create namespace)
make k3s-setup

# Build images and load into k3s
make k3s-build
make k3s-import

# Deploy to k3s
make k3s-deploy
```

### Available Commands

| Command | Description |
|---------|-------------|
| `make k3s-setup` | Verify k3s and create namespace |
| `make k3s-build` | Build and load images into k3s |
| `make k3s-deploy` | Deploy to k3s |
| `make k3s-up` | Build and deploy (one command) |
| `make k3s-status` | Check pod status |
| `make k3s-logs SERVICE=gateway` | View service logs |
| `make k3s-redeploy` | Redeploy services |
| `make k3s-clean` | Delete namespace (cleanup) |
| `make k3s-test` | Run API tests |

### Access URLs (After Deployment)

**NodePort (Direct):**
- Gateway: http://localhost:30080
- Health: http://localhost:30080/health
- GraphQL: http://localhost:30080/query

**Ingress (without host):**
- Gateway: http://localhost
- Health: http://localhost/health
- GraphQL: http://localhost/query

**Ingress (with host):**
- Add `127.0.0.1 microservices.local` to /etc/hosts
- Gateway: http://microservices.local

For more details, see [k8s/README.md](k8s/README.md).

## 📊 Monitoring

| Tool | URL | Purpose |
|------|-----|---------|
| Nginx | http://localhost/health | Reverse proxy health |
| Adminer | http://localhost:8080 | Database management UI |
| RabbitMQ Management | http://localhost:15672 | Message queue monitoring |
| Prometheus | http://localhost:9090 | Metrics collection (Zombie) |

### Connecting to Databases via Adminer

When using Adminer (http://localhost:8080) to manage databases, use the following connection details:

- **System**: `PostgreSQL`
- **Server**: (Use the service name from docker-compose)
  - `postgres-user` (for User Service)
  - `postgres-order` (for Order Service)
  - `postgres-payment` (for Payment Service)
- **Username**: `postgres` (or as defined in your `.env`)
- **Password**: `password` (or as defined in your `.env`)
- **Database**: `user`, `order`, or `payment` (corresponding to the server)

*Note: Use port `5432` (default) within the Docker network.*

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Web Framework | Chi Router |
| GraphQL | gqlgen |
| Database | PostgreSQL 15 |
| Message Broker | RabbitMQ |
| Reverse Proxy | Nginx |
| Authentication | JWT |
| Logging | Zerolog |
| Configuration | godotenv (Auto-loading) |
| Metrics | Prometheus (Zombie) |
| Validation | go-playground/validator |
| Testing | Go testing + testify |
| Performance Testing | Apache Bench (ab) |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## 📝 License

MIT License - see LICENSE file for details.

## 📧 Support

For questions or issues, please open a GitHub issue or contact the maintainers.
