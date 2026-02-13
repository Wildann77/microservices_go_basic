package user

import (
	"context"
	"time"

	"github.com/microservices-go/shared/cache"
	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
	"github.com/microservices-go/shared/validator"
	"gorm.io/gorm"
)

// Service handles user business logic
type Service struct {
	repo      *Repository
	cache     *cache.Cache
	validator *validator.Validator
	jwtConfig *config.JWTConfig
	publisher EventPublisher
	cacheTTL  time.Duration
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

// NewService creates a new user service
func NewService(repo *Repository, jwtConfig *config.JWTConfig, publisher EventPublisher, cacheClient *cache.Cache) *Service {
	return &Service{
		repo:      repo,
		cache:     cacheClient,
		validator: validator.New(),
		jwtConfig: jwtConfig,
		publisher: publisher,
		cacheTTL:  5 * time.Minute,
	}
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) (*LoginResponse, error) {
	log := logger.WithContext(ctx)

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Create user object
	user := &User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Create user transactionally
	if err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Check if email exists
		exists, err := s.repo.ExistsByEmailWithDB(ctx, tx, req.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New(errors.ErrConflict, "Email already exists")
		}

		// Create user
		if err := s.repo.CreateWithDB(ctx, tx, user); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Invalidate list caches
	if s.cache != nil {
		if err := s.cache.DeletePattern(ctx, "users:list:*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate users list cache")
		}
	}

	// Publish event
	if s.publisher != nil {
		event := &UserCreatedEvent{
			UserID:    user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		}
		if err := s.publisher.PublishEvent(ctx, "user.created", event); err != nil {
			log.WithError(err).Warn("Failed to publish user created event")
		}
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, s.jwtConfig)
	if err != nil {
		log.WithError(err).Warn("Failed to generate token for new user")
		return nil, errors.Wrap(err, errors.ErrInternalServer, "Failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetByID gets user by ID with caching
func (s *Service) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	cacheKey := "user:id:" + id

	// Try to get from cache
	if s.cache != nil {
		var cachedUser UserResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// Get from database
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()

	// Store in cache
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, response, s.cacheTTL); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to cache user")
		}
	}

	return response, nil
}

// GetByEmail gets user by email with caching
func (s *Service) GetByEmail(ctx context.Context, email string) (*UserResponse, error) {
	cacheKey := "user:email:" + email

	// Try to get from cache
	if s.cache != nil {
		var cachedUser UserResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// Get from database
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()

	// Store in cache
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, response, s.cacheTTL); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to cache user by email")
		}
	}

	return response, nil
}

// List lists users with pagination and caching
func (s *Service) List(ctx context.Context, limit, offset int) ([]*UserResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	cacheKey := "users:list:limit:" + string(rune(limit)) + ":offset:" + string(rune(offset))

	// Try to get from cache
	if s.cache != nil {
		var cachedUsers []*UserResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedUsers); err == nil {
			return cachedUsers, nil
		}
	}

	// Get from database
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	// Store in cache
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, responses, s.cacheTTL); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to cache users list")
		}
	}

	return responses, nil
}

// Update updates user
func (s *Service) Update(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Invalidate caches
	if s.cache != nil {
		log := logger.WithContext(ctx)
		if err := s.cache.Delete(ctx, "user:id:"+id); err != nil {
			log.WithError(err).Warn("Failed to invalidate user cache")
		}
		if err := s.cache.Delete(ctx, "user:email:"+user.Email); err != nil {
			log.WithError(err).Warn("Failed to invalidate user email cache")
		}
		if err := s.cache.DeletePattern(ctx, "users:list:*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate users list cache")
		}
	}

	return user.ToResponse(), nil
}

// Delete deletes user and invalidates cache
func (s *Service) Delete(ctx context.Context, id string) error {
	// Get user email for cache invalidation
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	if s.cache != nil {
		log := logger.WithContext(ctx)
		if err := s.cache.Delete(ctx, "user:id:"+id); err != nil {
			log.WithError(err).Warn("Failed to invalidate user cache")
		}
		if err := s.cache.Delete(ctx, "user:email:"+user.Email); err != nil {
			log.WithError(err).Warn("Failed to invalidate user email cache")
		}
		if err := s.cache.DeletePattern(ctx, "users:list:*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate users list cache")
		}
	}

	return nil
}

// Login authenticates user and returns token
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.New(errors.ErrUnauthorized, "Invalid email or password")
		}
		return nil, err
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.New(errors.ErrUnauthorized, "Invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New(errors.ErrForbidden, "Account is deactivated")
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, s.jwtConfig)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer, "Failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// Count returns total user count
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// GetByIDs gets multiple users by IDs with caching
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*UserResponse, error) {
	if len(ids) == 0 {
		return []*UserResponse{}, nil
	}

	// Get from database
	users, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()

		// Cache individual users
		if s.cache != nil {
			cacheKey := "user:id:" + user.ID
			if err := s.cache.Set(ctx, cacheKey, responses[i], s.cacheTTL); err != nil {
				logger.WithContext(ctx).WithError(err).Warn("Failed to cache user")
			}
		}
	}

	return responses, nil
}
