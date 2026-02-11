package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
)

// Repository handles user data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user with transaction
func (r *Repository) Create(ctx context.Context, user *User) error {
	log := logger.WithContext(ctx)
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	user.BeforeCreate()
	if err := user.HashPassword(); err != nil {
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to hash password")
	}

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.FirstName, user.LastName,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		log.WithError(err).Error("Failed to create user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create user")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("User created successfully")
	return nil
}

// GetByID gets user by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	log := logger.WithContext(ctx)
	
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		log.WithError(err).Error("Failed to get user by ID")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get user")
	}

	return &user, nil
}

// GetByEmail gets user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	log := logger.WithContext(ctx)
	
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		log.WithError(err).Error("Failed to get user by email")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get user")
	}

	return &user, nil
}

// List lists all users with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	log := logger.WithContext(ctx)
	
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to list users")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list users")
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Failed to scan user")
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}

// Update updates user with transaction
func (r *Repository) Update(ctx context.Context, user *User) error {
	log := logger.WithContext(ctx)
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	user.BeforeUpdate()

	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := tx.ExecContext(ctx, query,
		user.FirstName, user.LastName, user.IsActive, user.UpdatedAt, user.ID,
	)
	if err != nil {
		log.WithError(err).Error("Failed to update user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update user")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("User updated successfully")
	return nil
}

// Delete deletes user by ID
func (r *Repository) Delete(ctx context.Context, id string) error {
	log := logger.WithContext(ctx)
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	query := `DELETE FROM users WHERE id = $1`
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).Error("Failed to delete user")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to delete user")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("User deleted successfully")
	return nil
}

// Count returns total user count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count users")
	}
	return count, nil
}

// ExistsByEmail checks if email exists
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, errors.ErrDatabaseError, "Failed to check email existence")
	}
	return exists, nil
}