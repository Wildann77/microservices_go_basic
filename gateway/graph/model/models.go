package model

import (
	"context"
	"time"
)

// User represents a user in GraphQL
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FullName returns user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// OrderItem represents an order item in GraphQL
type OrderItem struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
}

// Order represents an order in GraphQL
type Order struct {
	ID           string       `json:"id"`
	UserID       string       `json:"userId"`
	Status       string       `json:"status"`
	TotalAmount  float64      `json:"totalAmount"`
	Currency     string       `json:"currency"`
	ShippingAddr string       `json:"shippingAddress"`
	Notes        *string      `json:"notes"`
	Items        []*OrderItem `json:"items"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// Payment represents a payment in GraphQL
type Payment struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"orderId"`
	UserID        string     `json:"userId"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	Method        string     `json:"method"`
	TransactionID *string    `json:"transactionId"`
	Provider      *string    `json:"provider"`
	Description   *string    `json:"description"`
	FailureReason *string    `json:"failureReason"`
	PaidAt        *time.Time `json:"paidAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
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

// LoaderKey is the key for dataloaders in context
type LoaderKey struct{}

// Loaders holds all dataloaders
type Loaders struct {
	UserLoader    *UserLoader
	OrderLoader   *OrderLoader
	PaymentLoader *PaymentLoader
}

// GetLoaders gets loaders from context
func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(LoaderKey{}).(*Loaders)
}