package graph

import (
	"context"
	"net/http"
	"os"

	"github.com/microservices-go/gateway/internal/order"
	"github.com/microservices-go/gateway/internal/payment"
	"github.com/microservices-go/gateway/internal/user"
)

type LoaderKey struct{}

type Loaders struct {
	UserLoader    *user.Loader
	OrderLoader   *order.Loader
	PaymentLoader *payment.Loader
}

func CreateLoaders(userServiceURL, orderServiceURL, paymentServiceURL string) *Loaders {
	return &Loaders{
		UserLoader:    user.NewLoader(userServiceURL),
		OrderLoader:   order.NewLoader(orderServiceURL),
		PaymentLoader: payment.NewLoader(paymentServiceURL),
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
	loaders, _ := ctx.Value(LoaderKey{}).(*Loaders)
	return loaders
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
