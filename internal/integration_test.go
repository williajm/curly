package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/williajm/curly/internal/app"
	"github.com/williajm/curly/internal/domain"
	httpinfra "github.com/williajm/curly/internal/infrastructure/http"
	"github.com/williajm/curly/internal/infrastructure/repository"
	"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
)

// TestEndToEnd_RequestExecution tests the complete flow of creating, executing, and saving a request
func TestEndToEnd_RequestExecution(t *testing.T) {
	// Set up test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "success",
			"method":  r.Method,
		})
	}))
	defer server.Close()

	// Set up test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	err = sqlite.MigrateDB(db)
	require.NoError(t, err)

	// Initialize repositories
	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)

	// Initialize HTTP client
	httpClient := httpinfra.NewClient(&httpinfra.Config{
		Timeout:         10 * time.Second,
		MaxRedirects:    10,
		FollowRedirects: true,
		InsecureSkipTLS: false,
	})

	// Initialize services
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	requestService := app.NewRequestService(requestRepo, httpClient, historyRepo, logger)
	historyService := app.NewHistoryService(historyRepo, logger)

	// Create a test request using the service (which sets timestamps and ID)
	ctx := context.Background()
	req, err := requestService.CreateRequest(ctx, &domain.Request{
		Name:   "Test Request",
		Method: "GET",
		URL:    server.URL + "/test",
		Headers: map[string]string{
			"Accept": "application/json",
		},
		QueryParams: map[string]string{
			"foo": "bar",
		},
	})
	require.NoError(t, err)

	// Save the request first (required for foreign key constraint)
	err = requestService.SaveRequest(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, req.ID)

	// Execute the request and save to history
	resp, err2 := requestService.ExecuteAndSave(ctx, req)
	require.NoError(t, err2)
	require.NotNil(t, resp)

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
	assert.NotZero(t, resp.Duration)

	// Load the request back
	loaded, err := requestService.LoadRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, req.Name, loaded.Name)
	assert.Equal(t, req.Method, loaded.Method)
	assert.Equal(t, req.URL, loaded.URL)

	// Verify history was saved
	history, err := historyService.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, req.ID, history[0].RequestID)
	assert.Equal(t, http.StatusOK, history[0].StatusCode)
}

// TestEndToEnd_AuthenticationFlow tests request execution with different auth methods
func TestEndToEnd_AuthenticationFlow(t *testing.T) {
	tests := []struct {
		name     string
		authType string
		setup    func(req *domain.Request)
		verify   func(t *testing.T, r *http.Request)
	}{
		{
			name:     "Basic Auth",
			authType: "basic",
			setup: func(req *domain.Request) {
				req.AuthConfig = &domain.BasicAuth{
					Username: "testuser",
					Password: "testpass",
				}
			},
			verify: func(t *testing.T, r *http.Request) {
				username, password, ok := r.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, "testuser", username)
				assert.Equal(t, "testpass", password)
			},
		},
		{
			name:     "Bearer Token",
			authType: "bearer",
			setup: func(req *domain.Request) {
				req.AuthConfig = &domain.BearerAuth{
					Token: "test-token-123",
				}
			},
			verify: func(t *testing.T, r *http.Request) {
				auth := r.Header.Get("Authorization")
				assert.Equal(t, "Bearer test-token-123", auth)
			},
		},
		{
			name:     "API Key Header",
			authType: "apikey",
			setup: func(req *domain.Request) {
				req.AuthConfig = &domain.APIKeyAuth{
					Key:      "X-API-Key",
					Value:    "secret-key",
					Location: domain.APIKeyLocationHeader,
				}
			},
			verify: func(t *testing.T, r *http.Request) {
				apiKey := r.Header.Get("X-API-Key")
				assert.Equal(t, "secret-key", apiKey)
			},
		},
		{
			name:     "API Key Query",
			authType: "apikey-query",
			setup: func(req *domain.Request) {
				req.AuthConfig = &domain.APIKeyAuth{
					Key:      "api_key",
					Value:    "secret-key",
					Location: domain.APIKeyLocationQuery,
				}
			},
			verify: func(t *testing.T, r *http.Request) {
				apiKey := r.URL.Query().Get("api_key")
				assert.Equal(t, "secret-key", apiKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.verify(t, r)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Set up test database
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
			require.NoError(t, err)
			defer db.Close()

			err = sqlite.MigrateDB(db)
			require.NoError(t, err)

			// Initialize services
			requestRepo := sqlite.NewRequestRepository(db)
			historyRepo := sqlite.NewHistoryRepository(db)
			httpClient := httpinfra.NewClient(httpinfra.DefaultConfig())
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			requestService := app.NewRequestService(requestRepo, httpClient, historyRepo, logger)

			// Create request
			req := &domain.Request{
				ID:     uuid.New().String(),
				Method: "GET",
				URL:    server.URL,
			}
			tt.setup(req)

			// Execute request (auth is applied automatically by the HTTP client)
			resp, err := requestService.ExecuteRequest(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// TestEndToEnd_PersistenceFlow tests database persistence across app restarts
func TestEndToEnd_PersistenceFlow(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First session: Create and save request
	req1 := &domain.Request{
		ID:     uuid.New().String(),
		Name:   "Persisted Request",
		Method: "POST",
		URL:    "https://api.example.com/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"test": "data"}`,
	}

	func() {
		db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
		require.NoError(t, err)
		defer db.Close()

		err = sqlite.MigrateDB(db)
		require.NoError(t, err)

		requestRepo := sqlite.NewRequestRepository(db)
		err = requestRepo.Create(context.Background(), req1)
		require.NoError(t, err)
	}()

	// Second session: Load request from database
	func() {
		db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
		require.NoError(t, err)
		defer db.Close()

		requestRepo := sqlite.NewRequestRepository(db)
		loaded, err := requestRepo.FindByID(context.Background(), req1.ID)
		require.NoError(t, err)

		assert.Equal(t, req1.Name, loaded.Name)
		assert.Equal(t, req1.Method, loaded.Method)
		assert.Equal(t, req1.URL, loaded.URL)
		assert.Equal(t, req1.Body, loaded.Body)
		assert.Equal(t, req1.Headers, loaded.Headers)
	}()
}

// TestEndToEnd_HistoryManagement tests history save, retrieve, and cleanup
func TestEndToEnd_HistoryManagement(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	err = sqlite.MigrateDB(db)
	require.NoError(t, err)

	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)
	httpClient := httpinfra.NewClient(httpinfra.DefaultConfig())
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	requestService := app.NewRequestService(requestRepo, httpClient, historyRepo, logger)
	historyService := app.NewHistoryService(historyRepo, logger)
	ctx := context.Background()

	// Create requests first (required for foreign key)
	requestIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		req, err := requestService.CreateRequest(ctx, &domain.Request{
			Name:   "Test Request " + string(rune(i)),
			Method: "GET",
			URL:    "https://api.example.com/test",
		})
		require.NoError(t, err)
		err = requestService.SaveRequest(ctx, req)
		require.NoError(t, err)
		requestIDs[i] = req.ID
	}

	// Create multiple history entries
	for i := 0; i < 5; i++ {
		entry := &repository.HistoryEntry{
			ID:             uuid.New().String(),
			RequestID:      requestIDs[i],
			StatusCode:     200 + i,
			ResponseTimeMs: int64(i + 1),
			ResponseBody:   "response body " + uuid.New().String(),
			ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		err := historyService.SaveExecution(ctx, entry)
		require.NoError(t, err)
	}

	// List history entries
	history, err := historyService.GetHistory(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, history, 5)

	// Test limited results
	limited, err := historyService.GetHistory(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, limited, 2)
}

// TestEndToEnd_HTTPMethods tests all supported HTTP methods
func TestEndToEnd_HTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, method, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			db, err := sqlite.Open(&sqlite.Config{Path: dbPath})
			require.NoError(t, err)
			defer db.Close()

			err = sqlite.MigrateDB(db)
			require.NoError(t, err)

			requestRepo := sqlite.NewRequestRepository(db)
			historyRepo := sqlite.NewHistoryRepository(db)
			httpClient := httpinfra.NewClient(httpinfra.DefaultConfig())
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			requestService := app.NewRequestService(requestRepo, httpClient, historyRepo, logger)

			req := &domain.Request{
				ID:     uuid.New().String(),
				Method: method,
				URL:    server.URL,
			}

			if method == "POST" || method == "PUT" || method == "PATCH" {
				req.Body = `{"test": "data"}`
				req.Headers = map[string]string{"Content-Type": "application/json"}
			}

			resp, err := requestService.ExecuteRequest(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
