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
