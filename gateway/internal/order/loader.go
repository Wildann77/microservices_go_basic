package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Loader batches order requests
type Loader struct {
	OrderServiceURL string
	Client          *http.Client
}

// NewLoader creates a new order loader
func NewLoader(orderServiceURL string) *Loader {
	return &Loader{
		OrderServiceURL: orderServiceURL,
		Client:          &http.Client{Timeout: 5 * time.Second},
	}
}

// LoadByID loads an order by ID
func (l *Loader) LoadByID(ctx context.Context, orderID string) (*Order, error) {
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
		Data *Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// LoadByUser loads orders for a specific user
func (l *Loader) LoadByUser(ctx context.Context, userID string, limit, offset int) ([]*Order, error) {
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
		Data []*Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
