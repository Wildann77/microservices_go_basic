package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Loader batches payment requests
type Loader struct {
	PaymentServiceURL string
	Client            *http.Client
}

// NewLoader creates a new payment loader
func NewLoader(paymentServiceURL string) *Loader {
	return &Loader{
		PaymentServiceURL: paymentServiceURL,
		Client:            &http.Client{Timeout: 5 * time.Second},
	}
}

// LoadByID loads a payment by ID
func (l *Loader) LoadByID(ctx context.Context, paymentID string) (*Payment, error) {
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
		Data *Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// LoadByOrder loads a payment for a specific order
func (l *Loader) LoadByOrder(ctx context.Context, orderID string) (*Payment, error) {
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
		Data *Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
