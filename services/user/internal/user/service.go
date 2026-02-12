package user

import (
	"context"

	"encoding/json"

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
	validator *validator.Validator
	jwtConfig *config.JWTConfig
	publisher EventPublisher
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

// NewService creates a new user service
func NewService(repo *Repository, jwtConfig *config.JWTConfig, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		validator: validator.New(),
		jwtConfig: jwtConfig,
		publisher: publisher,
	}
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, req *CreateUserRequest) (*LoginResponse, error) {
	log := logger.WithContext(ctx)

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Check if email exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.ErrConflict, "Email already exists")
	}

	// Create user
	user := &User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Transaction with Outbox Pattern
	var outboxEvent *OutboxEvent

	if err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Create user
		if err := s.repo.CreateWithDB(ctx, tx, user); err != nil {
			return err
		}

		// Create Outbox Event
		eventPayload := &UserCreatedEvent{
			UserID:    user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return errors.Wrap(err, errors.ErrInternalServer, "Failed to marshal event payload")
		}

		outboxEvent = &OutboxEvent{
			AggregateType: "User",
			AggregateID:   user.ID,
			Type:          "user.created",
			Payload:       payloadBytes,
			Status:        "PENDING",
		}

		if err := s.repo.CreateOutboxEvent(ctx, tx, outboxEvent); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Publish event asynchronously (Best Effort)
	if s.publisher != nil && outboxEvent != nil {
		go func(eventID string, payloadBytes []byte) {
			// Create a background context
			bgCtx := context.Background()

			// Re-unmarshal to pass to publisher interface which expects interface{}
			// Or better, just pass the struct we created if we had it, but payloadBytes is what we have.
			// Actually publisher expects interface{}.
			var eventStruct UserCreatedEvent
			if err := json.Unmarshal(payloadBytes, &eventStruct); err != nil {
				logger.WithContext(bgCtx).WithError(err).Error("Failed to unmarshal event for publishing")
				return
			}

			if err := s.publisher.PublishEvent(bgCtx, "user.created", &eventStruct); err != nil {
				logger.WithContext(bgCtx).WithError(err).Warn("Failed to publish user created event from outbox")
				// Update retry count or status to FAILED in a real system
				_ = s.repo.UpdateOutboxEvent(bgCtx, eventID, "FAILED", err.Error())
			} else {
				// Mark as processed
				_ = s.repo.UpdateOutboxEvent(bgCtx, eventID, "PROCESSED", "")
			}
		}(outboxEvent.ID, outboxEvent.Payload)
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

// GetByID gets user by ID
func (s *Service) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

// GetByEmail gets user by email
func (s *Service) GetByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

// List lists users with pagination
func (s *Service) List(ctx context.Context, limit, offset int) ([]*UserResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
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

	return user.ToResponse(), nil
}

// Delete deletes user
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
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
