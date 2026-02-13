package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/graph-gophers/dataloader/v7"
)

// BatchLoadUsers returns a batch function for loading users
func BatchLoadUsers(userServiceURL string) dataloader.BatchFunc[string, *User] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*User] {
		results := make([]*dataloader.Result[*User], len(keys))

		if len(keys) == 0 {
			return results
		}

		// Prepare request body
		reqBody, _ := json.Marshal(map[string][]string{"ids": keys})

		url := fmt.Sprintf("%s/api/v1/users/batch", userServiceURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*User]{Error: err}
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
				results[i] = &dataloader.Result[*User]{Error: err}
			}
			return results
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("failed to load users: %d", resp.StatusCode)
			for i := range results {
				results[i] = &dataloader.Result[*User]{Error: err}
			}
			return results
		}

		var result struct {
			Data []*User `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			for i := range results {
				results[i] = &dataloader.Result[*User]{Error: err}
			}
			return results
		}

		// Create map for O(1) lookup
		userMap := make(map[string]*User)
		for _, user := range result.Data {
			userMap[user.ID] = user
		}

		// Map results to keys (maintain order)
		for i, key := range keys {
			if user, ok := userMap[key]; ok {
				results[i] = &dataloader.Result[*User]{Data: user}
			} else {
				results[i] = &dataloader.Result[*User]{Error: fmt.Errorf("user not found: %s", key)}
			}
		}

		return results
	}
}
