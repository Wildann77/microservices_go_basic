package order

import (
	"context"
	"time"

	"github.com/microservices-go/shared/cache"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/validator"
)

// Service handles order business logic
type Service struct {
	repo      *Repository
	cache     *cache.Cache
	validator *validator.Validator
	publisher EventPublisher
	cacheTTL  time.Duration
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

// NewService creates a new order service
func NewService(repo *Repository, publisher EventPublisher, cacheClient *cache.Cache) *Service {
	return &Service{
		repo:      repo,
		cache:     cacheClient,
		validator: validator.New(),
		publisher: publisher,
		cacheTTL:  3 * time.Minute,
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

	// Invalidate user orders cache
	if s.cache != nil {
		if err := s.cache.DeletePattern(ctx, "orders:user:"+order.UserID+":*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate user orders cache")
		}
		if err := s.cache.DeletePattern(ctx, "orders:list:*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate orders list cache")
		}
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

// GetByID gets order by ID with caching
func (s *Service) GetByID(ctx context.Context, id string) (*OrderResponse, error) {
	cacheKey := "order:id:" + id

	// Try to get from cache
	if s.cache != nil {
		var cachedOrder OrderResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedOrder); err == nil {
			return &cachedOrder, nil
		}
	}

	// Get from database
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := order.ToResponse()

	// Store in cache
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, response, s.cacheTTL); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to cache order")
		}
	}

	return response, nil
}

// GetByUserID gets orders by user ID with caching
func (s *Service) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*OrderResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	cacheKey := "orders:user:" + userID + ":limit:" + string(rune(limit)) + ":offset:" + string(rune(offset))

	// Try to get from cache
	if s.cache != nil {
		var cachedOrders []*OrderResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedOrders); err == nil {
			return cachedOrders, nil
		}
	}

	// Get from database
	orders, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	// Store in cache
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, responses, s.cacheTTL); err != nil {
			logger.WithContext(ctx).WithError(err).Warn("Failed to cache user orders")
		}
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

	// Invalidate caches
	if s.cache != nil {
		log := logger.WithContext(ctx)
		if err := s.cache.Delete(ctx, "order:id:"+id); err != nil {
			log.WithError(err).Warn("Failed to invalidate order cache")
		}
		if err := s.cache.DeletePattern(ctx, "orders:user:"+order.UserID+":*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate user orders cache")
		}
		if err := s.cache.DeletePattern(ctx, "orders:list:*"); err != nil {
			log.WithError(err).Warn("Failed to invalidate orders list cache")
		}
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

// GetByIDs gets multiple orders by IDs with caching
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*OrderResponse, error) {
	if len(ids) == 0 {
		return []*OrderResponse{}, nil
	}

	// Get from database
	orders, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	responses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()

		// Cache individual orders
		if s.cache != nil {
			cacheKey := "order:id:" + order.ID
			if err := s.cache.Set(ctx, cacheKey, responses[i], s.cacheTTL); err != nil {
				logger.WithContext(ctx).WithError(err).Warn("Failed to cache order")
			}
		}
	}

	return responses, nil
}
