package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	gatewayURL = "http://localhost:4000/query"
)

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func doGraphQLRequest(query string, variables map[string]interface{}, token string) (*GraphQLResponse, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})

	req, err := http.NewRequest("POST", gatewayURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", string(body))
	}

	return &gqlResp, nil
}

func TestAPIFlow(t *testing.T) {
	// Skip if gateway is not running
	_, err := http.Get("http://localhost:4000/health") // Assuming there's a health endpoint
	if err != nil {
		t.Skip("Gateway not running, skipping API integration test")
	}

	timestamp := time.Now().Unix()
	email := fmt.Sprintf("test_%d@example.com", timestamp)
	password := "password123"

	// 1. Register
	t.Run("Register", func(t *testing.T) {
		query := `
			mutation Register($input: RegisterInput!) {
				register(input: $input) {
					token
					user {
						id
						email
						firstName
					}
				}
			}
		`
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"email":     email,
				"password":  password,
				"firstName": "Test",
				"lastName":  "User",
			},
		}

		resp, err := doGraphQLRequest(query, variables, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		data := resp.Data.(map[string]interface{})["register"].(map[string]interface{})
		if data["token"] == "" {
			t.Fatal("Expected token, got empty")
		}
		user := data["user"].(map[string]interface{})
		if user["email"] != email {
			t.Errorf("Expected email %s, got %s", email, user["email"])
		}
	})

	// 2. Login
	var token string
	t.Run("Login", func(t *testing.T) {
		query := `
			mutation Login($input: LoginInput!) {
				login(input: $input) {
					token
					user {
						id
					}
				}
			}
		`
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"email":    email,
				"password": password,
			},
		}

		resp, err := doGraphQLRequest(query, variables, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		data := resp.Data.(map[string]interface{})["login"].(map[string]interface{})
		token = data["token"].(string)
		if token == "" {
			t.Fatal("Expected token, got empty")
		}
	})

	// 3. Get Me
	t.Run("GetMe", func(t *testing.T) {
		query := `
			query {
				me {
					id
					email
					fullName
				}
			}
		`
		resp, err := doGraphQLRequest(query, nil, token)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		user := resp.Data.(map[string]interface{})["me"].(map[string]interface{})
		if user["email"] != email {
			t.Errorf("Expected email %s, got %s", email, user["email"])
		}
	})
}
