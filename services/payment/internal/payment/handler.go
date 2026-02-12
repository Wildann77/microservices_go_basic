package payment

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
)

// Handler handles HTTP requests for payment service
type Handler struct {
	service *Service
}

// NewHandler creates a new payment handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/api/v1/payments", func(r chi.Router) {
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/", h.Create)
			r.Get("/", h.List)
			r.Get("/my-payments", h.GetMyPayments)
			r.Get("/{id}", h.GetByID)
			r.Get("/order/{orderId}", h.GetByOrderID)
			r.Post("/{id}/process", h.Process)
			r.Post("/{id}/refund", h.Refund)
		})
	})
}

// Create handles payment creation
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	payment, err := h.service.Create(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to create payment").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payment,
	})
}

// GetByID gets payment by ID
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	payment, err := h.service.GetByID(ctx, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get payment").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payment,
	})
}

// GetByOrderID gets payment by order ID
func (h *Handler) GetByOrderID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderID := chi.URLParam(r, "orderId")

	payment, err := h.service.GetByOrderID(ctx, orderID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get payment").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payment,
	})
}

// GetMyPayments gets payments for current user
func (h *Handler) GetMyPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		errors.New(errors.ErrUnauthorized, "User not authenticated").WriteHTTPResponse(w)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	payments, err := h.service.GetByUserID(ctx, claims.UserID, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get payments").WriteHTTPResponse(w)
		return
	}

	count, _ := h.service.CountByUserID(ctx, claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payments,
		"meta": map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// List lists all payments (admin only)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	payments, err := h.service.List(ctx, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to list payments").WriteHTTPResponse(w)
		return
	}

	count, _ := h.service.Count(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payments,
		"meta": map[string]interface{}{
			"total":  count,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// Process processes a payment
func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body
		req = ProcessPaymentRequest{}
	}

	payment, err := h.service.Process(ctx, id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to process payment").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payment,
	})
}

// Refund refunds a payment
func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body for full refund
		req = RefundRequest{}
	}

	payment, err := h.service.Refund(ctx, id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to refund payment").WriteHTTPResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": payment,
	})
}

// HealthCheck handles health check
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Health check called")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "payment-service",
	})
}
