package order

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order represents an order entity
type Order struct {
	ID          string      `json:"id" db:"id"`
	UserID      string      `json:"user_id" db:"user_id"`
	Status      OrderStatus `json:"status" db:"status"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	Currency    string      `json:"currency" db:"currency"`
	ShippingAddr string     `json:"shipping_address" db:"shipping_address"`
	Notes       string      `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	Items       []*OrderItem `json:"items,omitempty" db:"-"`
}

// TableName returns the table name
func (Order) TableName() string {
	return "orders"
}

// BeforeCreate prepares order before creation
func (o *Order) BeforeCreate() {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now
	if o.Status == "" {
		o.Status = OrderStatusPending
	}
	if o.Currency == "" {
		o.Currency = "USD"
	}
}

// BeforeUpdate updates timestamps
func (o *Order) BeforeUpdate() {
	o.UpdatedAt = time.Now()
}

// CalculateTotal calculates total amount from items
func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}
	o.TotalAmount = total
}

// OrderItem represents an order item
type OrderItem struct {
	ID        string    `json:"id" db:"id"`
	OrderID   string    `json:"order_id" db:"order_id"`
	ProductID string    `json:"product_id" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Quantity  int       `json:"quantity" db:"quantity"`
	UnitPrice float64   `json:"unit_price" db:"unit_price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TableName returns the table name
func (OrderItem) TableName() string {
	return "order_items"
}

// BeforeCreate prepares order item before creation
func (oi *OrderItem) BeforeCreate() {
	if oi.ID == "" {
		oi.ID = uuid.New().String()
	}
	oi.CreatedAt = time.Now()
}

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	UserID       string                `json:"user_id" validate:"required,uuid"`
	Currency     string                `json:"currency,omitempty" validate:"omitempty,len=3"`
	ShippingAddr string                `json:"shipping_address" validate:"required,max=500"`
	Notes        string                `json:"notes,omitempty" validate:"omitempty,max=1000"`
	Items        []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

// CreateOrderItemRequest represents order item creation request
type CreateOrderItemRequest struct {
	ProductID   string  `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name" validate:"required,max=255"`
	Quantity    int     `json:"quantity" validate:"required,gte=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// UpdateOrderStatusRequest represents order status update request
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=pending confirmed processing shipped delivered cancelled"`
}

// OrderResponse represents order response
type OrderResponse struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Status       OrderStatus       `json:"status"`
	TotalAmount  float64           `json:"total_amount"`
	Currency     string            `json:"currency"`
	ShippingAddr string            `json:"shipping_address"`
	Notes        string            `json:"notes,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Items        []*OrderItemResponse `json:"items,omitempty"`
}

// OrderItemResponse represents order item response
type OrderItemResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// ToResponse converts Order to OrderResponse
func (o *Order) ToResponse() *OrderResponse {
	items := make([]*OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = item.ToResponse()
	}
	return &OrderResponse{
		ID:           o.ID,
		UserID:       o.UserID,
		Status:       o.Status,
		TotalAmount:  o.TotalAmount,
		Currency:     o.Currency,
		ShippingAddr: o.ShippingAddr,
		Notes:        o.Notes,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
		Items:        items,
	}
}

// ToResponse converts OrderItem to OrderItemResponse
func (oi *OrderItem) ToResponse() *OrderItemResponse {
	return &OrderItemResponse{
		ID:          oi.ID,
		ProductID:   oi.ProductID,
		ProductName: oi.ProductName,
		Quantity:    oi.Quantity,
		UnitPrice:   oi.UnitPrice,
	}
}

// OrderCreatedEvent represents order created event
type OrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrderStatusChangedEvent represents order status changed event
type OrderStatusChangedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}