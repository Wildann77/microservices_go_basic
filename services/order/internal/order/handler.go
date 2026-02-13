package order

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
)

// Handler handles HTTP requests for order service
type Handler struct {
	service *Service
}

// NewHandler creates a new order handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/", h.Create)
			r.Get("/", h.List)
			r.Post("/batch", h.GetBatch)
			r.Get("/my-orders", h.GetMyOrders)
			r.Get("/user/{userId}", h.GetByUserID)
			r.Get("/{id}", h.GetByID)
			r.Patch("/{id}/status", h.UpdateStatus)
		})
	})
}

// Create handles order creation
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	order, err := h.service.Create(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to create order").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": order,
	})
}

// GetByID gets order by ID
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	order, err := h.service.GetByID(ctx, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get order").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": order,
	})
}

// GetMyOrders gets orders for current user
func (h *Handler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		errors.New(errors.ErrUnauthorized, "User not authenticated").WriteHTTPResponse(w)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.service.GetByUserID(ctx, claims.UserID, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get orders").WriteHTTPResponse(w)
		return
	}

	count, _ := h.service.CountByUserID(ctx, claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": orders,
		"meta": map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// List lists all orders (admin only)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.service.List(ctx, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to list orders").WriteHTTPResponse(w)
		return
	}

	count, _ := h.service.Count(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": orders,
		"meta": map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateStatus updates order status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	order, err := h.service.UpdateStatus(ctx, id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to update order status").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": order,
	})
}

// GetBatch gets multiple orders by IDs
func (h *Handler) GetBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	if len(req.IDs) == 0 {
		errors.New(errors.ErrInvalidInput, "IDs array is required").WriteHTTPResponse(w)
		return
	}

	orders, err := h.service.GetByIDs(ctx, req.IDs)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get orders").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": orders,
	})
}

// GetByUserID gets orders for a specific user
func (h *Handler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userId")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.service.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get orders").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": orders,
	})
}

// HealthCheck handles health check
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Health check called")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "order-service",
	})
}
