// Package app provides application-level services that orchestrate business logic.
//
// Services in this package coordinate between the domain layer (business rules)
// and infrastructure layer (persistence, HTTP, etc.) to implement use cases.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/williajm/curly/internal/domain"
	"github.com/williajm/curly/internal/infrastructure/http"
	"github.com/williajm/curly/internal/infrastructure/repository"
)

// RequestService orchestrates the full lifecycle of HTTP requests.
// It handles creation, validation, execution, persistence, and retrieval of requests.
type RequestService struct {
	repo        repository.RequestRepository
	httpClient  http.Client
	historyRepo repository.HistoryRepository
	logger      *slog.Logger
}

// NewRequestService creates a new RequestService with the provided dependencies.
// All dependencies are required and must not be nil.
func NewRequestService(
	repo repository.RequestRepository,
	httpClient http.Client,
	historyRepo repository.HistoryRepository,
	logger *slog.Logger,
) *RequestService {
	if repo == nil {
		panic("request repository cannot be nil")
	}
	if httpClient == nil {
		panic("http client cannot be nil")
	}
	if historyRepo == nil {
		panic("history repository cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &RequestService{
		repo:        repo,
		httpClient:  httpClient,
		historyRepo: historyRepo,
		logger:      logger,
	}
}

// CreateRequest creates a new request with validation.
// It generates a unique ID and sets timestamps.
// Returns an error if the request is invalid or cannot be persisted.
func (s *RequestService) CreateRequest(ctx context.Context, req *domain.Request) (*domain.Request, error) {
	// Ensure ID is set
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now

	// Validate request
	if err := req.Validate(); err != nil {
		s.logger.Warn("request validation failed",
			"request_id", req.ID,
			"error", err,
		)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	s.logger.Info("creating request",
		"request_id", req.ID,
		"name", req.Name,
		"method", req.Method,
		"url", req.URL,
	)

	return req, nil
}

// ExecuteRequest executes an HTTP request and returns the response.
// It validates the request, applies authentication, and captures timing metrics.
// The request is NOT saved to the repository - use ExecuteAndSave for that.
func (s *RequestService) ExecuteRequest(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// Validate request before execution
	if err := req.Validate(); err != nil {
		s.logger.Warn("request validation failed before execution",
			"request_id", req.ID,
			"error", err,
		)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	s.logger.Info("executing request",
		"request_id", req.ID,
		"method", req.Method,
		"url", req.URL,
	)

	// Execute HTTP request
	resp, err := s.httpClient.Execute(ctx, req)
	if err != nil {
		s.logger.Error("request execution failed",
			"request_id", req.ID,
			"method", req.Method,
			"url", req.URL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	s.logger.Info("request executed successfully",
		"request_id", req.ID,
		"status_code", resp.StatusCode,
		"duration_ms", resp.DurationMillis(),
	)

	return resp, nil
}

// SaveRequest persists a request to the repository.
// If the request already exists (by ID), it will be updated.
// Returns an error if the request cannot be saved.
func (s *RequestService) SaveRequest(ctx context.Context, req *domain.Request) error {
	// Validate before saving
	if err := req.Validate(); err != nil {
		s.logger.Warn("cannot save invalid request",
			"request_id", req.ID,
			"error", err,
		)
		return fmt.Errorf("invalid request: %w", err)
	}

	s.logger.Info("saving request",
		"request_id", req.ID,
		"name", req.Name,
	)

	// Check if request exists
	existing, err := s.repo.FindByID(ctx, req.ID)
	if err == nil && existing != nil {
		// Request exists, update it
		req.UpdatedAt = time.Now()
		if err := s.repo.Update(ctx, req); err != nil {
			s.logger.Error("failed to update request",
				"request_id", req.ID,
				"error", err,
			)
			return fmt.Errorf("failed to update request: %w", err)
		}
		s.logger.Info("request updated successfully", "request_id", req.ID)
		return nil
	}

	// Request doesn't exist, create it
	if err := s.repo.Create(ctx, req); err != nil {
		s.logger.Error("failed to create request",
			"request_id", req.ID,
			"error", err,
		)
		return fmt.Errorf("failed to save request: %w", err)
	}

	s.logger.Info("request saved successfully", "request_id", req.ID)
	return nil
}

// LoadRequest retrieves a saved request by ID.
// Returns an error if the request is not found.
func (s *RequestService) LoadRequest(ctx context.Context, id string) (*domain.Request, error) {
	s.logger.Debug("loading request", "request_id", id)

	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Warn("request not found",
			"request_id", id,
			"error", err,
		)
		return nil, fmt.Errorf("failed to load request: %w", err)
	}

	s.logger.Debug("request loaded successfully",
		"request_id", id,
		"name", req.Name,
	)

	return req, nil
}

// ListRequests retrieves all saved requests.
// Results are ordered by created_at descending (newest first).
func (s *RequestService) ListRequests(ctx context.Context) ([]*domain.Request, error) {
	s.logger.Debug("listing all requests")

	requests, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("failed to list requests", "error", err)
		return nil, fmt.Errorf("failed to list requests: %w", err)
	}

	s.logger.Debug("requests listed successfully", "count", len(requests))
	return requests, nil
}

// DeleteRequest removes a saved request by ID.
// Returns an error if the request is not found.
func (s *RequestService) DeleteRequest(ctx context.Context, id string) error {
	s.logger.Info("deleting request", "request_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete request",
			"request_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to delete request: %w", err)
	}

	s.logger.Info("request deleted successfully", "request_id", id)
	return nil
}

// ExecuteAndSave executes a request and saves the result to history.
// This is an atomic operation that:
// 1. Validates the request
// 2. Executes the HTTP request
// 3. Saves the execution to history (even if the HTTP request failed)
// 4. Returns the response
//
// If saving to history fails, it logs the error but doesn't fail the request.
func (s *RequestService) ExecuteAndSave(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		s.logger.Warn("request validation failed",
			"request_id", req.ID,
			"error", err,
		)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	s.logger.Info("executing and saving request",
		"request_id", req.ID,
		"method", req.Method,
		"url", req.URL,
	)

	// Execute HTTP request
	resp, err := s.httpClient.Execute(ctx, req)

	// Create history entry regardless of success or failure
	historyEntry := &repository.HistoryEntry{
		ID:             uuid.New().String(),
		RequestID:      req.ID,
		ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
		ResponseTimeMs: 0,
	}

	if err != nil {
		// Request failed - record the error
		historyEntry.Error = err.Error()
		s.logger.Error("request execution failed",
			"request_id", req.ID,
			"error", err,
		)
	} else {
		// Request succeeded - record the response
		historyEntry.StatusCode = resp.StatusCode
		historyEntry.Status = resp.Status
		historyEntry.ResponseTimeMs = resp.DurationMillis()
		historyEntry.ResponseBody = resp.Body

		// Convert headers map to JSON string using proper JSON marshaling
		headersBytes, err := json.Marshal(resp.Headers)
		if err != nil {
			// If marshaling fails, use empty JSON object
			s.logger.Error("failed to marshal response headers", "error", err)
			historyEntry.ResponseHeaders = "{}"
		} else {
			historyEntry.ResponseHeaders = string(headersBytes)
		}

		s.logger.Info("request executed successfully",
			"request_id", req.ID,
			"status_code", resp.StatusCode,
			"duration_ms", resp.DurationMillis(),
		)
	}

	// Save to history (best effort - don't fail the request if history save fails)
	if saveErr := s.historyRepo.Save(ctx, historyEntry); saveErr != nil {
		s.logger.Error("failed to save execution to history",
			"request_id", req.ID,
			"history_id", historyEntry.ID,
			"error", saveErr,
		)
		// Continue - we don't want to fail the request just because history save failed
	} else {
		s.logger.Debug("execution saved to history",
			"request_id", req.ID,
			"history_id", historyEntry.ID,
		)
	}

	// Return the original error if execution failed
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}
