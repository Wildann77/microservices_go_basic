package graph

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/microservices-go/gateway/internal/order"
	"github.com/microservices-go/gateway/internal/payment"
	"github.com/microservices-go/gateway/internal/user"
)

// LoaderKey is the context key for loaders
type LoaderKey struct{}

// Loaders holds all dataloaders
type Loaders struct {
	UserLoader           *dataloader.Loader[string, *user.User]
	OrderLoader          *dataloader.Loader[string, *order.Order]
	PaymentLoader        *dataloader.Loader[string, *payment.Payment]
	PaymentByOrderLoader *dataloader.Loader[string, *payment.Payment]
}

// NewLoaders creates new dataloaders with batch functions
func NewLoaders(userServiceURL, orderServiceURL, paymentServiceURL string) *Loaders {
	return &Loaders{
		UserLoader:           dataloader.NewBatchedLoader(user.BatchLoadUsers(userServiceURL), dataloader.WithWait[string, *user.User](time.Millisecond*5)),
		OrderLoader:          dataloader.NewBatchedLoader(order.BatchLoadOrders(orderServiceURL), dataloader.WithWait[string, *order.Order](time.Millisecond*5)),
		PaymentLoader:        dataloader.NewBatchedLoader(payment.BatchLoadPayments(paymentServiceURL), dataloader.WithWait[string, *payment.Payment](time.Millisecond*5)),
		PaymentByOrderLoader: dataloader.NewBatchedLoader(payment.BatchLoadPaymentsByOrder(paymentServiceURL), dataloader.WithWait[string, *payment.Payment](time.Millisecond*5)),
	}
}

// DataLoaderMiddleware injects dataloaders into context
func DataLoaderMiddleware(userServiceURL, orderServiceURL, paymentServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := NewLoaders(userServiceURL, orderServiceURL, paymentServiceURL)
			ctx := context.WithValue(r.Context(), LoaderKey{}, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLoaders retrieves dataloaders from context
func GetLoaders(ctx context.Context) *Loaders {
	loaders, _ := ctx.Value(LoaderKey{}).(*Loaders)
	return loaders
}

// GetServiceURLs returns service URLs from environment
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
