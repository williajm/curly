package sqlite

// This file contains example code demonstrating how to use the SQLite repository layer.
// These examples are for documentation purposes only and are not intended to be executed directly.

/*
Example 1: Opening a database with migrations

	import (
		"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	)

	// Use default config (database at ~/.local/share/curly/curly.db)
	db, err := sqlite.Open(sqlite.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Or use custom config
	config := &sqlite.Config{
		Path:           "/custom/path/to/database.db",
		MigrationsPath: "migrations",
	}
	db, err = sqlite.Open(config)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

Example 2: Creating and using a request repository

	import (
		"context"
		"github.com/williajm/curly/internal/domain"
		"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	)

	// Open database
	db, err := sqlite.Open(sqlite.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create repository
	repo := sqlite.NewRequestRepository(db)

	// Create a new request
	req := domain.NewRequest()
	req.Name = "Get Users"
	req.Method = "GET"
	req.URL = "https://api.example.com/users"
	req.SetHeader("Accept", "application/json")
	req.SetAuth(domain.NewBearerAuth("my-token"))

	// Save to database
	ctx := context.Background()
	if err := repo.Create(ctx, req); err != nil {
		log.Fatalf("failed to create request: %v", err)
	}

	// Retrieve by ID
	retrieved, err := repo.FindByID(ctx, req.ID)
	if err != nil {
		log.Fatalf("failed to find request: %v", err)
	}

	// Update request
	retrieved.Name = "Get All Users"
	retrieved.SetHeader("Content-Type", "application/json")
	if err := repo.Update(ctx, retrieved); err != nil {
		log.Fatalf("failed to update request: %v", err)
	}

	// Get all requests
	allRequests, err := repo.FindAll(ctx)
	if err != nil {
		log.Fatalf("failed to find all requests: %v", err)
	}
	fmt.Printf("Found %d requests\n", len(allRequests))

	// Delete request
	if err := repo.Delete(ctx, req.ID); err != nil {
		log.Fatalf("failed to delete request: %v", err)
	}

Example 3: Using the history repository

	import (
		"context"
		"time"
		"github.com/google/uuid"
		"github.com/williajm/curly/internal/infrastructure/repository"
		"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	)

	// Open database
	db, err := sqlite.Open(sqlite.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create history repository
	historyRepo := sqlite.NewHistoryRepository(db)

	// Save a successful execution
	entry := &repository.HistoryEntry{
		ID:              uuid.New().String(),
		RequestID:       "some-request-id",
		ExecutedAt:      time.Now().Format(time.RFC3339),
		StatusCode:      200,
		Status:          "200 OK",
		ResponseTimeMs:  123,
		ResponseHeaders: `{"Content-Type": "application/json"}`,
		ResponseBody:    `{"users": []}`,
		Error:           "",
	}

	ctx := context.Background()
	if err := historyRepo.Save(ctx, entry); err != nil {
		log.Fatalf("failed to save history entry: %v", err)
	}

	// Save a failed execution
	failedEntry := &repository.HistoryEntry{
		ID:              uuid.New().String(),
		RequestID:       "some-request-id",
		ExecutedAt:      time.Now().Format(time.RFC3339),
		StatusCode:      0,
		Status:          "",
		ResponseTimeMs:  0,
		ResponseHeaders: "",
		ResponseBody:    "",
		Error:           "connection timeout",
	}

	if err := historyRepo.Save(ctx, failedEntry); err != nil {
		log.Fatalf("failed to save failed entry: %v", err)
	}

	// Get history for a specific request (latest 10 entries)
	requestHistory, err := historyRepo.FindByRequestID(ctx, "some-request-id", 10)
	if err != nil {
		log.Fatalf("failed to find history: %v", err)
	}
	fmt.Printf("Found %d history entries\n", len(requestHistory))

	// Get all history (latest 100 entries)
	allHistory, err := historyRepo.FindAll(ctx, 100)
	if err != nil {
		log.Fatalf("failed to find all history: %v", err)
	}
	fmt.Printf("Found %d total history entries\n", len(allHistory))

	// Clean up old history (older than 90 days)
	cutoff := time.Now().AddDate(0, 0, -90).Format(time.RFC3339)
	deleted, err := historyRepo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		log.Fatalf("failed to delete old history: %v", err)
	}
	fmt.Printf("Deleted %d old history entries\n", deleted)

Example 4: Using repositories together with HTTP client

	import (
		"context"
		"encoding/json"
		"time"
		"github.com/google/uuid"
		"github.com/williajm/curly/internal/domain"
		"github.com/williajm/curly/internal/infrastructure/http"
		"github.com/williajm/curly/internal/infrastructure/repository"
		"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	)

	// Open database
	db, err := sqlite.Open(sqlite.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create repositories
	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)

	// Create HTTP client
	httpClient := http.NewClient(http.DefaultConfig())

	ctx := context.Background()

	// Create and save a request
	req := domain.NewRequest()
	req.Name = "Get GitHub User"
	req.Method = "GET"
	req.URL = "https://api.github.com/users/octocat"
	req.SetHeader("Accept", "application/vnd.github.v3+json")

	if err := requestRepo.Create(ctx, req); err != nil {
		log.Fatalf("failed to save request: %v", err)
	}

	// Execute the request
	resp, err := httpClient.Execute(ctx, req)

	// Save execution to history
	entry := &repository.HistoryEntry{
		ID:             uuid.New().String(),
		RequestID:      req.ID,
		ExecutedAt:     time.Now().Format(time.RFC3339),
		ResponseTimeMs: resp.DurationMillis(),
	}

	if err != nil {
		// Request failed
		entry.Error = err.Error()
	} else {
		// Request succeeded
		entry.StatusCode = resp.StatusCode
		entry.Status = resp.Status
		headersJSON, _ := json.Marshal(resp.Headers)
		entry.ResponseHeaders = string(headersJSON)
		entry.ResponseBody = resp.Body
	}

	if err := historyRepo.Save(ctx, entry); err != nil {
		log.Fatalf("failed to save history: %v", err)
	}

	// View execution history for this request
	history, err := historyRepo.FindByRequestID(ctx, req.ID, 0)
	if err != nil {
		log.Fatalf("failed to get history: %v", err)
	}

	fmt.Printf("Request '%s' has been executed %d times\n", req.Name, len(history))
	for _, h := range history {
		if h.Error != "" {
			fmt.Printf("  - %s: FAILED - %s\n", h.ExecutedAt, h.Error)
		} else {
			fmt.Printf("  - %s: %s (%dms)\n", h.ExecutedAt, h.Status, h.ResponseTimeMs)
		}
	}

Example 5: Using in-memory database for testing

	import (
		"context"
		"testing"
		"github.com/williajm/curly/internal/domain"
		"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	)

	func TestMyFeature(t *testing.T) {
		// Use in-memory database for testing
		config := &sqlite.Config{
			Path:           ":memory:",
			MigrationsPath: "../../../../migrations",
		}

		db, err := sqlite.Open(config)
		if err != nil {
			t.Fatalf("failed to open test database: %v", err)
		}
		defer db.Close()

		// Create repository
		repo := sqlite.NewRequestRepository(db)

		// Test CRUD operations
		ctx := context.Background()
		req := domain.NewRequest()
		req.Name = "Test Request"
		req.Method = "GET"
		req.URL = "https://api.example.com/test"

		// Create
		if err := repo.Create(ctx, req); err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		// Read
		retrieved, err := repo.FindByID(ctx, req.ID)
		if err != nil {
			t.Fatalf("failed to find request: %v", err)
		}

		if retrieved.Name != req.Name {
			t.Errorf("expected name %s, got %s", req.Name, retrieved.Name)
		}

		// Update
		retrieved.Name = "Updated Test Request"
		if err := repo.Update(ctx, retrieved); err != nil {
			t.Fatalf("failed to update request: %v", err)
		}

		// Verify update
		updated, err := repo.FindByID(ctx, req.ID)
		if err != nil {
			t.Fatalf("failed to find updated request: %v", err)
		}

		if updated.Name != "Updated Test Request" {
			t.Errorf("expected updated name, got %s", updated.Name)
		}

		// Delete
		if err := repo.Delete(ctx, req.ID); err != nil {
			t.Fatalf("failed to delete request: %v", err)
		}

		// Verify deletion
		_, err = repo.FindByID(ctx, req.ID)
		if err != sqlite.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	}
*/
