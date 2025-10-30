// Package repository defines interfaces for data persistence operations.
package repository

import (
	"context"

	"github.com/williajm/curly/internal/domain"
)

// RequestRepository defines operations for persisting and retrieving HTTP requests.
// Implementations should handle serialization of complex fields (headers, auth config).
// and ensure proper transactional semantics where appropriate.
type RequestRepository interface {
	// Create persists a new request to the repository.
	// Returns an error if the request already exists or validation fails.
	Create(ctx context.Context, req *domain.Request) error

	// FindByID retrieves a request by its unique identifier.
	// Returns ErrNotFound if the request does not exist.
	FindByID(ctx context.Context, id string) (*domain.Request, error)

	// FindAll retrieves all saved requests.
	// Results are ordered by created_at descending (newest first).
	FindAll(ctx context.Context) ([]*domain.Request, error)

	// Update modifies an existing request.
	// Returns ErrNotFound if the request does not exist.
	Update(ctx context.Context, req *domain.Request) error

	// Delete removes a request from the repository.
	// Returns ErrNotFound if the request does not exist.
	Delete(ctx context.Context, id string) error
}

// HistoryEntry represents a single execution of an HTTP request.
// It captures the response details and any errors that occurred.
type HistoryEntry struct {
	// ID is a unique identifier for this history entry.
	ID string

	// RequestID links this history entry to a saved request (may be empty for ad-hoc requests).
	RequestID string

	// ExecutedAt is when the request was executed.
	ExecutedAt string

	// StatusCode is the HTTP status code received.
	StatusCode int

	// Status is the HTTP status text.
	Status string

	// ResponseTimeMs is how long the request took in milliseconds.
	ResponseTimeMs int64

	// ResponseHeaders contains the response headers as JSON.
	ResponseHeaders string

	// ResponseBody is the response body content.
	ResponseBody string

	// Error contains the error message if the request failed, empty on success.
	Error string
}

// HistoryRepository defines operations for persisting and retrieving request execution history.
type HistoryRepository interface {
	// Save persists a history entry to the repository.
	// This records the result of executing a request.
	Save(ctx context.Context, entry *HistoryEntry) error

	// FindByID retrieves a history entry by its unique identifier.
	// Returns ErrNotFound if the entry does not exist.
	FindByID(ctx context.Context, id string) (*HistoryEntry, error)

	// FindAll retrieves all history entries.
	// Results are ordered by executed_at descending (newest first).
	// Limit controls the maximum number of entries returned (0 = unlimited).
	FindAll(ctx context.Context, limit int) ([]*HistoryEntry, error)

	// FindByRequestID retrieves all history entries for a specific request.
	// Results are ordered by executed_at descending (newest first).
	// Limit controls the maximum number of entries returned (0 = unlimited).
	FindByRequestID(ctx context.Context, requestID string, limit int) ([]*HistoryEntry, error)

	// Delete removes a history entry from the repository.
	// Returns ErrNotFound if the entry does not exist.
	Delete(ctx context.Context, id string) error

	// DeleteOlderThan removes all history entries older than the specified timestamp.
	// Returns the number of entries deleted.
	DeleteOlderThan(ctx context.Context, timestamp string) (int64, error)
}
