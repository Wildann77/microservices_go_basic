package payment

import (
	"time"

	"github.com/google/uuid"
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
	PaymentMethodCard       PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodEWallet    PaymentMethod = "e_wallet"
	PaymentMethodCash       PaymentMethod = "cash"
)

// Payment represents a payment entity
type Payment struct {
	ID            string        `json:"id" db:"id"`
	OrderID       string        `json:"order_id" db:"order_id"`
	UserID        string        `json:"user_id" db:"user_id"`
	Amount        float64       `json:"amount" db:"amount"`
	Currency      string        `json:"currency" db:"currency"`
	Status        PaymentStatus `json:"status" db:"status"`
	Method        PaymentMethod `json:"method" db:"method"`
	TransactionID string        `json:"transaction_id,omitempty" db:"transaction_id"`
	Provider      string        `json:"provider,omitempty" db:"provider"`
	Description   string        `json:"description,omitempty" db:"description"`
	FailureReason string        `json:"failure_reason,omitempty" db:"failure_reason"`
	PaidAt        *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate prepares payment before creation
func (p *Payment) BeforeCreate() {
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
}

// BeforeUpdate updates timestamps
func (p *Payment) BeforeUpdate() {
	p.UpdatedAt = time.Now()
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
	Token          string `json:"token,omitempty"`
}

// RefundRequest represents refund request
type RefundRequest struct {
	Amount  float64 `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Reason  string  `json:"reason,omitempty" validate:"omitempty,max=500"`
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