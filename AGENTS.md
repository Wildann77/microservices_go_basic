# AGENTS.md - Coding Guidelines for Microservices Go

## Build/Test/Lint Commands

### Testing
```bash
# Run all tests
make test

# Run tests for specific service
make test-user      # User service
make test-order     # Order service
make test-payment   # Payment service

# Run a single test (example)
cd services/user && go test -v -run TestCreateUser ./internal/user/...

# Run tests with coverage
cd services/<service> && go test -cover ./...
```

### Linting & Formatting
```bash
# Run linter on all services
make lint

# Format all code
make fmt

# Tidy Go modules
make tidy
```

### Local Development
```bash
# Start infrastructure (databases, RabbitMQ, etc.)
make infra

# Run services locally with hot reload (requires Air)
make dev-user       # User service
make dev-order      # Order service
make dev-payment    # Payment service
make dev-gateway    # Gateway
make dev-all        # Run all services with hot reload simultaneously

# Run services locally (in separate terminals, without hot reload)
make run-user
make run-order
make run-payment
make run-gateway

# Full Docker setup
make up-d
make down

# Install tools
make install-air    # Install Air for hot reload
```

### Database Migrations

This project uses **golang-migrate** for SQL-based database migrations with both auto-migration on service start and manual CLI tools.

#### Auto-Migration (Default)
Migrations run automatically when services start:
```bash
make run-order    # Migrations run before service starts
make run-user
make run-payment
```

**Environment Variables** (add to `.env`):
```env
# Enable/disable auto-migration (default: true)
USER_AUTO_MIGRATE=true
ORDER_AUTO_MIGRATE=true
PAYMENT_AUTO_MIGRATE=true
```

#### Manual Migration Commands

**Via Makefile** (recommended):
```bash
# Run all pending migrations for all services
make migrate-up

# Rollback one migration for all services
make migrate-down

# Check migration status for all services
make migrate-status

# Service-specific migrations
make migrate-user-up
make migrate-user-down
make migrate-user-status

make migrate-order-up
make migrate-order-down
make migrate-order-status

make migrate-payment-up
make migrate-payment-down
make migrate-payment-status
```

**Via CLI directly**:
```bash
cd services/order

# Run all pending migrations
go run cmd/migrate/main.go -action=up

# Rollback one migration
go run cmd/migrate/main.go -action=down

# Check current version
go run cmd/migrate/main.go -action=version

# Check detailed status
go run cmd/migrate/main.go -action=status

# Force specific version (use with caution!)
go run cmd/migrate/main.go -action=force -force=1
```

#### Creating New Migrations

**Naming convention**: `NNN_description.up.sql` and `NNN_description.down.sql`

Create files in `services/<name>/migrations/`:
```bash
services/order/migrations/
├── 001_create_orders.up.sql
├── 001_create_orders.down.sql
├── 002_add_indexes.up.sql
└── 002_add_indexes.down.sql
```

**Example migration file** (`002_add_indexes.up.sql`):
```sql
-- Create new indexes
CREATE INDEX IF NOT EXISTS idx_orders_new_field ON orders (new_field);

-- Add new column
ALTER TABLE orders ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);
```

**Corresponding down file** (`002_add_indexes.down.sql`):
```sql
-- Remove indexes
DROP INDEX IF EXISTS idx_orders_new_field;

-- Remove column
ALTER TABLE orders DROP COLUMN IF EXISTS new_field;
```

#### Migration Best Practices

1. **Always create both `.up.sql` and `.down.sql` files**
2. **Make migrations idempotent** using `IF EXISTS` / `IF NOT EXISTS`
3. **Test down migrations** before committing
4. **Never modify existing migration files** after they've been applied
5. **One logical change per migration file**
6. **Keep migrations small and focused**

#### Migration Status Codes

- **Version**: Current migration version number
- **Dirty**: `true` if migration failed mid-way (requires manual fix)
- **No Change**: Database is already up to date

## Code Style Guidelines

### Project Structure
- Each service is a separate Go module in `services/<name>/`
- Gateway is in `gateway/`
- Shared code is in `shared/` (imported via replace directive)
- Standard internal structure: `internal/<domain>/handler.go, service.go, repository.go, model.go`

### Imports Order
1. Standard library imports
2. Third-party imports (blank line separator)
3. Internal/shared imports (blank line separator)

Example:
```go
import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
)
```

### Naming Conventions
- **Modules**: Each service is a separate Go module.
- **Packages**: lowercase, single word (e.g., `user`, `order`)
- **Files**: snake_case (e.g., `handler.go`, `stripe_provider.go`)
- **Types**: PascalCase (e.g., `UserService`, `CreateUserRequest`)
- **Interfaces**: PascalCase with descriptive names (e.g., `EventPublisher`)
- **Variables**: camelCase (e.g., `userRepo`, `jwtConfig`)
- **Constants**: PascalCase for exported, camelCase for private
- **GORM/DB tags**: snake_case (e.g., `gorm:"column:created_at"`)
- **JSON tags**: snake_case (e.g., `json:"first_name"`)

### Error Handling
Use the shared errors package:
```go
import "github.com/microservices-go/shared/errors"

// Create new error
return nil, errors.New(errors.ErrNotFound, "User not found")

// Wrap existing error
return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to fetch user")

// Check error type
if errors.IsNotFound(err) { ... }
```

Error codes: `ErrInvalidInput`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`, `ErrValidationFailed`, `ErrInternalServer`, `ErrDatabaseError`

### Logging
Use the shared logger (zerolog wrapper):
```go
import "github.com/microservices-go/shared/logger"

// Create service logger
log := logger.New("user-service")

// With context for trace ID
log := logger.WithContext(ctx)
log.WithError(err).Warn("Failed to publish event")

// Levels: Debug(), Info(), Warn(), Error(), Fatal()
```

### Struct Tags (JSON, GORM, & Validation)
```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey;column:id"`
    Email     string    `json:"email" gorm:"unique;column:email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}
```

### Interface Design
Define interfaces in the package that uses them:
```go
// In service.go
type EventPublisher interface {
    PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}
```

### Context Usage
Always pass context as first parameter:
```go
func (s *Service) GetByID(ctx context.Context, id string) (*UserResponse, error)
```

### Model Methods
```go
// Conversion method
func (u *User) ToResponse() *UserResponse

// Lifecycle hooks
func (u *User) BeforeCreate()
func (u *User) BeforeUpdate()

// Domain logic
func (u *User) CheckPassword(password string) bool
```

### Validation
Use shared validator with struct tags:
```go
validator := validator.New()
if err := validator.ValidateStruct(req); err != nil {
    return nil, err
}
```

### Environment Variables
- Use service-specific prefixes (e.g., `USER_PORT`, `ORDER_DB_HOST`)
- **Auto-loading**: Proyek ini menggunakan `godotenv` yang secara otomatis memuat file `.env` dari root folder saat inisialisasi paket `shared/config`.
- Tidak perlu memanggil `godotenv.Load()` manual di setiap `main.go` karena sudah ditangani di level `shared`.

### Module Imports
Each service imports shared via replace:
```go
replace github.com/microservices-go/shared => ../shared
```

### Testing Standards
- No test files currently exist - create `*_test.go` files in same package
- Use `testing` package from standard library
- Mock interfaces for unit tests
- Table-driven tests preferred

## Rate Limiting

This project uses **Redis-based rate limiting** with Fixed Window algorithm across all services.

### Architecture
- **Gateway**: Global rate limit (1000 req/min per IP) - entry point protection
- **Services**: Service-specific limits for granular control:
  - **User Service**: 500 req/min
  - **Order Service**: 200 req/min
  - **Payment Service**: 100 req/min

### Configuration

Add to your `.env` file:
```env
# Redis Configuration (required for rate limiting)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiting Configuration (requests per minute)
GATEWAY_RATE_LIMIT_RPM=1000
GATEWAY_RATE_LIMIT_WINDOW=60

USER_RATE_LIMIT_RPM=500
USER_RATE_LIMIT_WINDOW=60

ORDER_RATE_LIMIT_RPM=200
ORDER_RATE_LIMIT_WINDOW=60

PAYMENT_RATE_LIMIT_RPM=100
PAYMENT_RATE_LIMIT_WINDOW=60
```

### Docker Compose
Redis is automatically started with `make infra` or `make up-d`:
```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
```

### Usage in Code

The rate limiter is automatically initialized in each service's `main.go`:

```go
// Load Redis config
redisConfig := config.LoadRedisConfig()

// Connect to Redis
redisClient, err := redis.NewClient(redisConfig)
if err != nil {
    log.Warn("Failed to connect to Redis: " + err.Error())
    log.Warn("Rate limiting will be disabled")
} else {
    defer redisClient.Close()
}

// Load rate limit config
rateLimitConfig := config.LoadRateLimitConfig("service-name")

// Create rate limiter
var rateLimiter *middleware.RateLimiter
if redisClient != nil {
    rateLimiter = middleware.NewRateLimiter(redisClient, rateLimitConfig, "service-name")
}

// Apply middleware
if rateLimiter != nil {
    r.Use(rateLimiter.Middleware)
}
```

### Graceful Fallback
If Redis is unavailable, services will:
- Log a warning
- Continue operating without rate limiting (fail-open strategy)
- Not block legitimate traffic

### Rate Limit Headers
When rate limiting is active, responses include:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window

### Response on Rate Limit Exceeded
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests, please try again later"
  }
}
```

### Per-User Rate Limiting
For authenticated endpoints, use per-user rate limiting:
```go
r.Use(rateLimiter.PerUserMiddleware)
```

This limits requests per user ID instead of per IP address.

## Caching

This project uses **Redis-based caching** at the Service Layer to improve performance for frequently accessed data.

### Architecture

Caching is implemented at the **Service Layer** for optimal performance:
- **Read Operations**: Cache-Aside pattern (check cache first, fallback to DB)
- **Write Operations**: Cache invalidation on create/update/delete
- **TTL per Service**: Configurable time-to-live for cached data

### Cached Endpoints by Service

| Service | Cached Endpoints | TTL | Cache Keys |
|---------|-----------------|-----|------------|
| **User** | GetByID, GetByEmail, List | 5 min | `user:id:{id}`, `user:email:{email}`, `users:list:*` |
| **Order** | GetByID, GetByUserID | 3 min | `order:id:{id}`, `orders:user:{userID}:*` |
| **Payment** | GetByID, GetByOrderID, GetByUserID | 3 min | `payment:id:{id}`, `payment:order:{orderID}`, `payments:user:{userID}:*` |

### Configuration

Add to your `.env` file:
```env
# Cache Configuration (in seconds)
# User Service: 300 seconds (5 minutes)
USER_CACHE_TTL=300

# Order Service: 180 seconds (3 minutes)
ORDER_CACHE_TTL=180

# Payment Service: 180 seconds (3 minutes)
PAYMENT_CACHE_TTL=180

# Enable/disable caching per service
CACHE_ENABLED=true
```

### Usage in Code

The cache is automatically initialized in each service's `main.go`:

```go
// Create cache client
var cacheClient *cache.Cache
if redisClient != nil {
    cacheClient = cache.NewCache(redisClient.GetClient(), "service-name")
}

// Inject into service
userService := user.NewService(userRepo, jwtConfig, publisher, cacheClient)
```

### Cache Operations in Services

```go
// Get with caching
func (s *Service) GetByID(ctx context.Context, id string) (*Response, error) {
    cacheKey := "entity:id:" + id
    
    // Try cache first
    if s.cache != nil {
        var cached Response
        if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
            return &cached, nil
        }
    }
    
    // Get from database
    entity, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    response := entity.ToResponse()
    
    // Store in cache
    if s.cache != nil {
        s.cache.Set(ctx, cacheKey, response, s.cacheTTL)
    }
    
    return response, nil
}

// Invalidate cache on update
func (s *Service) Update(ctx context.Context, id string, req *Request) (*Response, error) {
    // ... update logic ...
    
    // Invalidate caches
    if s.cache != nil {
        s.cache.Delete(ctx, "entity:id:"+id)
        s.cache.DeletePattern(ctx, "entities:list:*")
    }
    
    return response, nil
}
```

### Cache Key Patterns

- **Single Entity**: `{entity}:id:{id}` (e.g., `user:id:123`)
- **By Secondary Key**: `{entity}:{field}:{value}` (e.g., `user:email:john@example.com`)
- **List/Collection**: `{entities}:list:*` or `{entities}:user:{userID}:*`

### Cache Invalidation Strategy

1. **Write-through on Create**: Invalidate list caches
2. **Write-through on Update**: Invalidate specific entity + list caches
3. **Write-through on Delete**: Invalidate specific entity + list caches
4. **TTL Expiration**: Automatic cleanup of stale data

### Graceful Degradation

If Redis is unavailable:
- Services continue to work (fallback to database)
- Cache operations are skipped silently
- No impact on functionality, only performance

### Performance Benefits

- **User Service**: ~80% reduction in DB reads for user lookups
- **Order Service**: ~70% reduction in DB reads for order history
- **Payment Service**: ~75% reduction in DB reads for payment status checks
