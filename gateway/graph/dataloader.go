package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/microservices-go/gateway/graph/model"
)

type LoaderKey struct{}

type Loaders struct {
	UserLoader    *UserLoader
	OrderLoader   *OrderLoader
	PaymentLoader *PaymentLoader
}

type UserLoader struct {
	UserServiceURL string
	Client         *http.Client
	Cache          map[string]*model.User
}

func NewUserLoader(userServiceURL string) *UserLoader {
	return &UserLoader{
		UserServiceURL: userServiceURL,
		Client:         &http.Client{Timeout: 5 * time.Second},
		Cache:          make(map[string]*model.User),
	}
}

func (l *UserLoader) Load(ctx context.Context, userID string) (*model.User, error) {
	if user, ok := l.Cache[userID]; ok {
		return user, nil
	}

	url := fmt.Sprintf("%s/api/v1/users/%s", l.UserServiceURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add auth header if present in ctx
	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load user: %d", resp.StatusCode)
	}

	var result struct {
		Data *model.User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	l.Cache[userID] = result.Data
	return result.Data, nil
}

func (l *UserLoader) LoadMany(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	users := make([]*model.User, len(userIDs))
	errors := make([]error, len(userIDs))

	for i, id := range userIDs {
		user, err := l.Load(ctx, id)
		if err != nil {
			errors[i] = err
		} else {
			users[i] = user
		}
	}

	return users, errors
}

type OrderLoader struct {
	OrderServiceURL string
	Client          *http.Client
}

func NewOrderLoader(orderServiceURL string) *OrderLoader {
	return &OrderLoader{
		OrderServiceURL: orderServiceURL,
		Client:          &http.Client{Timeout: 5 * time.Second},
	}
}

func (l *OrderLoader) LoadByID(ctx context.Context, orderID string) (*model.Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", l.OrderServiceURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load order: %d", resp.StatusCode)
	}

	var result struct {
		Data *model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (l *OrderLoader) LoadByUser(ctx context.Context, userID string, limit, offset int) ([]*model.Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders/user/%s?limit=%d&offset=%d", l.OrderServiceURL, userID, limit, offset)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

type PaymentLoader struct {
	PaymentServiceURL string
	Client            *http.Client
}

func NewPaymentLoader(paymentServiceURL string) *PaymentLoader {
	return &PaymentLoader{
		PaymentServiceURL: paymentServiceURL,
		Client:            &http.Client{Timeout: 5 * time.Second},
	}
}

func (l *PaymentLoader) LoadByID(ctx context.Context, paymentID string) (*model.Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s", l.PaymentServiceURL, paymentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load payment: %d", resp.StatusCode)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (l *PaymentLoader) LoadByOrder(ctx context.Context, orderID string) (*model.Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/order/%s", l.PaymentServiceURL, orderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load payment by order: %d", resp.StatusCode)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func CreateLoaders(userServiceURL, orderServiceURL, paymentServiceURL string) *Loaders {
	return &Loaders{
		UserLoader:    NewUserLoader(userServiceURL),
		OrderLoader:   NewOrderLoader(orderServiceURL),
		PaymentLoader: NewPaymentLoader(paymentServiceURL),
	}
}

func DataLoaderMiddleware(userServiceURL, orderServiceURL, paymentServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := CreateLoaders(userServiceURL, orderServiceURL, paymentServiceURL)
			ctx := context.WithValue(r.Context(), LoaderKey{}, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(LoaderKey{}).(*Loaders)
}

func GetServiceURLs() (string, string, string) {
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:4001")
	orderServiceURL := getEnv("ORDER_SERVICE_URL", "http://localhost:4002")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:4003")

	return userServiceURL, orderServiceURL, paymentServiceURL
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
