// Package app provides example usage patterns for the application services.
// This file demonstrates how to initialize and use the services together.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/williajm/curly/internal/domain"
	"github.com/williajm/curly/internal/infrastructure/http"
	"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
)

// ExampleRequestServiceUsage demonstrates the basic usage of RequestService.
func ExampleRequestServiceUsage() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize database
	db, err := sqlite.Open(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return
	}
	defer db.Close()

	// Create repositories
	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)

	// Create HTTP client
	httpClient := http.NewClient(http.DefaultConfig())

	// Create service
	service := NewRequestService(requestRepo, httpClient, historyRepo, logger)

	// Create a new request
	req := domain.NewRequestWithMethodAndURL("GET", "https://api.github.com/users/octocat")
	req.Name = "Get GitHub User"
	req.SetHeader("Accept", "application/json")

	// Add authentication (optional)
	auth := domain.NewBearerAuth("your-github-token")
	req.SetAuth(auth)

	// Create and validate the request
	createdReq, err := service.CreateRequest(context.Background(), req)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		return
	}

	// Save the request to database
	if err := service.SaveRequest(context.Background(), createdReq); err != nil {
		logger.Error("failed to save request", "error", err)
		return
	}

	// Execute the request and save to history
	resp, err := service.ExecuteAndSave(context.Background(), createdReq)
	if err != nil {
		logger.Error("failed to execute request", "error", err)
		return
	}

	logger.Info("request executed successfully",
		"status_code", resp.StatusCode,
		"duration_ms", resp.DurationMillis(),
		"body_length", len(resp.Body),
	)

	// List all saved requests
	requests, err := service.ListRequests(context.Background())
	if err != nil {
		logger.Error("failed to list requests", "error", err)
		return
	}

	logger.Info("saved requests", "count", len(requests))
}

// ExampleHistoryServiceUsage demonstrates the basic usage of HistoryService.
func ExampleHistoryServiceUsage() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize database
	db, err := sqlite.Open(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return
	}
	defer db.Close()

	// Create history repository
	historyRepo := sqlite.NewHistoryRepository(db)

	// Create service
	service := NewHistoryService(historyRepo, logger)

	// Get all history entries (limit to 10)
	entries, err := service.GetHistory(context.Background(), 10)
	if err != nil {
		logger.Error("failed to get history", "error", err)
		return
	}

	logger.Info("history entries", "count", len(entries))

	// Get history for a specific request
	requestID := "some-request-id"
	requestHistory, err := service.GetRequestHistory(context.Background(), requestID, 5)
	if err != nil {
		logger.Error("failed to get request history", "error", err)
		return
	}

	logger.Info("request history", "request_id", requestID, "count", len(requestHistory))

	// Cleanup old history (keep last 30 days)
	deletedCount, err := service.CleanupOldHistory(context.Background(), 30)
	if err != nil {
		logger.Error("failed to cleanup history", "error", err)
		return
	}

	logger.Info("cleaned up old history", "deleted_count", deletedCount)
}

// ExampleAuthServiceUsage demonstrates the basic usage of AuthService.
func ExampleAuthServiceUsage() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create service
	service := NewAuthService(logger)

	// Example 1: Create Basic Auth
	basicAuth, err := service.CreateAuth("basic", map[string]string{
		"username": "myuser",
		"password": "mypassword",
	})
	if err != nil {
		logger.Error("failed to create basic auth", "error", err)
		return
	}

	logger.Info("created basic auth", "type", basicAuth.Type())

	// Example 2: Create Bearer Token Auth
	bearerAuth, err := service.CreateAuth("bearer", map[string]string{
		"token": "my-secret-token",
	})
	if err != nil {
		logger.Error("failed to create bearer auth", "error", err)
		return
	}

	logger.Info("created bearer auth", "type", bearerAuth.Type())

	// Example 3: Create API Key Auth (Header)
	apiKeyAuthHeader, err := service.CreateAuth("apikey", map[string]string{
		"key":      "X-API-Key",
		"value":    "my-api-key",
		"location": "header",
	})
	if err != nil {
		logger.Error("failed to create api key auth", "error", err)
		return
	}

	logger.Info("created api key auth", "type", apiKeyAuthHeader.Type())

	// Example 4: Create API Key Auth (Query)
	apiKeyAuthQuery, err := service.CreateAuth("apikey", map[string]string{
		"key":      "api_key",
		"value":    "my-api-key",
		"location": "query",
	})
	if err != nil {
		logger.Error("failed to create api key auth", "error", err)
		return
	}

	logger.Info("created api key auth", "type", apiKeyAuthQuery.Type())

	// Validate auth config
	if err := service.ValidateAuth(basicAuth); err != nil {
		logger.Error("auth validation failed", "error", err)
		return
	}

	logger.Info("auth validated successfully")

	// List supported auth types
	supportedTypes := service.SupportedTypes()
	logger.Info("supported auth types", "types", supportedTypes)
}

// ExampleFullWorkflow demonstrates a complete workflow using all services together.
func ExampleFullWorkflow() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize database
	db, err := sqlite.Open(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return
	}
	defer db.Close()

	// Create repositories
	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)

	// Create HTTP client
	httpClient := http.NewClient(&http.Config{
		Timeout:         30 * time.Second,
		MaxRedirects:    10,
		FollowRedirects: true,
	})

	// Create services
	requestService := NewRequestService(requestRepo, httpClient, historyRepo, logger)
	historyService := NewHistoryService(historyRepo, logger)
	authService := NewAuthService(logger)

	// Step 1: Create authentication
	auth, err := authService.CreateAuth("bearer", map[string]string{
		"token": "github-token",
	})
	if err != nil {
		logger.Error("failed to create auth", "error", err)
		return
	}

	// Step 2: Create a request
	req := domain.NewRequestWithMethodAndURL("GET", "https://api.github.com/user")
	req.Name = "Get Current User"
	req.SetHeader("Accept", "application/vnd.github.v3+json")
	req.SetAuth(auth)

	// Step 3: Save the request
	if err := requestService.SaveRequest(context.Background(), req); err != nil {
		logger.Error("failed to save request", "error", err)
		return
	}

	logger.Info("request saved", "request_id", req.ID)

	// Step 4: Execute the request and save to history
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := requestService.ExecuteAndSave(ctx, req)
	if err != nil {
		logger.Error("failed to execute request", "error", err)
		return
	}

	logger.Info("request executed successfully",
		"status_code", resp.StatusCode,
		"duration_ms", resp.DurationMillis(),
	)

	// Step 5: Retrieve history for this request
	history, err := historyService.GetRequestHistory(context.Background(), req.ID, 0)
	if err != nil {
		logger.Error("failed to get request history", "error", err)
		return
	}

	logger.Info("request history retrieved", "count", len(history))

	// Step 6: List all saved requests
	requests, err := requestService.ListRequests(context.Background())
	if err != nil {
		logger.Error("failed to list requests", "error", err)
		return
	}

	logger.Info("all saved requests", "count", len(requests))

	// Step 7: Cleanup old history (keep last 90 days)
	deletedCount, err := historyService.CleanupOldHistory(context.Background(), 90)
	if err != nil {
		logger.Error("failed to cleanup history", "error", err)
		return
	}

	logger.Info("cleaned up old history", "deleted_count", deletedCount)

	fmt.Println("Workflow completed successfully!")
}
