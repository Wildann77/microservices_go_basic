package order

import (
	"context"
	"database/sql"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
)

// Repository handles order data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new order repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new order with transaction
func (r *Repository) Create(ctx context.Context, order *Order) error {
	log := logger.WithContext(ctx)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	order.BeforeCreate()
	order.CalculateTotal()

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, user_id, status, total_amount, currency, shipping_address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.Currency,
		order.ShippingAddr, order.Notes, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		log.WithError(err).Error("Failed to create order")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create order")
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, product_name, quantity, unit_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, item := range order.Items {
		item.BeforeCreate()
		item.OrderID = order.ID
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.OrderID, item.ProductID, item.ProductName,
			item.Quantity, item.UnitPrice, item.CreatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Failed to create order item")
			return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create order item")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("Order created successfully")
	return nil
}

// GetByID gets order by ID with items
func (r *Repository) GetByID(ctx context.Context, id string) (*Order, error) {
	log := logger.WithContext(ctx)

	// Get order
	orderQuery := `
		SELECT id, user_id, status, total_amount, currency, shipping_address, notes, created_at, updated_at
		FROM orders WHERE id = $1
	`
	var order Order
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.Currency,
		&order.ShippingAddr, &order.Notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrOrderNotFound
	}
	if err != nil {
		log.WithError(err).Error("Failed to get order")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get order")
	}

	// Get order items
	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

// GetByUserID gets orders by user ID with pagination
func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, user_id, status, total_amount, currency, shipping_address, notes, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to get orders by user")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get orders")
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.Currency,
			&order.ShippingAddr, &order.Notes, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Failed to scan order")
			continue
		}

		// Get items for each order
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			log.WithError(err).Warn("Failed to get order items")
		}
		order.Items = items
		orders = append(orders, &order)
	}

	return orders, nil
}

// List lists all orders with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Order, error) {
	log := logger.WithContext(ctx)

	query := `
		SELECT id, user_id, status, total_amount, currency, shipping_address, notes, created_at, updated_at
		FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to list orders")
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to list orders")
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.Currency,
			&order.ShippingAddr, &order.Notes, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Failed to scan order")
			continue
		}

		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			log.WithError(err).Warn("Failed to get order items")
		}
		order.Items = items
		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateStatus updates order status
func (r *Repository) UpdateStatus(ctx context.Context, id string, status OrderStatus) error {
	log := logger.WithContext(ctx)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer tx.Rollback()

	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, status, id)
	if err != nil {
		log.WithError(err).Error("Failed to update order status")
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to update order status")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrOrderNotFound
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	log.Info("Order status updated successfully")
	return nil
}

// getOrderItems gets order items by order ID
func (r *Repository) getOrderItems(ctx context.Context, orderID string) ([]*OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, created_at
		FROM order_items WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.CreatedAt,
		)
		if err != nil {
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

// Count returns total order count
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count orders")
	}
	return count, nil
}

// CountByUserID returns order count for a user
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "Failed to count orders")
	}
	return count, nil
}