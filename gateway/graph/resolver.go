package graph

import (
	"github.com/microservices-go/gateway/internal/order"
	"github.com/microservices-go/gateway/internal/payment"
	"github.com/microservices-go/gateway/internal/user"
)

// Resolver is the root resolver
type Resolver struct {
	UserClient    *user.Client
	OrderClient   *order.Client
	PaymentClient *payment.Client
}

// NewResolver creates a new resolver
func NewResolver(userServiceURL, orderServiceURL, paymentServiceURL string) *Resolver {
	return &Resolver{
		UserClient:    user.NewClient(userServiceURL),
		OrderClient:   order.NewClient(orderServiceURL),
		PaymentClient: payment.NewClient(paymentServiceURL),
	}
}
