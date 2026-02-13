package order

import (
	"context"

	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/validator"
)

// Service handles order business logic
type Service struct {
	repo      *Repository
	validator *validator.Validator
	publisher EventPublisher
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

// NewService creates a new order service
func NewService(repo *Repository, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		validator: validator.New(),
		publisher: publisher,
	}
}

// Create creates a new order
func (s *Service) Create(ctx context.Context, req *CreateOrderRequest) (*OrderResponse, error) {
	log := logger.WithContext(ctx)

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Create order items
	items := make([]*OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		items[i] = &OrderItem{
			ProductID:   itemReq.ProductID,
			ProductName: itemReq.ProductName,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
		}
	}

	// Create order
	order := &Order{
		UserID:       req.UserID,
		Currency:     req.Currency,
		ShippingAddr: req.ShippingAddr,
		Notes:        req.Notes,
		Items:        items,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Publish event
	if s.publisher != nil {
		event := &OrderCreatedEvent{
			OrderID:     order.ID,
			UserID:      order.UserID,
			TotalAmount: order.TotalAmount,
			Currency:    order.Currency,
			Status:      string(order.Status),
			Notes:       order.Notes,
			CreatedAt:   order.CreatedAt,
		}
		if err := s.publisher.PublishEvent(ctx, "order.created", event); err != nil {
			log.WithError(err).Warn("Failed to publish order created event")
		}
	}

	return order.ToResponse(), nil
}

// GetByID gets order by ID
func (s *Service) GetByID(ctx context.Context, id string) (*OrderResponse, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return order.ToResponse(), nil
}

// GetByUserID gets orders by user ID
func (s *Service) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*OrderResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}
	return responses, nil
}

// List lists all orders
func (s *Service) List(ctx context.Context, limit, offset int) ([]*OrderResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}
	return responses, nil
}

// UpdateStatus updates order status
func (s *Service) UpdateStatus(ctx context.Context, id string, req *UpdateOrderStatusRequest) (*OrderResponse, error) {
	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Get existing order
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := order.Status

	// Update status
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, err
	}

	// Publish event
	if s.publisher != nil && oldStatus != req.Status {
		event := &OrderStatusChangedEvent{
			OrderID:   order.ID,
			UserID:    order.UserID,
			OldStatus: string(oldStatus),
			NewStatus: string(req.Status),
			ChangedAt: order.UpdatedAt,
		}
		if err := s.publisher.PublishEvent(ctx, "order.status_changed", event); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to publish order status changed event")
		}
	}

	// Get updated order
	return s.GetByID(ctx, id)
}

// Count returns total order count
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// CountByUserID returns order count for a user
func (s *Service) CountByUserID(ctx context.Context, userID string) (int, error) {
	return s.repo.CountByUserID(ctx, userID)
}

// HandleUserCreated handles user created event
func (s *Service) HandleUserCreated(ctx context.Context, userID string) error {
	logger.WithContext(ctx).Infof("Handling user created event for user: %s", userID)
	return nil
}

// HandlePaymentSuccess handles payment success event
func (s *Service) HandlePaymentSuccess(ctx context.Context, orderID string) error {
	log := logger.WithContext(ctx)
	log.Infof("Handling payment success for order: %s", orderID)

	// Update order status to confirmed
	_, err := s.UpdateStatus(ctx, orderID, &UpdateOrderStatusRequest{
		Status: OrderStatusConfirmed,
	})
	return err
}

// HandlePaymentFailed handles payment failed event
func (s *Service) HandlePaymentFailed(ctx context.Context, orderID string) error {
	log := logger.WithContext(ctx)
	log.Infof("Handling payment failed for order: %s", orderID)

	// Update order status to cancelled
	_, err := s.UpdateStatus(ctx, orderID, &UpdateOrderStatusRequest{
		Status: OrderStatusCancelled,
	})
	return err
}
