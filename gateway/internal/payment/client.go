package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/microservices-go/gateway/graph/model"
	"github.com/microservices-go/gateway/internal/client"
)

type Client struct {
	*client.BaseClient
	url string
}

func NewClient(url string) *Client {
	return &Client{
		BaseClient: client.NewBaseClient(),
		url:        url,
	}
}

func (c *Client) CreatePayment(ctx context.Context, input model.CreatePaymentInput, userID string) (*model.Payment, error) {
	body := map[string]interface{}{
		"order_id":    input.OrderID,
		"user_id":     userID,
		"amount":      input.Amount,
		"currency":    input.Currency,
		"method":      input.Method,
		"description": input.Description,
	}

	url := fmt.Sprintf("%s/api/v1/payments", c.url)
	resp, err := c.MakeRequest(ctx, "POST", url, body, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("failed to create payment: %v", errResp)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) ProcessPayment(ctx context.Context, id string) (*model.Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s/process", c.url, id)
	resp, err := c.MakeRequest(ctx, "POST", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to process payment")
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) RefundPayment(ctx context.Context, id string, amount *float64, reason *string) (*model.Payment, error) {
	body := map[string]interface{}{}
	if amount != nil {
		body["amount"] = *amount
	}
	if reason != nil {
		body["reason"] = *reason
	}

	url := fmt.Sprintf("%s/api/v1/payments/%s/refund", c.url, id)
	resp, err := c.MakeRequest(ctx, "POST", url, body, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refund payment")
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) ListPayments(ctx context.Context, limit, offset int) (*model.PaymentConnection, error) {
	url := fmt.Sprintf("%s/api/v1/payments?limit=%d&offset=%d", c.url, limit, offset)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Payment       `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &model.PaymentConnection{
		Data: result.Data,
		PageInfo: model.PageInfo{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}, nil
}

func (c *Client) ListMyPayments(ctx context.Context, limit, offset int) (*model.PaymentConnection, error) {
	url := fmt.Sprintf("%s/api/v1/payments/my-payments?limit=%d&offset=%d", c.url, limit, offset)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Payment       `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &model.PaymentConnection{
		Data: result.Data,
		PageInfo: model.PageInfo{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}, nil
}
