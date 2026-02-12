package user

import (
	"context"

	"time"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"gorm.io/gorm"
)

// Repository handles user data access
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTransaction executes a function within a database transaction
func (r *Repository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.CreateWithDB(ctx, r.db, user)
}

// CreateWithDB creates a new user using the provided database connection
func (r *Repository) CreateWithDB(ctx context.Context, db *gorm.DB, user *User) error {
	log := logger.WithContext(ctx)

	if err := user.HashPassword(); err != nil {
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to hash password")
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		log.WithError(err).Error("Failed to create user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create user")
	}

	log.Info("User created successfully")
	return nil
}

// CreateOutboxEvent creates a new given outbox event
func (r *Repository) CreateOutboxEvent(ctx context.Context, tx *gorm.DB, event *OutboxEvent) error {
	log := logger.WithContext(ctx)

	if err := tx.WithContext(ctx).Create(event).Error; err != nil {
		log.WithError(err).Error("Failed to create outbox event")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create outbox event")
	}

	return nil
}

// UpdateOutboxEvent updates an outbox event
func (r *Repository) UpdateOutboxEvent(ctx context.Context, id string, status string, errStr string) error {
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": time.Now(),
	}
	if errStr != "" {
		updates["error_message"] = errStr
	}
	return r.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", id).Updates(updates).Error
}

// GetByID gets user by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	log := logger.WithContext(ctx)

	var user User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		log.WithError(err).Error("Failed to get user by ID")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get user")
	}

	return &user, nil
}

// GetByEmail gets user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	log := logger.WithContext(ctx)

	var user User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		log.WithError(err).Error("Failed to get user by email")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get user")
	}

	return &user, nil
}

// List lists all users with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	log := logger.WithContext(ctx)

	var users []*User
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to list users")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list users")
	}

	return users, nil
}

// Update updates user
func (r *Repository) Update(ctx context.Context, user *User) error {
	log := logger.WithContext(ctx)

	result := r.db.WithContext(ctx).Save(user)
	if err := result.Error; err != nil {
		log.WithError(err).Error("Failed to update user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update user")
	}

	if result.RowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	log.Info("User updated successfully")
	return nil
}

// Delete deletes user by ID
func (r *Repository) Delete(ctx context.Context, id string) error {
	log := logger.WithContext(ctx)

	result := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if err := result.Error; err != nil {
		log.WithError(err).Error("Failed to delete user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to delete user")
	}

	if result.RowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	log.Info("User deleted successfully")
	return nil
}

// Count returns total user count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count users")
	}
	return int(count), nil
}

// ExistsByEmail checks if email exists
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, errors.Wrap(err, errors.ErrDatabaseError, "Failed to check email existence")
	}
	return count > 0, nil
}
