package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/graph-gophers/dataloader/v7"
)

// BatchLoadOrders returns a batch function for loading orders
func BatchLoadOrders(orderServiceURL string) dataloader.BatchFunc[string, *Order] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*Order] {
		results := make([]*dataloader.Result[*Order], len(keys))

		if len(keys) == 0 {
			return results
		}

		// Prepare request body
		reqBody, _ := json.Marshal(map[string][]string{"ids": keys})

		url := fmt.Sprintf("%s/api/v1/orders/batch", orderServiceURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Order]{Error: err}
			}
			return results
		}

		req.Header.Set("Content-Type", "application/json")

		// Add auth header if present in ctx
		if auth, ok := ctx.Value("Authorization").(string); ok {
			req.Header.Set("Authorization", auth)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Order]{Error: err}
			}
			return results
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("failed to load orders: %d", resp.StatusCode)
			for i := range results {
				results[i] = &dataloader.Result[*Order]{Error: err}
			}
			return results
		}

		var result struct {
			Data []*Order `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Order]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup
		orderMap := make(map[string]*Order)
		for _, order := range result.Data {
			orderMap[order.ID] = order
		}

		// Map results to keys (maintain order)
		for i, key := range keys {
			if order, ok := orderMap[key]; ok {
				results[i] = &dataloader.Result[*Order]{Data: order}
			} else {
				results[i] = &dataloader.Result[*Order]{Error: fmt.Errorf("order not found: %s", key)}
			}
		}

		return results
	}
}
