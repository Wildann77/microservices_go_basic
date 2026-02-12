package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/validator"
	"gorm.io/gorm"
)

// PaymentProvider interface for payment providers
type PaymentProvider interface {
	CreatePaymentIntent(ctx context.Context, amount float64, currency string) (*PaymentIntentResult, error)
	ConfirmPayment(ctx context.Context, paymentIntentID string) (*PaymentResult, error)
	Refund(ctx context.Context, transactionID string, amount float64) (*RefundResult, error)
}

// PaymentIntentResult represents payment intent result
type PaymentIntentResult struct {
	ClientSecret    string
	PaymentIntentID string
}

// PaymentResult represents payment result
type PaymentResult struct {
	Success       bool
	TransactionID string
	FailureReason string
}

// RefundResult represents refund result
type RefundResult struct {
	Success       bool
	RefundID      string
	FailureReason string
}

// Service handles payment business logic
type Service struct {
	repo      *Repository
	validator *validator.Validator
	provider  PaymentProvider
	publisher EventPublisher
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

// NewService creates a new payment service
func NewService(repo *Repository, provider PaymentProvider, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		validator: validator.New(),
		provider:  provider,
		publisher: publisher,
	}
}

// Create creates a new payment
func (s *Service) Create(ctx context.Context, req *CreatePaymentRequest) (*PaymentResponse, error) {
	log := logger.WithContext(ctx)

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Check if payment already exists for order
	existing, _ := s.repo.GetByOrderID(ctx, req.OrderID)
	if existing != nil {
		return nil, errors.New(errors.ErrConflict, "Payment already exists for this order")
	}

	// Create payment
	payment := &Payment{
		OrderID:     req.OrderID,
		UserID:      req.UserID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Description: req.Description,
	}

	// Use transaction for database operations
	var transactionID, provider string
	if req.Method == PaymentMethodCard && s.provider != nil && req.Token != "" {
		// Create payment intent with provider first (external call)
		result, err := s.provider.CreatePaymentIntent(ctx, payment.Amount, payment.Currency)
		if err != nil {
			log.WithError(err).Warn("Failed to create payment intent")
		} else {
			transactionID = result.PaymentIntentID
			provider = "stripe"
		}
	}

	// Create payment and update transaction ID in single transaction
	if err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.repo.CreateWithDB(ctx, tx, payment); err != nil {
			return err
		}
		if transactionID != "" {
			if err := s.repo.UpdateTransactionIDWithDB(ctx, tx, payment.ID, transactionID, provider); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return payment.ToResponse(), nil
}

// GetByID gets payment by ID
func (s *Service) GetByID(ctx context.Context, id string) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return payment.ToResponse(), nil
}

// GetByOrderID gets payment by order ID
func (s *Service) GetByOrderID(ctx context.Context, orderID string) (*PaymentResponse, error) {
	payment, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return payment.ToResponse(), nil
}

// GetByUserID gets payments by user ID
func (s *Service) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*PaymentResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	payments, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}
	return responses, nil
}

// List lists all payments
func (s *Service) List(ctx context.Context, limit, offset int) ([]*PaymentResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	payments, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}
	return responses, nil
}

// Process processes a payment
func (s *Service) Process(ctx context.Context, id string, req *ProcessPaymentRequest) (*PaymentResponse, error) {
	log := logger.WithContext(ctx)

	// Get payment
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if payment can be processed
	if !payment.IsPending() {
		return nil, errors.New(errors.ErrConflict, fmt.Sprintf("Payment cannot be processed, current status: %s", payment.Status))
	}

	// Process with provider if available
	if s.provider != nil && payment.TransactionID != "" {
		result, err := s.provider.ConfirmPayment(ctx, payment.TransactionID)
		if err != nil {
			log.WithError(err).Error("Payment processing failed")
			s.failPayment(ctx, payment, err.Error())
			return nil, errors.New(errors.ErrServiceUnavailable, "Payment processing failed")
		}

		if result.Success {
			if err := s.completePayment(ctx, payment); err != nil {
				return nil, err
			}
		} else {
			s.failPayment(ctx, payment, result.FailureReason)
			return nil, errors.New(errors.ErrServiceUnavailable, "Payment failed: "+result.FailureReason)
		}
	} else {
		// Simulate payment processing
		time.Sleep(100 * time.Millisecond)
		if err := s.completePayment(ctx, payment); err != nil {
			return nil, err
		}
	}

	return s.GetByID(ctx, id)
}

// completePayment marks payment as completed
func (s *Service) completePayment(ctx context.Context, payment *Payment) error {
	log := logger.WithContext(ctx)

	if err := s.repo.UpdateStatus(ctx, payment.ID, PaymentStatusSuccess, ""); err != nil {
		return err
	}

	// Publish success event
	if s.publisher != nil {
		event := &PaymentSuccessEvent{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			TransactionID: payment.TransactionID,
			PaidAt:        time.Now(),
		}
		if err := s.publisher.PublishEvent(ctx, "payment.success", event); err != nil {
			log.WithError(err).Warn("Failed to publish payment success event")
		}
	}

	log.Info("Payment completed successfully")
	return nil
}

// failPayment marks payment as failed
func (s *Service) failPayment(ctx context.Context, payment *Payment, reason string) {
	log := logger.WithContext(ctx)

	if err := s.repo.UpdateStatus(ctx, payment.ID, PaymentStatusFailed, reason); err != nil {
		log.WithError(err).Error("Failed to update payment status to failed")
	}

	// Publish failure event
	if s.publisher != nil {
		event := &PaymentFailedEvent{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			FailureReason: reason,
			FailedAt:      time.Now(),
		}
		if err := s.publisher.PublishEvent(ctx, "payment.failed", event); err != nil {
			log.WithError(err).Warn("Failed to publish payment failed event")
		}
	}
}

// Refund refunds a payment
func (s *Service) Refund(ctx context.Context, id string, req *RefundRequest) (*PaymentResponse, error) {
	log := logger.WithContext(ctx)

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Get payment
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if payment can be refunded
	if !payment.IsSuccessful() {
		return nil, errors.New(errors.ErrConflict, "Only successful payments can be refunded")
	}

	// Process refund with provider if available
	if s.provider != nil && payment.TransactionID != "" {
		refundAmount := req.Amount
		if refundAmount == 0 {
			refundAmount = payment.Amount
		}

		result, err := s.provider.Refund(ctx, payment.TransactionID, refundAmount)
		if err != nil {
			log.WithError(err).Error("Refund failed")
			return nil, errors.New(errors.ErrServiceUnavailable, "Refund processing failed")
		}

		if !result.Success {
			return nil, errors.New(errors.ErrServiceUnavailable, "Refund failed: "+result.FailureReason)
		}
	}

	// Update payment status
	if err := s.repo.UpdateStatus(ctx, payment.ID, PaymentStatusRefunded, req.Reason); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

// Count returns total payment count
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// CountByUserID returns payment count for a user
func (s *Service) CountByUserID(ctx context.Context, userID string) (int, error) {
	return s.repo.CountByUserID(ctx, userID)
}

// HandleOrderCreated handles order created event
func (s *Service) HandleOrderCreated(ctx context.Context, orderID, userID string, amount float64, currency string) error {
	log := logger.WithContext(ctx)
	log.Infof("Handling order.created event for order: %s", orderID)

	// Auto-create payment for order
	req := &CreatePaymentRequest{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   amount,
		Currency: currency,
		Method:   PaymentMethodCard,
	}

	_, err := s.Create(ctx, req)
	return err
}
