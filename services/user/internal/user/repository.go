package user

import (
	"context"

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

	// Hash password before saving
	if err := user.HashPassword(); err != nil {
		log.WithError(err).Error("Failed to hash password")
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to process password")
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		log.WithError(err).Error("Failed to create user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create user")
	}

	log.Info("User created successfully")
	return nil
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
		log.WithError(err).Error("Failed to get user")
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

// ExistsByEmail checks if a user with the given email exists
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.ExistsByEmailWithDB(ctx, r.db, email)
}

// ExistsByEmailWithDB checks if a user with the given email exists using the provided database connection
func (r *Repository) ExistsByEmailWithDB(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	log := logger.WithContext(ctx)

	var count int64
	err := db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		log.WithError(err).Error("Failed to check email existence")
		return false, errors.Wrap(err, errors.ErrDatabaseError, "Failed to check email existence")
	}

	return count > 0, nil
}

// List lists all users with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	log := logger.WithContext(ctx)

	var users []*User
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		log.WithError(err).Error("Failed to list users")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list users")
	}

	return users, nil
}

// Update updates user
func (r *Repository) Update(ctx context.Context, user *User) error {
	return r.UpdateWithDB(ctx, r.db, user)
}

// UpdateWithDB updates user using the provided database connection
func (r *Repository) UpdateWithDB(ctx context.Context, db *gorm.DB, user *User) error {
	log := logger.WithContext(ctx)

	result := db.WithContext(ctx).Save(user)
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

// Delete deletes user
func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DeleteWithDB(ctx, r.db, id)
}

// DeleteWithDB deletes user using the provided database connection
func (r *Repository) DeleteWithDB(ctx context.Context, db *gorm.DB, id string) error {
	log := logger.WithContext(ctx)

	result := db.WithContext(ctx).Delete(&User{}, "id = ?", id)
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

// GetByIDs gets multiple users by IDs
func (r *Repository) GetByIDs(ctx context.Context, ids []string) ([]*User, error) {
	log := logger.WithContext(ctx)

	if len(ids) == 0 {
		return []*User{}, nil
	}

	var users []*User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get users by IDs")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get users by IDs")
	}

	return users, nil
}
