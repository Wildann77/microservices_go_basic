package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
	"github.com/microservices-go/shared/response"
)

// Handler handles HTTP requests for user service
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/api/v1/users", func(r chi.Router) {
		// Public routes
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Get("/", h.List)
			r.Post("/batch", h.GetBatch)
			r.Get("/{id}", h.GetByID)
			r.Put("/{id}", h.Update)
			r.Delete("/{id}", h.Delete)
			r.Get("/me", h.GetMe)
		})
	})
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	user, err := h.service.Create(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to create user").WriteHTTPResponse(w)
		return
	}

	response.Created(w, user)
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	resp, err := h.service.Login(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Login failed").WriteHTTPResponse(w)
		return
	}

	response.OK(w, resp)
}

// GetByID gets user by ID
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	user, err := h.service.GetByID(ctx, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get user").WriteHTTPResponse(w)
		return
	}

	response.OK(w, user)
}

// GetMe gets current user
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		errors.New(errors.ErrUnauthorized, "User not authenticated").WriteHTTPResponse(w)
		return
	}

	user, err := h.service.GetByID(ctx, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get user").WriteHTTPResponse(w)
		return
	}

	response.OK(w, user)
}

// List lists all users
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, err := h.service.List(ctx, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to list users").WriteHTTPResponse(w)
		return
	}

	count, _ := h.service.Count(ctx)
	meta := response.NewMeta(count, limit, offset)

	response.List(w, users, meta)
}

// Update updates user
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.New(errors.ErrInvalidInput, "Invalid request body").WriteHTTPResponse(w)
		return
	}

	user, err := h.service.Update(ctx, id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to update user").WriteHTTPResponse(w)
		return
	}

	response.OK(w, user)
}

// Delete deletes user
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(ctx, id); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to delete user").WriteHTTPResponse(w)
		return
	}

	response.Deleted(w, "User deleted successfully")
}

// GetBatch gets multiple users by IDs
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

	users, err := h.service.GetByIDs(ctx, req.IDs)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteHTTPResponse(w)
			return
		}
		errors.New(errors.ErrInternalServer, "Failed to get users").WriteHTTPResponse(w)
		return
	}

	response.Batch(w, users)
}

// HealthCheck handles health check
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Health check called")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "user-service",
	})
}
