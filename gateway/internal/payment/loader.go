package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/graph-gophers/dataloader/v7"
)

// BatchLoadPayments returns a batch function for loading payments by ID
func BatchLoadPayments(paymentServiceURL string) dataloader.BatchFunc[string, *Payment] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*Payment] {
		results := make([]*dataloader.Result[*Payment], len(keys))

		if len(keys) == 0 {
			return results
		}

		// Prepare request body
		reqBody, _ := json.Marshal(map[string][]string{"ids": keys})

		url := fmt.Sprintf("%s/api/v1/payments/batch", paymentServiceURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
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
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("failed to load payments: %d", resp.StatusCode)
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}

		var result struct {
			Data []*Payment `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup
		paymentMap := make(map[string]*Payment)
		for _, payment := range result.Data {
			paymentMap[payment.ID] = payment
		}

		// Map results to keys (maintain order)
		for i, key := range keys {
			if payment, ok := paymentMap[key]; ok {
				results[i] = &dataloader.Result[*Payment]{Data: payment}
			} else {
				results[i] = &dataloader.Result[*Payment]{Error: fmt.Errorf("payment not found: %s", key)}
			}
		}

		return results
	}
}

// BatchLoadPaymentsByOrder returns a batch function for loading payments by order ID
func BatchLoadPaymentsByOrder(paymentServiceURL string) dataloader.BatchFunc[string, *Payment] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*Payment] {
		results := make([]*dataloader.Result[*Payment], len(keys))

		if len(keys) == 0 {
			return results
		}

		// Prepare request body
		reqBody, _ := json.Marshal(map[string][]string{"order_ids": keys})

		url := fmt.Sprintf("%s/api/v1/payments/batch-by-order", paymentServiceURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
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
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("failed to load payments by order: %d", resp.StatusCode)
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}

		var result struct {
			Data []*Payment `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*Payment]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup (by order ID)
		paymentMap := make(map[string]*Payment)
		for _, payment := range result.Data {
			paymentMap[payment.OrderID] = payment
		}

		// Map results to keys (maintain order)
		for i, key := range keys {
			if payment, ok := paymentMap[key]; ok {
				results[i] = &dataloader.Result[*Payment]{Data: payment}
			} else {
				// Payment might not exist for order
				results[i] = &dataloader.Result[*Payment]{Data: nil}
			}
		}

		return results
	}
}
