package order

import (
	"context"
	"time"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"gorm.io/gorm"
)

// Repository handles order data access
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new order repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTransaction executes a function within a database transaction
func (r *Repository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create creates a new order
func (r *Repository) Create(ctx context.Context, order *Order) error {
	return r.CreateWithDB(ctx, r.db, order)
}

// CreateWithDB creates a new order using the provided database connection
func (r *Repository) CreateWithDB(ctx context.Context, db *gorm.DB, order *Order) error {
	log := logger.WithContext(ctx)

	// Calculate total before saving
	order.CalculateTotal()

	if err := db.WithContext(ctx).Create(order).Error; err != nil {
		log.WithError(err).Error("Failed to create order")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create order")
	}

	log.Info("Order created successfully")
	return nil
}

// GetByID gets order by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Order, error) {
	log := logger.WithContext(ctx)

	var order Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrOrderNotFound
		}
		log.WithError(err).Error("Failed to get order")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get order")
	}

	return &order, nil
}

// GetByUserID gets orders by user ID with pagination
func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error) {
	log := logger.WithContext(ctx)

	var orders []*Order
	err := r.db.WithContext(ctx).Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		log.WithError(err).Error("Failed to get orders by user")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get orders")
	}

	return orders, nil
}

// List lists all orders with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Order, error) {
	log := logger.WithContext(ctx)

	var orders []*Order
	err := r.db.WithContext(ctx).Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		log.WithError(err).Error("Failed to list orders")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list orders")
	}

	return orders, nil
}

// UpdateStatus updates order status
func (r *Repository) UpdateStatus(ctx context.Context, id string, status OrderStatus) error {
	return r.UpdateStatusWithDB(ctx, r.db, id, status)
}

// UpdateStatusWithDB updates order status using the provided database connection
func (r *Repository) UpdateStatusWithDB(ctx context.Context, db *gorm.DB, id string, status OrderStatus) error {
	log := logger.WithContext(ctx)

	result := db.WithContext(ctx).Model(&Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

	if err := result.Error; err != nil {
		log.WithError(err).Error("Failed to update order status")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update order status")
	}

	if result.RowsAffected == 0 {
		return errors.ErrOrderNotFound
	}

	log.Info("Order status updated successfully")
	return nil
}

// Count returns total order count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Order{}).Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count orders")
	}
	return int(count), nil
}

// CountByUserID returns order count for a user
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Order{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count orders")
	}
	return int(count), nil
}

// GetByIDs gets multiple orders by IDs
func (r *Repository) GetByIDs(ctx context.Context, ids []string) ([]*Order, error) {
	log := logger.WithContext(ctx)

	if len(ids) == 0 {
		return []*Order{}, nil
	}

	var orders []*Order
	err := r.db.WithContext(ctx).Preload("Items").Where("id IN ?", ids).Find(&orders).Error
	if err != nil {
		log.WithError(err).Error("Failed to get orders by IDs")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get orders by IDs")
	}

	return orders, nil
}
