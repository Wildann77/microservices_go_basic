package user

import (
	"time"

	"github.com/microservices-go/gateway/internal/common"
)

// User represents a user in GraphQL
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FullName returns user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// UserConnection represents user connection in GraphQL
type UserConnection struct {
	Data     []*User         `json:"data"`
	PageInfo common.PageInfo `json:"pageInfo"`
}

// RegisterInput represents registration input
type RegisterInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginInput represents login input
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
