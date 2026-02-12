package payment

import (
	"context"
	"time"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"gorm.io/gorm"
)

// Repository handles payment data access
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new payment repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTransaction executes a function within a database transaction
func (r *Repository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create creates a new payment
func (r *Repository) Create(ctx context.Context, payment *Payment) error {
	return r.CreateWithDB(ctx, r.db, payment)
}

// CreateWithDB creates a new payment using the provided database connection
func (r *Repository) CreateWithDB(ctx context.Context, db *gorm.DB, payment *Payment) error {
	log := logger.WithContext(ctx)

	if err := db.WithContext(ctx).Create(payment).Error; err != nil {
		log.WithError(err).Error("Failed to create payment")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create payment")
	}

	log.Info("Payment created successfully")
	return nil
}

// CreateOutboxEvent creates a new given outbox event
func (r *Repository) CreateOutboxEvent(ctx context.Context, tx *gorm.DB, event *OutboxEvent) error {
	log := logger.WithContext(ctx)

	if err := tx.WithContext(ctx).Create(event).Error; err != nil {
		log.WithError(err).Error("Failed to create outbox event")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create outbox event")
	}

	return nil
}

// UpdateOutboxEvent updates an outbox event
func (r *Repository) UpdateOutboxEvent(ctx context.Context, id string, status string, errStr string) error {
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": time.Now(),
	}
	if errStr != "" {
		updates["error_message"] = errStr
	}
	return r.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", id).Updates(updates).Error
}

// GetByID gets payment by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	log := logger.WithContext(ctx)

	var payment Payment
	err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPaymentNotFound
		}
		log.WithError(err).Error("Failed to get payment")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payment")
	}

	return &payment, nil
}

// GetByOrderID gets payment by order ID
func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	log := logger.WithContext(ctx)

	var payment Payment
	err := r.db.WithContext(ctx).First(&payment, "order_id = ?", orderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPaymentNotFound
		}
		log.WithError(err).Error("Failed to get payment by order ID")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payment")
	}

	return &payment, nil
}

// GetByUserID gets payments by user ID with pagination
func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Payment, error) {
	log := logger.WithContext(ctx)

	var payments []*Payment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	if err != nil {
		log.WithError(err).Error("Failed to get payments by user")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payments")
	}

	return payments, nil
}

// List lists all payments with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Payment, error) {
	log := logger.WithContext(ctx)

	var payments []*Payment
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	if err != nil {
		log.WithError(err).Error("Failed to list payments")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list payments")
	}

	return payments, nil
}

// UpdateStatus updates payment status
func (r *Repository) UpdateStatus(ctx context.Context, id string, status PaymentStatus, failureReason string) error {
	return r.UpdateStatusWithDB(ctx, r.db, id, status, failureReason)
}

// UpdateStatusWithDB updates payment status using the provided database connection
func (r *Repository) UpdateStatusWithDB(ctx context.Context, db *gorm.DB, id string, status PaymentStatus, failureReason string) error {
	log := logger.WithContext(ctx)

	updates := map[string]interface{}{
		"status":         status,
		"failure_reason": failureReason,
		"updated_at":     time.Now(),
	}

	if status == PaymentStatusSuccess {
		updates["paid_at"] = time.Now()
	}

	result := db.WithContext(ctx).Model(&Payment{}).
		Where("id = ?", id).
		Updates(updates)

	if err := result.Error; err != nil {
		log.WithError(err).Error("Failed to update payment status")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update payment status")
	}

	if result.RowsAffected == 0 {
		return errors.ErrPaymentNotFound
	}

	log.Info("Payment status updated successfully")
	return nil
}

// UpdateTransactionID updates payment transaction ID
func (r *Repository) UpdateTransactionID(ctx context.Context, id string, transactionID, provider string) error {
	return r.UpdateTransactionIDWithDB(ctx, r.db, id, transactionID, provider)
}

// UpdateTransactionIDWithDB updates payment transaction ID using the provided database connection
func (r *Repository) UpdateTransactionIDWithDB(ctx context.Context, db *gorm.DB, id string, transactionID, provider string) error {
	log := logger.WithContext(ctx)

	result := db.WithContext(ctx).Model(&Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"transaction_id": transactionID,
			"provider":       provider,
			"updated_at":     time.Now(),
		})

	if err := result.Error; err != nil {
		log.WithError(err).Error("Failed to update transaction ID")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update transaction ID")
	}

	if result.RowsAffected == 0 {
		return errors.ErrPaymentNotFound
	}

	return nil
}

// Count returns total payment count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Payment{}).Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count payments")
	}
	return int(count), nil
}

// CountByUserID returns payment count for a user
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Payment{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count payments")
	}
	return int(count), nil
}
