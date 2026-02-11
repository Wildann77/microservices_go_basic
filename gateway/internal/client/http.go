package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// BaseClient handles HTTP requests to microservices
type BaseClient struct {
	HTTPClient *http.Client
}

// NewBaseClient creates a new base client
func NewBaseClient() *BaseClient {
	return &BaseClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// MakeRequest makes an HTTP request and decodes the response
func (c *BaseClient) MakeRequest(ctx context.Context, method, url string, body interface{}, authHeader string) (*http.Response, error) {
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

	return c.HTTPClient.Do(req)
}

// GetAuthHeader retrieves auth header from context
func GetAuthHeader(ctx context.Context) string {
	if auth, ok := ctx.Value("Authorization").(string); ok {
		return auth
	}
	return ""
}
