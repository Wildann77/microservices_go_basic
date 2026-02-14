package response

import "encoding/json"

// Response is the standard API response wrapper for single resources
type Response[T any] struct {
	Data T `json:"data"`
}

// ListResponse is the standard API response wrapper for paginated lists
type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// BatchResponse is the standard API response wrapper for batch operations
type BatchResponse[T any] struct {
	Data  []T `json:"data"`
	Count int `json:"count"`
}

// DeleteResponse is the standard API response for delete operations
type DeleteResponse struct {
	Message string `json:"message"`
}

// ToJSON serializes the response to JSON bytes
func (r Response[T]) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes the list response to JSON bytes
func (r ListResponse[T]) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes the batch response to JSON bytes
func (r BatchResponse[T]) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON serializes the delete response to JSON bytes
func (r DeleteResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
