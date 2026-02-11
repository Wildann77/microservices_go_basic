package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Resolver is the root resolver
type Resolver struct {
	userServiceURL    string
	orderServiceURL   string
	paymentServiceURL string
	client            *http.Client
}

// NewResolver creates a new resolver
func NewResolver(userServiceURL, orderServiceURL, paymentServiceURL string) *Resolver {
	return &Resolver{
		userServiceURL:    userServiceURL,
		orderServiceURL:   orderServiceURL,
		paymentServiceURL: paymentServiceURL,
		client:            &http.Client{Timeout: 10 * time.Second},
	}
}

// Helper methods
func (r *Resolver) makeRequest(ctx context.Context, method, url string, body interface{}, authHeader string) (*http.Response, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	return r.client.Do(req)
}

func (r *Resolver) getAuthHeader(ctx context.Context) string {
	if auth := ctx.Value("Authorization"); auth != nil {
		return auth.(string)
	}
	return ""
}
