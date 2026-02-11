package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Loader batches user requests
type Loader struct {
	UserServiceURL string
	Client         *http.Client
	Cache          map[string]*User
}

// NewLoader creates a new user loader
func NewLoader(userServiceURL string) *Loader {
	return &Loader{
		UserServiceURL: userServiceURL,
		Client:         &http.Client{Timeout: 5 * time.Second},
		Cache:          make(map[string]*User),
	}
}

// Load loads a user by ID
func (l *Loader) Load(ctx context.Context, userID string) (*User, error) {
	if user, ok := l.Cache[userID]; ok {
		return user, nil
	}

	url := fmt.Sprintf("%s/api/v1/users/%s", l.UserServiceURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add auth header if present in ctx
	if auth, ok := ctx.Value("Authorization").(string); ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load user: %d", resp.StatusCode)
	}

	var result struct {
		Data *User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	l.Cache[userID] = result.Data
	return result.Data, nil
}

// LoadMany loads multiple users by IDs
func (l *Loader) LoadMany(ctx context.Context, userIDs []string) ([]*User, []error) {
	users := make([]*User, len(userIDs))
	errors := make([]error, len(userIDs))

	for i, id := range userIDs {
		user, err := l.Load(ctx, id)
		if err != nil {
			errors[i] = err
		} else {
			users[i] = user
		}
	}

	return users, errors
}
