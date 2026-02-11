package order

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

func (c *Client) CreateOrder(ctx context.Context, input model.CreateOrderInput, userID string) (*model.Order, error) {
	// Build request body
	items := make([]map[string]interface{}, len(input.Items))
	for i, item := range input.Items {
		items[i] = map[string]interface{}{
			"product_id":   item.ProductID,
			"product_name": item.ProductName,
			"quantity":     item.Quantity,
			"unit_price":   item.UnitPrice,
		}
	}

	body := map[string]interface{}{
		"user_id":          userID,
		"currency":         input.Currency,
		"shipping_address": input.ShippingAddress,
		"notes":            input.Notes,
		"items":            items,
	}

	url := fmt.Sprintf("%s/api/v1/orders", c.url)
	resp, err := c.MakeRequest(ctx, "POST", url, body, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("failed to create order: %v", errResp)
	}

	var result struct {
		Data *model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) UpdateStatus(ctx context.Context, id string, status string) (*model.Order, error) {
	body := map[string]interface{}{"status": status}

	url := fmt.Sprintf("%s/api/v1/orders/%s/status", c.url, id)
	resp, err := c.MakeRequest(ctx, "PATCH", url, body, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update order status")
	}

	var result struct {
		Data *model.Order `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) ListOrders(ctx context.Context, limit, offset int) (*model.OrderConnection, error) {
	url := fmt.Sprintf("%s/api/v1/orders?limit=%d&offset=%d", c.url, limit, offset)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Order         `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &model.OrderConnection{
		Data: result.Data,
		PageInfo: model.PageInfo{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}, nil
}

func (c *Client) ListMyOrders(ctx context.Context, limit, offset int) (*model.OrderConnection, error) {
	url := fmt.Sprintf("%s/api/v1/orders/my-orders?limit=%d&offset=%d", c.url, limit, offset)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Order         `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &model.OrderConnection{
		Data: result.Data,
		PageInfo: model.PageInfo{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}, nil
}
