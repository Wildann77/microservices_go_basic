package payment

import (
	"context"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
)

// StripeProvider implements PaymentProvider for Stripe
type StripeProvider struct {
	secretKey string
}

// NewStripeProvider creates a new Stripe provider
func NewStripeProvider() *StripeProvider {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		key = "sk_test_dummy_key" // For development
	}
	stripe.Key = key

	return &StripeProvider{
		secretKey: key,
	}
}

// CreatePaymentIntent creates a Stripe payment intent
func (s *StripeProvider) CreatePaymentIntent(ctx context.Context, amount float64, currency string) (*PaymentIntentResult, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)), // Convert to cents
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	return &PaymentIntentResult{
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
	}, nil
}

// ConfirmPayment confirms a Stripe payment
func (s *StripeProvider) ConfirmPayment(ctx context.Context, paymentIntentID string) (*PaymentResult, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return &PaymentResult{
			Success:       false,
			FailureReason: err.Error(),
		}, nil
	}

	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		return &PaymentResult{
			Success:       true,
			TransactionID: pi.ID,
		}, nil
	}

	return &PaymentResult{
		Success:       false,
		FailureReason: string(pi.Status),
	}, nil
}

// Refund processes a Stripe refund
func (s *StripeProvider) Refund(ctx context.Context, transactionID string, amount float64) (*RefundResult, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(transactionID),
	}

	if amount > 0 {
		params.Amount = stripe.Int64(int64(amount * 100))
	}

	ref, err := refund.New(params)
	if err != nil {
		return &RefundResult{
			Success:       false,
			FailureReason: err.Error(),
		}, nil
	}

	if ref.Status == stripe.RefundStatusSucceeded {
		return &RefundResult{
			Success:  true,
			RefundID: ref.ID,
		}, nil
	}

	return &RefundResult{
		Success:       false,
		FailureReason: string(ref.Status),
	}, nil
}
