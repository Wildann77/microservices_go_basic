package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/microservices-go/gateway/graph/model"
)

// UserLoader batches user requests
type UserLoader struct {
	userServiceURL string
	client         *http.Client
	cache          map[string]*model.User
}

// NewUserLoader creates a new user loader
func NewUserLoader(userServiceURL string) *UserLoader {
	return &UserLoader{
		userServiceURL: userServiceURL,
		client:         &http.Client{Timeout: 5 * time.Second},
		cache:          make(map[string]*model.User),
	}
}

// Load loads a user by ID
func (l *UserLoader) Load(ctx context.Context, userID string) (*model.User, error) {
	// Check cache
	if user, ok := l.cache[userID]; ok {
		return user, nil
	}

	// Fetch from service
	url := fmt.Sprintf("%s/api/v1/users/%s", l.userServiceURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Forward auth header
	if auth := ctx.Value("Authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	var result struct {
		Data *model.User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Cache result
	l.cache[userID] = result.Data

	return result.Data, nil
}

// LoadMany loads multiple users
func (l *UserLoader) LoadMany(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	users := make([]*model.User, len(userIDs))
	errors := make([]error, len(userIDs))

	// For simplicity, load one by one
	// In production, implement batch endpoint
	for i, id := range userIDs {
		user, err := l.Load(ctx, id)
		users[i] = user
		errors[i] = err
	}

	return users, errors
}

// OrderLoader batches order requests
type OrderLoader struct {
	orderServiceURL string
	client          *http.Client
	cache           map[string][]*model.Order
}

// NewOrderLoader creates a new order loader
func NewOrderLoader(orderServiceURL string) *OrderLoader {
	return &OrderLoader{
		orderServiceURL: orderServiceURL,
		client:          &http.Client{Timeout: 5 * time.Second},
		cache:           make(map[string][]*model.Order),
	}
}

// LoadByUser loads orders by user ID
func (l *OrderLoader) LoadByUser(ctx context.Context, userID string, limit, offset int) ([]*model.Order, error) {
	cacheKey := fmt.Sprintf("%s_%d_%d", userID, limit, offset)
	if orders, ok := l.cache[cacheKey]; ok {
		return orders, nil
	}

	url := fmt.Sprintf("%s/api/v1/orders?user_id=%s&limit=%d&offset=%d", l.orderServiceURL, userID, limit, offset)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth := ctx.Value("Authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load orders for user: %s", userID)
	}

	var result struct {
		Data []*model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	l.cache[cacheKey] = result.Data

	return result.Data, nil
}

// LoadByID loads order by ID
func (l *OrderLoader) LoadByID(ctx context.Context, orderID string) (*model.Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", l.orderServiceURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth := ctx.Value("Authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	var result struct {
		Data *model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// PaymentLoader batches payment requests
type PaymentLoader struct {
	paymentServiceURL string
	client            *http.Client
	cache             map[string]*model.Payment
}

// NewPaymentLoader creates a new payment loader
func NewPaymentLoader(paymentServiceURL string) *PaymentLoader {
	return &PaymentLoader{
		paymentServiceURL: paymentServiceURL,
		client:            &http.Client{Timeout: 5 * time.Second},
		cache:             make(map[string]*model.Payment),
	}
}

// LoadByOrder loads payment by order ID
func (l *PaymentLoader) LoadByOrder(ctx context.Context, orderID string) (*model.Payment, error) {
	if payment, ok := l.cache[orderID]; ok {
		return payment, nil
	}

	url := fmt.Sprintf("%s/api/v1/payments/order/%s", l.paymentServiceURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth := ctx.Value("Authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Payment may not exist yet
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load payment for order: %s", orderID)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	l.cache[orderID] = result.Data

	return result.Data, nil
}

// LoadByID loads payment by ID
func (l *PaymentLoader) LoadByID(ctx context.Context, paymentID string) (*model.Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s", l.paymentServiceURL, paymentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth := ctx.Value("Authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment not found: %s", paymentID)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateLoaders creates all dataloaders
func CreateLoaders(userServiceURL, orderServiceURL, paymentServiceURL string) *model.Loaders {
	return &model.Loaders{
		UserLoader:    NewUserLoader(userServiceURL),
		OrderLoader:   NewOrderLoader(orderServiceURL),
		PaymentLoader: NewPaymentLoader(paymentServiceURL),
	}
}

// Middleware creates a middleware that adds dataloaders to context
func DataLoaderMiddleware(userServiceURL, orderServiceURL, paymentServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := CreateLoaders(userServiceURL, orderServiceURL, paymentServiceURL)

			// Extract auth header
			authHeader := r.Header.Get("Authorization")
			ctx := context.WithValue(r.Context(), "Authorization", authHeader)
			ctx = context.WithValue(ctx, model.LoaderKey{}, loaders)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper function to extract service URLs from environment
func GetServiceURLs() (string, string, string) {
	userURL := getEnv("USER_SERVICE_URL", "http://localhost:8081")
	orderURL := getEnv("ORDER_SERVICE_URL", "http://localhost:8082")
	paymentURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083")

	// Remove trailing slashes
	userURL = strings.TrimSuffix(userURL, "/")
	orderURL = strings.TrimSuffix(orderURL, "/")
	paymentURL = strings.TrimSuffix(paymentURL, "/")

	return userURL, orderURL, paymentURL
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}