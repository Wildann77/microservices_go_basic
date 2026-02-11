package payment

import (
	"time"

	"github.com/microservices-go/gateway/internal/common"
)

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

// PaymentConnection represents payment connection in GraphQL
type PaymentConnection struct {
	Data     []*Payment      `json:"data"`
	PageInfo common.PageInfo `json:"pageInfo"`
}

// CreatePaymentInput represents payment creation input
type CreatePaymentInput struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	Currency    *string `json:"currency"`
	Method      string  `json:"method"`
	Description *string `json:"description"`
}
