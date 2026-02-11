package order

import (
	"time"

	"github.com/microservices-go/gateway/internal/common"
)

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

// OrderConnection represents order connection in GraphQL
type OrderConnection struct {
	Data     []*Order        `json:"data"`
	PageInfo common.PageInfo `json:"pageInfo"`
}

// CreateOrderInput represents order creation input
type CreateOrderInput struct {
	Currency        *string                 `json:"currency"`
	ShippingAddress string                  `json:"shipping_address"`
	Notes           *string                 `json:"notes"`
	Items           []*CreateOrderItemInput `json:"items"`
}

// CreateOrderItemInput represents order item creation input
type CreateOrderItemInput struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}
