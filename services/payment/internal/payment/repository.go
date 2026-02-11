package payment

import (
	"context"
	"database/sql"
	"time"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
)

// Repository handles payment data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new payment repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new payment with transaction
func (r *Repository) Create(ctx context.Context, payment *Payment) error {
	log := logger.WithContext(ctx)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	payment.BeforeCreate()

	query := `
		INSERT INTO payments (id, order_id, user_id, amount, currency, status, method, 
		transaction_id, provider, description, failure_reason, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = tx.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount, payment.Currency,
		payment.Status, payment.Method, payment.TransactionID, payment.Provider,
		payment.Description, payment.FailureReason, payment.PaidAt,
		payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		log.WithError(err).Error("Failed to create payment")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create payment")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("Payment created successfully")
	return nil
}

// GetByID gets payment by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, order_id, user_id, amount, currency, status, method,
		transaction_id, provider, description, failure_reason, paid_at, created_at, updated_at
		FROM payments WHERE id = $1
	`

	var payment Payment
	var paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.Method, &payment.TransactionID, &payment.Provider,
		&payment.Description, &payment.FailureReason, &paidAt,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrPaymentNotFound
	}
	if err != nil {
		log.WithError(err).Error("Failed to get payment")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payment")
	}

	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}

	return &payment, nil
}

// GetByOrderID gets payment by order ID
func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, order_id, user_id, amount, currency, status, method,
		transaction_id, provider, description, failure_reason, paid_at, created_at, updated_at
		FROM payments WHERE order_id = $1
	`

	var payment Payment
	var paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.Method, &payment.TransactionID, &payment.Provider,
		&payment.Description, &payment.FailureReason, &paidAt,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrPaymentNotFound
	}
	if err != nil {
		log.WithError(err).Error("Failed to get payment by order ID")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payment")
	}

	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}

	return &payment, nil
}

// GetByUserID gets payments by user ID with pagination
func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Payment, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, order_id, user_id, amount, currency, status, method,
		transaction_id, provider, description, failure_reason, paid_at, created_at, updated_at
		FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to get payments by user")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get payments")
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// List lists all payments with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Payment, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, order_id, user_id, amount, currency, status, method,
		transaction_id, provider, description, failure_reason, paid_at, created_at, updated_at
		FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to list payments")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list payments")
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// UpdateStatus updates payment status
func (r *Repository) UpdateStatus(ctx context.Context, id string, status PaymentStatus, failureReason string) error {
	log := logger.WithContext(ctx)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	var query string
	var args []interface{}

	if status == PaymentStatusSuccess {
		now := time.Now()
		query = `UPDATE payments SET status = $1, failure_reason = $2, paid_at = $3, updated_at = NOW() WHERE id = $4`
		args = []interface{}{status, failureReason, now, id}
	} else {
		query = `UPDATE payments SET status = $1, failure_reason = $2, updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, failureReason, id}
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		log.WithError(err).Error("Failed to update payment status")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update payment status")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrPaymentNotFound
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("Payment status updated successfully")
	return nil
}

// UpdateTransactionID updates payment transaction ID
func (r *Repository) UpdateTransactionID(ctx context.Context, id string, transactionID, provider string) error {
	log := logger.WithContext(ctx)

	query := `UPDATE payments SET transaction_id = $1, provider = $2, updated_at = NOW() WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, transactionID, provider, id)
	if err != nil {
		log.WithError(err).Error("Failed to update transaction ID")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update transaction ID")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrPaymentNotFound
	}

	return nil
}

// scanPayments scans payment rows
func (r *Repository) scanPayments(rows *sql.Rows) ([]*Payment, error) {
	var payments []*Payment

	for rows.Next() {
		var payment Payment
		var paidAt sql.NullTime

		err := rows.Scan(
			&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.Method, &payment.TransactionID, &payment.Provider,
			&payment.Description, &payment.FailureReason, &paidAt,
			&payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if paidAt.Valid {
			payment.PaidAt = &paidAt.Time
		}

		payments = append(payments, &payment)
	}

	return payments, nil
}

// Count returns total payment count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments").Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count payments")
	}
	return count, nil
}

// CountByUserID returns payment count for a user
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count payments")
	}
	return count, nil
}