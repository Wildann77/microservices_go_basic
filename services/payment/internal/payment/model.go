package payment

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSuccess    PaymentStatus = "success"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// PaymentMethod represents payment method
type PaymentMethod string

const (
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodEWallet      PaymentMethod = "e_wallet"
	PaymentMethodCash         PaymentMethod = "cash"
)

// Payment represents a payment entity
type Payment struct {
	ID            string        `json:"id" gorm:"primaryKey;column:id"`
	OrderID       string        `json:"order_id" gorm:"column:order_id"`
	UserID        string        `json:"user_id" gorm:"column:user_id"`
	Amount        float64       `json:"amount" gorm:"column:amount"`
	Currency      string        `json:"currency" gorm:"column:currency"`
	Status        PaymentStatus `json:"status" gorm:"column:status"`
	Method        PaymentMethod `json:"method" gorm:"column:method"`
	TransactionID string        `json:"transaction_id,omitempty" gorm:"column:transaction_id"`
	Provider      string        `json:"provider,omitempty" gorm:"column:provider"`
	Description   string        `json:"description,omitempty" gorm:"column:description"`
	FailureReason string        `json:"failure_reason,omitempty" gorm:"column:failure_reason"`
	PaidAt        *time.Time    `json:"paid_at,omitempty" gorm:"column:paid_at"`
	CreatedAt     time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate prepares payment before creation
func (p *Payment) BeforeCreate(db *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = PaymentStatusPending
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}
	return nil
}

// BeforeUpdate updates timestamps
func (p *Payment) BeforeUpdate(db *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// IsSuccessful checks if payment was successful
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusSuccess
}

// IsPending checks if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// CreatePaymentRequest represents payment creation request
type CreatePaymentRequest struct {
	OrderID     string        `json:"order_id" validate:"required"`
	UserID      string        `json:"user_id" validate:"required,uuid"`
	Amount      float64       `json:"amount" validate:"required,gt=0"`
	Currency    string        `json:"currency,omitempty" validate:"omitempty,len=3"`
	Method      PaymentMethod `json:"method" validate:"required,oneof=card bank_transfer e_wallet cash"`
	Description string        `json:"description,omitempty" validate:"omitempty,max=500"`
	Token       string        `json:"token,omitempty"` // For card payments
}

// ProcessPaymentRequest represents payment processing request
type ProcessPaymentRequest struct {
	PaymentMethodID string `json:"payment_method_id,omitempty"`
	Token           string `json:"token,omitempty"`
}

// RefundRequest represents refund request
type RefundRequest struct {
	Amount float64 `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Reason string  `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	UserID        string        `json:"user_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	Method        PaymentMethod `json:"method"`
	TransactionID string        `json:"transaction_id,omitempty"`
	Provider      string        `json:"provider,omitempty"`
	Description   string        `json:"description,omitempty"`
	FailureReason string        `json:"failure_reason,omitempty"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// PaymentIntentResponse represents payment intent response
type PaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
	PaymentID    string `json:"payment_id"`
}

// ToResponse converts Payment to PaymentResponse
func (p *Payment) ToResponse() *PaymentResponse {
	return &PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		UserID:        p.UserID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        p.Status,
		Method:        p.Method,
		TransactionID: p.TransactionID,
		Provider:      p.Provider,
		Description:   p.Description,
		FailureReason: p.FailureReason,
		PaidAt:        p.PaidAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// PaymentSuccessEvent represents payment success event
type PaymentSuccessEvent struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	TransactionID string    `json:"transaction_id"`
	PaidAt        time.Time `json:"paid_at"`
}

// PaymentFailedEvent represents payment failed event
type PaymentFailedEvent struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	FailureReason string    `json:"failure_reason"`
	FailedAt      time.Time `json:"failed_at"`
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
