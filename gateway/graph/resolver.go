package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/microservices-go/gateway/graph/model"
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

// Query returns QueryResolver implementation
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation returns MutationResolver implementation
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// User returns UserResolver implementation
func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

// Order returns OrderResolver implementation
func (r *Resolver) Order() OrderResolver {
	return &orderResolver{r}
}

// Payment returns PaymentResolver implementation
func (r *Resolver) Payment() PaymentResolver {
	return &paymentResolver{r}
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

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
type orderResolver struct{ *Resolver }
type paymentResolver struct{ *Resolver }

// Query resolvers
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	url := fmt.Sprintf("%s/api/v1/users/me", r.userServiceURL)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get current user")
	}

	var result struct {
		Data *model.User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.UserLoader.Load(ctx, id)
}

func (r *queryResolver) Users(ctx context.Context, limit *int, offset *int) (*model.UserConnection, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	url := fmt.Sprintf("%s/api/v1/users?limit=%d&offset=%d", r.userServiceURL, l, o)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.User      `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	total := 0
	if t, ok := result.Meta["total"].(float64); ok {
		total = int(t)
	}

	return &model.UserConnection{
		Data: result.Data,
		PageInfo: model.PageInfo{
			Total:   total,
			Limit:   l,
			Offset:  o,
			HasMore: o+l < total,
		},
	}, nil
}

func (r *queryResolver) Order(ctx context.Context, id string) (*model.Order, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.OrderLoader.LoadByID(ctx, id)
}

func (r *queryResolver) Orders(ctx context.Context, limit *int, offset *int) (*model.OrderConnection, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	url := fmt.Sprintf("%s/api/v1/orders?limit=%d&offset=%d", r.orderServiceURL, l, o)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Order     `json:"data"`
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
			Limit:   l,
			Offset:  o,
			HasMore: o+l < total,
		},
	}, nil
}

func (r *queryResolver) MyOrders(ctx context.Context, limit *int, offset *int) (*model.OrderConnection, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	url := fmt.Sprintf("%s/api/v1/orders/my-orders?limit=%d&offset=%d", r.orderServiceURL, l, o)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Order     `json:"data"`
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
			Limit:   l,
			Offset:  o,
			HasMore: o+l < total,
		},
	}, nil
}

func (r *queryResolver) Payment(ctx context.Context, id string) (*model.Payment, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.PaymentLoader.LoadByID(ctx, id)
}

func (r *queryResolver) PaymentByOrder(ctx context.Context, orderID string) (*model.Payment, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.PaymentLoader.LoadByOrder(ctx, orderID)
}

func (r *queryResolver) Payments(ctx context.Context, limit *int, offset *int) (*model.PaymentConnection, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	url := fmt.Sprintf("%s/api/v1/payments?limit=%d&offset=%d", r.paymentServiceURL, l, o)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Payment   `json:"data"`
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
			Limit:   l,
			Offset:  o,
			HasMore: o+l < total,
		},
	}, nil
}

func (r *queryResolver) MyPayments(ctx context.Context, limit *int, offset *int) (*model.PaymentConnection, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	url := fmt.Sprintf("%s/api/v1/payments/my-payments?limit=%d&offset=%d", r.paymentServiceURL, l, o)
	resp, err := r.makeRequest(ctx, "GET", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []*model.Payment   `json:"data"`
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
			Limit:   l,
			Offset:  o,
			HasMore: o+l < total,
		},
	}, nil
}

// Mutation resolvers
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.AuthResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/register", r.userServiceURL)
	resp, err := r.makeRequest(ctx, "POST", url, input, "")
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
		Data *model.AuthResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/login", r.userServiceURL)
	resp, err := r.makeRequest(ctx, "POST", url, input, "")
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
		Data *model.AuthResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id string, firstName, lastName *string, isActive *bool) (*model.User, error) {
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

	url := fmt.Sprintf("%s/api/v1/users/%s", r.userServiceURL, id)
	resp, err := r.makeRequest(ctx, "PUT", url, body, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update user")
	}

	var result struct {
		Data *model.User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", r.userServiceURL, id)
	resp, err := r.makeRequest(ctx, "DELETE", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusNoContent, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, input model.CreateOrderInput) (*model.Order, error) {
	// Get current user
	me, err := r.Query().Me(ctx)
	if err != nil {
		return nil, err
	}

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
		"user_id":          me.ID,
		"currency":         input.Currency,
		"shipping_address": input.ShippingAddress,
		"notes":            input.Notes,
		"items":            items,
	}

	url := fmt.Sprintf("%s/api/v1/orders", r.orderServiceURL)
	resp, err := r.makeRequest(ctx, "POST", url, body, r.getAuthHeader(ctx))
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

func (r *mutationResolver) UpdateOrderStatus(ctx context.Context, id string, status string) (*model.Order, error) {
	body := map[string]interface{}{"status": status}

	url := fmt.Sprintf("%s/api/v1/orders/%s/status", r.orderServiceURL, id)
	resp, err := r.makeRequest(ctx, "PATCH", url, body, r.getAuthHeader(ctx))
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

func (r *mutationResolver) CreatePayment(ctx context.Context, input model.CreatePaymentInput) (*model.Payment, error) {
	// Get current user
	me, err := r.Query().Me(ctx)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"order_id":    input.OrderID,
		"user_id":     me.ID,
		"amount":      input.Amount,
		"currency":    input.Currency,
		"method":      input.Method,
		"description": input.Description,
	}

	url := fmt.Sprintf("%s/api/v1/payments", r.paymentServiceURL)
	resp, err := r.makeRequest(ctx, "POST", url, body, r.getAuthHeader(ctx))
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

func (r *mutationResolver) ProcessPayment(ctx context.Context, id string) (*model.Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s/process", r.paymentServiceURL, id)
	resp, err := r.makeRequest(ctx, "POST", url, nil, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("failed to process payment: %v", errResp)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (r *mutationResolver) RefundPayment(ctx context.Context, id string, amount *float64, reason *string) (*model.Payment, error) {
	body := map[string]interface{}{}
	if amount != nil {
		body["amount"] = *amount
	}
	if reason != nil {
		body["reason"] = *reason
	}

	url := fmt.Sprintf("%s/api/v1/payments/%s/refund", r.paymentServiceURL, id)
	resp, err := r.makeRequest(ctx, "POST", url, body, r.getAuthHeader(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("failed to refund payment: %v", errResp)
	}

	var result struct {
		Data *model.Payment `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// User resolvers
func (r *userResolver) Orders(ctx context.Context, obj *model.User, limit *int, offset *int) ([]*model.Order, error) {
	l := 10
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	loaders := model.GetLoaders(ctx)
	return loaders.OrderLoader.LoadByUser(ctx, obj.ID, l, o)
}

// Order resolvers
func (r *orderResolver) User(ctx context.Context, obj *model.Order) (*model.User, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.UserLoader.Load(ctx, obj.UserID)
}

func (r *orderResolver) Payment(ctx context.Context, obj *model.Order) (*model.Payment, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.PaymentLoader.LoadByOrder(ctx, obj.ID)
}

// Payment resolvers
func (r *paymentResolver) Order(ctx context.Context, obj *model.Payment) (*model.Order, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.OrderLoader.LoadByID(ctx, obj.OrderID)
}

func (r *paymentResolver) User(ctx context.Context, obj *model.Payment) (*model.User, error) {
	loaders := model.GetLoaders(ctx)
	return loaders.UserLoader.Load(ctx, obj.UserID)
}