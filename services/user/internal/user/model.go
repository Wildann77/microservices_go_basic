package user

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user entity
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;column:id"`
	Email     string    `json:"email" gorm:"unique;column:email"`
	Password  string    `json:"-" gorm:"column:password_hash"`
	FirstName string    `json:"first_name" gorm:"column:first_name"`
	LastName  string    `json:"last_name" gorm:"column:last_name"`
	Role      string    `json:"role" gorm:"column:role"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name
func (User) TableName() string {
	return "users"
}

// FullName returns user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// HashPassword hashes the password
func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// BeforeCreate prepares user before creation
func (u *User) BeforeCreate(db *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.Role == "" {
		u.Role = "user"
	}
	u.IsActive = true
	return nil
}

// BeforeUpdate updates timestamps
func (u *User) BeforeUpdate(db *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// UserResponse represents user response (without sensitive data)
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserCreatedEvent represents user created event
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

// OutboxEvent represents an event to be published
type OutboxEvent struct {
	ID            string          `gorm:"primaryKey;column:id"`
	AggregateType string          `gorm:"column:aggregate_type"`
	AggregateID   string          `gorm:"column:aggregate_id"`
	Type          string          `gorm:"column:type"`
	Payload       json.RawMessage `gorm:"column:payload;type:jsonb"`
	Status        string          `gorm:"column:status"`
	RetryCount    int             `gorm:"column:retry_count"`
	ErrorMessage  string          `gorm:"column:error_message"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	ProcessedAt   *time.Time      `gorm:"column:processed_at"`
}

// TableName returns the table name
func (OutboxEvent) TableName() string {
	return "outbox_events"
}

// BeforeCreate sets default values
func (e *OutboxEvent) BeforeCreate(db *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	e.CreatedAt = time.Now()
	if e.Status == "" {
		e.Status = "PENDING"
	}
	return nil
}
