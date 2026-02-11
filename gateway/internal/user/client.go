package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/microservices-go/gateway/internal/client"
	"github.com/microservices-go/gateway/internal/common"
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

func (c *Client) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/register", c.url)
	resp, err := c.MakeRequest(ctx, "POST", url, input, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("registration failed: %v", errResp)
	}

	var result struct {
		Data *AuthResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/login", c.url)
	resp, err := c.MakeRequest(ctx, "POST", url, input, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("login failed: %v", errResp)
	}

	var result struct {
		Data *AuthResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, firstName *string, lastName *string, isActive *bool) (*User, error) {
	body := map[string]interface{}{}
	if firstName != nil {
		body["first_name"] = *firstName
	}
	if lastName != nil {
		body["last_name"] = *lastName
	}
	if isActive != nil {
		body["is_active"] = *isActive
	}

	url := fmt.Sprintf("%s/api/v1/users/%s", c.url, id)
	resp, err := c.MakeRequest(ctx, "PUT", url, body, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update user")
	}

	var result struct {
		Data *User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.url, id)
	resp, err := c.MakeRequest(ctx, "DELETE", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusNoContent, nil
}

func (c *Client) Me(ctx context.Context) (*User, error) {
	url := fmt.Sprintf("%s/api/v1/users/me", c.url)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get current user")
	}

	var result struct {
		Data *User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) ListUsers(ctx context.Context, limit, offset int) (*UserConnection, error) {
	url := fmt.Sprintf("%s/api/v1/users?limit=%d&offset=%d", c.url, limit, offset)
	resp, err := c.MakeRequest(ctx, "GET", url, nil, client.GetAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*User                `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &UserConnection{
		Data: result.Data,
		PageInfo: common.PageInfo{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}, nil
}
