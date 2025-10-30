package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/williajm/curly/internal/domain"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Apply pragmas.
	if err := applyPragmas(db); err != nil {
		t.Fatalf("failed to apply pragmas: %v", err)
	}

	// Run migrations.
	if err := runMigrations(db, "../../../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestRequestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		request *domain.Request
		wantErr bool
	}{
		{
			name: "valid request with no auth",
			request: &domain.Request{
				ID:          "test-1",
				Name:        "Test Request",
				Method:      "GET",
				URL:         "https://api.example.com/test",
				Headers:     map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]string{"key": "value"},
				Body:        "",
				AuthConfig:  domain.NewNoAuth(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid request with basic auth",
			request: &domain.Request{
				ID:          "test-2",
				Name:        "Basic Auth Request",
				Method:      "POST",
				URL:         "https://api.example.com/auth",
				Headers:     map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]string{},
				Body:        `{"test": "data"}`,
				AuthConfig:  domain.NewBasicAuth("user", "pass"),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid request with bearer token",
			request: &domain.Request{
				ID:          "test-3",
				Name:        "Bearer Token Request",
				Method:      "GET",
				URL:         "https://api.example.com/secure",
				Headers:     map[string]string{},
				QueryParams: map[string]string{},
				Body:        "",
				AuthConfig:  domain.NewBearerAuth("test-token-123"),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid request with API key",
			request: &domain.Request{
				ID:          "test-4",
				Name:        "API Key Request",
				Method:      "GET",
				URL:         "https://api.example.com/data",
				Headers:     map[string]string{},
				QueryParams: map[string]string{},
				Body:        "",
				AuthConfig:  domain.NewAPIKeyAuth("X-API-Key", "secret-key", domain.APIKeyLocationHeader),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
		{
			name: "invalid URL",
			request: &domain.Request{
				ID:          "test-5",
				Name:        "Invalid Request",
				Method:      "GET",
				URL:         "not-a-valid-url",
				Headers:     map[string]string{},
				QueryParams: map[string]string{},
				Body:        "",
				AuthConfig:  domain.NewNoAuth(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test request.
	testReq := &domain.Request{
		ID:          "test-find-1",
		Name:        "Find Test",
		Method:      "GET",
		URL:         "https://api.example.com/test",
		Headers:     map[string]string{"Accept": "application/json"},
		QueryParams: map[string]string{"page": "1"},
		Body:        "",
		AuthConfig:  domain.NewBasicAuth("user", "pass"),
		CreatedAt:   time.Now().Truncate(time.Second),
		UpdatedAt:   time.Now().Truncate(time.Second),
	}

	if err := repo.Create(ctx, testReq); err != nil {
		t.Fatalf("failed to create test request: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "existing request",
			id:      "test-find-1",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "non-existing request",
			id:      "non-existent",
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("FindByID() got nil = %v, wantNil %v", got == nil, tt.wantNil)
			}

			if !tt.wantErr && got != nil {
				// Verify fields.
				if got.ID != testReq.ID {
					t.Errorf("FindByID() ID = %v, want %v", got.ID, testReq.ID)
				}
				if got.Name != testReq.Name {
					t.Errorf("FindByID() Name = %v, want %v", got.Name, testReq.Name)
				}
				if got.Method != testReq.Method {
					t.Errorf("FindByID() Method = %v, want %v", got.Method, testReq.Method)
				}
				if got.URL != testReq.URL {
					t.Errorf("FindByID() URL = %v, want %v", got.URL, testReq.URL)
				}
				if got.AuthConfig.Type() != testReq.AuthConfig.Type() {
					t.Errorf("FindByID() AuthConfig.Type() = %v, want %v", got.AuthConfig.Type(), testReq.AuthConfig.Type())
				}
			}
		})
	}
}

func TestRequestRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	// Create multiple test requests.
	requests := []*domain.Request{
		{
			ID:          "test-all-1",
			Name:        "Request 1",
			Method:      "GET",
			URL:         "https://api.example.com/1",
			Headers:     map[string]string{},
			QueryParams: map[string]string{},
			Body:        "",
			AuthConfig:  domain.NewNoAuth(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "test-all-2",
			Name:        "Request 2",
			Method:      "POST",
			URL:         "https://api.example.com/2",
			Headers:     map[string]string{},
			QueryParams: map[string]string{},
			Body:        "",
			AuthConfig:  domain.NewNoAuth(),
			CreatedAt:   time.Now().Add(time.Second),
			UpdatedAt:   time.Now().Add(time.Second),
		},
	}

	for _, req := range requests {
		if err := repo.Create(ctx, req); err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(got) != len(requests) {
		t.Errorf("FindAll() returned %d requests, want %d", len(got), len(requests))
	}

	// Verify results are ordered by created_at descending.
	if len(got) >= 2 {
		if got[0].ID != "test-all-2" {
			t.Errorf("FindAll() first request ID = %v, want %v (should be newest)", got[0].ID, "test-all-2")
		}
	}
}

func TestRequestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	// Create initial request.
	original := &domain.Request{
		ID:          "test-update-1",
		Name:        "Original Name",
		Method:      "GET",
		URL:         "https://api.example.com/original",
		Headers:     map[string]string{"Accept": "application/json"},
		QueryParams: map[string]string{},
		Body:        "",
		AuthConfig:  domain.NewNoAuth(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, original); err != nil {
		t.Fatalf("failed to create test request: %v", err)
	}

	// Update the request.
	updated := &domain.Request{
		ID:          "test-update-1",
		Name:        "Updated Name",
		Method:      "POST",
		URL:         "https://api.example.com/updated",
		Headers:     map[string]string{"Content-Type": "application/json"},
		QueryParams: map[string]string{"key": "value"},
		Body:        `{"updated": true}`,
		AuthConfig:  domain.NewBearerAuth("new-token"),
		CreatedAt:   original.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	if err := repo.Update(ctx, updated); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update.
	got, err := repo.FindByID(ctx, "test-update-1")
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}

	if got.Name != updated.Name {
		t.Errorf("Update() Name = %v, want %v", got.Name, updated.Name)
	}
	if got.Method != updated.Method {
		t.Errorf("Update() Method = %v, want %v", got.Method, updated.Method)
	}
	if got.URL != updated.URL {
		t.Errorf("Update() URL = %v, want %v", got.URL, updated.URL)
	}
	if got.Body != updated.Body {
		t.Errorf("Update() Body = %v, want %v", got.Body, updated.Body)
	}
	if got.AuthConfig.Type() != updated.AuthConfig.Type() {
		t.Errorf("Update() AuthConfig.Type() = %v, want %v", got.AuthConfig.Type(), updated.AuthConfig.Type())
	}

	// Test updating non-existent request.
	nonExistent := &domain.Request{
		ID:          "non-existent",
		Name:        "Non-existent",
		Method:      "GET",
		URL:         "https://api.example.com/test",
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
		Body:        "",
		AuthConfig:  domain.NewNoAuth(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Update(ctx, nonExistent)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Update() with non-existent ID error = %v, want ErrNotFound", err)
	}
}

func TestRequestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test request.
	testReq := &domain.Request{
		ID:          "test-delete-1",
		Name:        "Delete Test",
		Method:      "GET",
		URL:         "https://api.example.com/test",
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
		Body:        "",
		AuthConfig:  domain.NewNoAuth(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, testReq); err != nil {
		t.Fatalf("failed to create test request: %v", err)
	}

	// Delete the request.
	if err := repo.Delete(ctx, "test-delete-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion.
	_, err := repo.FindByID(ctx, "test-delete-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID() after delete error = %v, want ErrNotFound", err)
	}

	// Test deleting non-existent request.
	err = repo.Delete(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() with non-existent ID error = %v, want ErrNotFound", err)
	}
}

func TestRequestRepository_AuthConfigSerialization(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewRequestRepository(db)
	ctx := context.Background()

	tests := []struct {
		name       string
		authConfig domain.AuthConfig
	}{
		{
			name:       "NoAuth",
			authConfig: domain.NewNoAuth(),
		},
		{
			name:       "BasicAuth",
			authConfig: domain.NewBasicAuth("testuser", "testpass"),
		},
		{
			name:       "BearerAuth",
			authConfig: domain.NewBearerAuth("test-token-xyz"),
		},
		{
			name:       "APIKeyAuth - Header",
			authConfig: domain.NewAPIKeyAuth("X-API-Key", "secret123", domain.APIKeyLocationHeader),
		},
		{
			name:       "APIKeyAuth - Query",
			authConfig: domain.NewAPIKeyAuth("api_key", "secret456", domain.APIKeyLocationQuery),
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &domain.Request{
				ID:          fmt.Sprintf("test-auth-%d", i),
				Name:        tt.name,
				Method:      "GET",
				URL:         "https://api.example.com/test",
				Headers:     map[string]string{},
				QueryParams: map[string]string{},
				Body:        "",
				AuthConfig:  tt.authConfig,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Create.
			if err := repo.Create(ctx, req); err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			// Retrieve.
			got, err := repo.FindByID(ctx, req.ID)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			// Verify auth type.
			if got.AuthConfig.Type() != tt.authConfig.Type() {
				t.Errorf("AuthConfig.Type() = %v, want %v", got.AuthConfig.Type(), tt.authConfig.Type())
			}

			// Verify auth config details based on type.
			switch expected := tt.authConfig.(type) {
			case *domain.BasicAuth:
				gotAuth, ok := got.AuthConfig.(*domain.BasicAuth)
				if !ok {
					t.Fatalf("expected BasicAuth, got %T", got.AuthConfig)
				}
				if gotAuth.Username != expected.Username || gotAuth.Password != expected.Password {
					t.Errorf("BasicAuth = {%s, %s}, want {%s, %s}",
						gotAuth.Username, gotAuth.Password, expected.Username, expected.Password)
				}
			case *domain.BearerAuth:
				gotAuth, ok := got.AuthConfig.(*domain.BearerAuth)
				if !ok {
					t.Fatalf("expected BearerAuth, got %T", got.AuthConfig)
				}
				if gotAuth.Token != expected.Token {
					t.Errorf("BearerAuth.Token = %s, want %s", gotAuth.Token, expected.Token)
				}
			case *domain.APIKeyAuth:
				gotAuth, ok := got.AuthConfig.(*domain.APIKeyAuth)
				if !ok {
					t.Fatalf("expected APIKeyAuth, got %T", got.AuthConfig)
				}
				if gotAuth.Key != expected.Key || gotAuth.Value != expected.Value || gotAuth.Location != expected.Location {
					t.Errorf("APIKeyAuth = {%s, %s, %s}, want {%s, %s, %s}",
						gotAuth.Key, gotAuth.Value, gotAuth.Location, expected.Key, expected.Value, expected.Location)
				}
			}
		})
	}
}
