package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes a generic response with the specified status code
func JSON[T any](w http.ResponseWriter, data T, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := Response[T]{Data: data}
	return json.NewEncoder(w).Encode(resp)
}

// OK writes a 200 OK response with single data
func OK[T any](w http.ResponseWriter, data T) error {
	return JSON(w, data, http.StatusOK)
}

// Created writes a 201 Created response
func Created[T any](w http.ResponseWriter, data T) error {
	return JSON(w, data, http.StatusCreated)
}

// List writes a paginated list response with 200 OK
func List[T any](w http.ResponseWriter, data []T, meta Meta) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := ListResponse[T]{Data: data, Meta: meta}
	return json.NewEncoder(w).Encode(resp)
}

// Batch writes a batch response with 200 OK
func Batch[T any](w http.ResponseWriter, data []T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := BatchResponse[T]{Data: data, Count: len(data)}
	return json.NewEncoder(w).Encode(resp)
}

// Deleted writes a delete response with custom message
func Deleted(w http.ResponseWriter, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := DeleteResponse{Message: message}
	return json.NewEncoder(w).Encode(resp)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
