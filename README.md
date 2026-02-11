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
│  GraphQL Gateway    │  ← Runs locally (port 4000)
│  (Authentication,   │
│   Rate Limiting,    │
│   DataLoader)       │
└────────┬────────────┘
         │
    ┌────┴────┬────────┐
    ▼         ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐
│  User  │ │ Order  │ │ Payment│  ← Docker containers
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

| Service | URL |
|---------|-----|
| GraphQL Playground | http://localhost:4000 |
| User Service REST | http://localhost:8081/api/v1/users |
| Order Service REST | http://localhost:8082/api/v1/orders |
| Payment Service REST | http://localhost:8083/api/v1/payments |
| RabbitMQ Management | http://localhost:15672 (guest/guest) |
| Jaeger UI | http://localhost:16686 |
| Prometheus | http://localhost:9090 |

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

### 7. Observability

- **Logging**: Zerolog with structured JSON
- **Tracing**: OpenTelemetry with Jaeger
- **Metrics**: Prometheus
- **Health Checks**: `/health` endpoint on each service

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

## 📊 Monitoring

| Tool | URL | Purpose |
|------|-----|---------|
| RabbitMQ Management | http://localhost:15672 | Message queue monitoring |
| Jaeger | http://localhost:16686 | Distributed tracing |
| Prometheus | http://localhost:9090 | Metrics collection |

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Web Framework | Chi Router |
| GraphQL | gqlgen |
| Database | PostgreSQL 15 |
| Message Broker | RabbitMQ |
| Authentication | JWT |
| Logging | Zerolog |
| Tracing | OpenTelemetry + Jaeger |
| Metrics | Prometheus |
| Validation | go-playground/validator |
| Testing | Go testing + testify |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## 📝 License

MIT License - see LICENSE file for details

## 📧 Support

For questions or issues, please open a GitHub issue or contact the maintainers.