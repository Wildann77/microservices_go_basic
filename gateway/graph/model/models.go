package model

import (
	"time"
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

// OrderItem represents an order item in GraphQL
type OrderItem struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Order represents an order in GraphQL
type Order struct {
	ID              string       `json:"id"`
	UserID          string       `json:"user_id"`
	Status          string       `json:"status"`
	TotalAmount     float64      `json:"total_amount"`
	Currency        string       `json:"currency"`
	ShippingAddress string       `json:"shipping_address"`
	Notes           *string      `json:"notes"`
	Items           []*OrderItem `json:"items"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// Payment represents a payment in GraphQL
type Payment struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	UserID        string     `json:"user_id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	Method        string     `json:"method"`
	TransactionID *string    `json:"transaction_id"`
	Provider      *string    `json:"provider"`
	Description   *string    `json:"description"`
	FailureReason *string    `json:"failure_reason"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// AuthResponse represents auth response in GraphQL
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// PageInfo represents pagination info in GraphQL
type PageInfo struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"hasMore"`
}

// UserConnection represents user connection in GraphQL
type UserConnection struct {
	Data     []*User  `json:"data"`
	PageInfo PageInfo `json:"pageInfo"`
}

// OrderConnection represents order connection in GraphQL
type OrderConnection struct {
	Data     []*Order `json:"data"`
	PageInfo PageInfo `json:"pageInfo"`
}

// PaymentConnection represents payment connection in GraphQL
type PaymentConnection struct {
	Data     []*Payment `json:"data"`
	PageInfo PageInfo   `json:"pageInfo"`
}
