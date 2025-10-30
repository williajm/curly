package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/williajm/curly/internal/domain"
	"github.com/williajm/curly/internal/infrastructure/repository"
)

// createTestRequest is a helper to create a test request in the database.
//
//nolint:revive // context-as-argument: testing.T should be first parameter in test helpers.
func createTestRequest(t *testing.T, ctx context.Context, repo *RequestRepository, id string) *domain.Request {
	t.Helper()
	req := &domain.Request{
		ID:          id,
		Name:        "Test Request",
		Method:      "GET",
		URL:         "https://api.example.com/test",
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
		Body:        "",
		AuthConfig:  domain.NewNoAuth(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, req); err != nil {
		t.Fatalf("failed to create test request: %v", err)
	}
	return req
}

func TestHistoryRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	ctx := context.Background()

	// Create a test request for foreign key reference.
	reqRepo := NewRequestRepository(db)
	testReq := createTestRequest(t, ctx, reqRepo, "req-123")

	tests := []struct {
		name    string
		entry   *repository.HistoryEntry
		wantErr bool
	}{
		{
			name: "valid entry with request ID",
			entry: &repository.HistoryEntry{
				ID:              uuid.New().String(),
				RequestID:       testReq.ID,
				ExecutedAt:      time.Now().Format(time.RFC3339),
				StatusCode:      200,
				Status:          "200 OK",
				ResponseTimeMs:  150,
				ResponseHeaders: `{"Content-Type": "application/json"}`,
				ResponseBody:    `{"success": true}`,
				Error:           "",
			},
			wantErr: false,
		},
		{
			name: "valid entry without request ID (ad-hoc request)",
			entry: &repository.HistoryEntry{
				ID:              uuid.New().String(),
				RequestID:       "",
				ExecutedAt:      time.Now().Format(time.RFC3339),
				StatusCode:      404,
				Status:          "404 Not Found",
				ResponseTimeMs:  100,
				ResponseHeaders: `{"Content-Type": "text/html"}`,
				ResponseBody:    "Not found",
				Error:           "",
			},
			wantErr: false,
		},
		{
			name: "entry with error",
			entry: &repository.HistoryEntry{
				ID:              uuid.New().String(),
				RequestID:       testReq.ID,
				ExecutedAt:      time.Now().Format(time.RFC3339),
				StatusCode:      0,
				Status:          "",
				ResponseTimeMs:  0,
				ResponseHeaders: "",
				ResponseBody:    "",
				Error:           "connection timeout",
			},
			wantErr: false,
		},
		{
			name:    "nil entry",
			entry:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(ctx, tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHistoryRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	reqRepo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test request for foreign key reference.
	testReq := createTestRequest(t, ctx, reqRepo, "req-123")

	// Create test entry.
	testEntry := &repository.HistoryEntry{
		ID:              "hist-123",
		RequestID:       testReq.ID,
		ExecutedAt:      time.Now().Format(time.RFC3339),
		StatusCode:      200,
		Status:          "200 OK",
		ResponseTimeMs:  123,
		ResponseHeaders: `{"Content-Type": "application/json"}`,
		ResponseBody:    `{"test": "data"}`,
		Error:           "",
	}

	if err := repo.Save(ctx, testEntry); err != nil {
		t.Fatalf("failed to save test entry: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "existing entry",
			id:      "hist-123",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "non-existing entry",
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
				if got.ID != testEntry.ID {
					t.Errorf("FindByID() ID = %v, want %v", got.ID, testEntry.ID)
				}
				if got.RequestID != testEntry.RequestID {
					t.Errorf("FindByID() RequestID = %v, want %v", got.RequestID, testEntry.RequestID)
				}
				if got.StatusCode != testEntry.StatusCode {
					t.Errorf("FindByID() StatusCode = %v, want %v", got.StatusCode, testEntry.StatusCode)
				}
				if got.Status != testEntry.Status {
					t.Errorf("FindByID() Status = %v, want %v", got.Status, testEntry.Status)
				}
			}
		})
	}
}

func TestHistoryRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	reqRepo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test requests for foreign key references.
	req1 := createTestRequest(t, ctx, reqRepo, "req-1")
	req2 := createTestRequest(t, ctx, reqRepo, "req-2")
	req3 := createTestRequest(t, ctx, reqRepo, "req-3")

	// Create test entries with different timestamps.
	baseTime := time.Now()
	entries := []*repository.HistoryEntry{
		{
			ID:              "hist-1",
			RequestID:       req1.ID,
			ExecutedAt:      baseTime.Format(time.RFC3339),
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  100,
			ResponseHeaders: `{}`,
			ResponseBody:    "body1",
			Error:           "",
		},
		{
			ID:              "hist-2",
			RequestID:       req2.ID,
			ExecutedAt:      baseTime.Add(time.Second).Format(time.RFC3339),
			StatusCode:      201,
			Status:          "201 Created",
			ResponseTimeMs:  150,
			ResponseHeaders: `{}`,
			ResponseBody:    "body2",
			Error:           "",
		},
		{
			ID:              "hist-3",
			RequestID:       req3.ID,
			ExecutedAt:      baseTime.Add(2 * time.Second).Format(time.RFC3339),
			StatusCode:      404,
			Status:          "404 Not Found",
			ResponseTimeMs:  50,
			ResponseHeaders: `{}`,
			ResponseBody:    "body3",
			Error:           "",
		},
	}

	for _, entry := range entries {
		if err := repo.Save(ctx, entry); err != nil {
			t.Fatalf("failed to save test entry: %v", err)
		}
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "no limit",
			limit:     0,
			wantCount: 3,
		},
		{
			name:      "limit 2",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "limit greater than total",
			limit:     10,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindAll(ctx, tt.limit)
			if err != nil {
				t.Fatalf("FindAll() error = %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("FindAll() returned %d entries, want %d", len(got), tt.wantCount)
			}

			// Verify ordering (newest first).
			if len(got) >= 2 {
				if got[0].ID != "hist-3" {
					t.Errorf("FindAll() first entry ID = %v, want hist-3 (newest)", got[0].ID)
				}
			}
		})
	}
}

func TestHistoryRepository_FindByRequestID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	reqRepo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test requests for foreign key references.
	reqA := createTestRequest(t, ctx, reqRepo, "req-a")
	reqB := createTestRequest(t, ctx, reqRepo, "req-b")

	// Create test entries for the same request.
	baseTime := time.Now()
	entries := []*repository.HistoryEntry{
		{
			ID:              "hist-a1",
			RequestID:       reqA.ID,
			ExecutedAt:      baseTime.Format(time.RFC3339),
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  100,
			ResponseHeaders: `{}`,
			ResponseBody:    "execution1",
			Error:           "",
		},
		{
			ID:              "hist-a2",
			RequestID:       reqA.ID,
			ExecutedAt:      baseTime.Add(time.Second).Format(time.RFC3339),
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  110,
			ResponseHeaders: `{}`,
			ResponseBody:    "execution2",
			Error:           "",
		},
		{
			ID:              "hist-b1",
			RequestID:       reqB.ID,
			ExecutedAt:      baseTime.Format(time.RFC3339),
			StatusCode:      404,
			Status:          "404 Not Found",
			ResponseTimeMs:  50,
			ResponseHeaders: `{}`,
			ResponseBody:    "not found",
			Error:           "",
		},
	}

	for _, entry := range entries {
		if err := repo.Save(ctx, entry); err != nil {
			t.Fatalf("failed to save test entry: %v", err)
		}
	}

	tests := []struct {
		name      string
		requestID string
		limit     int
		wantCount int
		wantFirst string
	}{
		{
			name:      "request with multiple executions",
			requestID: reqA.ID,
			limit:     0,
			wantCount: 2,
			wantFirst: "hist-a2", // newest first
		},
		{
			name:      "request with single execution",
			requestID: reqB.ID,
			limit:     0,
			wantCount: 1,
			wantFirst: "hist-b1",
		},
		{
			name:      "non-existent request",
			requestID: "req-nonexistent",
			limit:     0,
			wantCount: 0,
			wantFirst: "",
		},
		{
			name:      "with limit",
			requestID: reqA.ID,
			limit:     1,
			wantCount: 1,
			wantFirst: "hist-a2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByRequestID(ctx, tt.requestID, tt.limit)
			if err != nil {
				t.Fatalf("FindByRequestID() error = %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("FindByRequestID() returned %d entries, want %d", len(got), tt.wantCount)
			}

			if tt.wantFirst != "" && len(got) > 0 {
				if got[0].ID != tt.wantFirst {
					t.Errorf("FindByRequestID() first entry ID = %v, want %v", got[0].ID, tt.wantFirst)
				}
			}
		})
	}
}

func TestHistoryRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	reqRepo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test request for foreign key reference.
	testReq := createTestRequest(t, ctx, reqRepo, "req-delete")

	// Create test entry.
	testEntry := &repository.HistoryEntry{
		ID:              "hist-delete-1",
		RequestID:       testReq.ID,
		ExecutedAt:      time.Now().Format(time.RFC3339),
		StatusCode:      200,
		Status:          "200 OK",
		ResponseTimeMs:  100,
		ResponseHeaders: `{}`,
		ResponseBody:    "test",
		Error:           "",
	}

	if err := repo.Save(ctx, testEntry); err != nil {
		t.Fatalf("failed to save test entry: %v", err)
	}

	// Delete the entry.
	if err := repo.Delete(ctx, "hist-delete-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion.
	_, err := repo.FindByID(ctx, "hist-delete-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID() after delete error = %v, want ErrNotFound", err)
	}

	// Test deleting non-existent entry.
	err = repo.Delete(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() with non-existent ID error = %v, want ErrNotFound", err)
	}
}

func TestHistoryRepository_DeleteOlderThan(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	reqRepo := NewRequestRepository(db)
	ctx := context.Background()

	// Create test requests for foreign key references.
	req1 := createTestRequest(t, ctx, reqRepo, "req-old-1")
	req2 := createTestRequest(t, ctx, reqRepo, "req-old-2")
	req3 := createTestRequest(t, ctx, reqRepo, "req-old-3")

	// Create entries with different timestamps.
	now := time.Now()
	entries := []*repository.HistoryEntry{
		{
			ID:              "hist-old-1",
			RequestID:       req1.ID,
			ExecutedAt:      now.Add(-48 * time.Hour).Format(time.RFC3339), // 2 days ago
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  100,
			ResponseHeaders: `{}`,
			ResponseBody:    "old",
			Error:           "",
		},
		{
			ID:              "hist-old-2",
			RequestID:       req2.ID,
			ExecutedAt:      now.Add(-25 * time.Hour).Format(time.RFC3339), // 25 hours ago
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  100,
			ResponseHeaders: `{}`,
			ResponseBody:    "old",
			Error:           "",
		},
		{
			ID:              "hist-recent",
			RequestID:       req3.ID,
			ExecutedAt:      now.Add(-1 * time.Hour).Format(time.RFC3339), // 1 hour ago
			StatusCode:      200,
			Status:          "200 OK",
			ResponseTimeMs:  100,
			ResponseHeaders: `{}`,
			ResponseBody:    "recent",
			Error:           "",
		},
	}

	for _, entry := range entries {
		if err := repo.Save(ctx, entry); err != nil {
			t.Fatalf("failed to save test entry: %v", err)
		}
	}

	// Delete entries older than 24 hours.
	cutoff := now.Add(-24 * time.Hour).Format(time.RFC3339)
	deleted, err := repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}

	if deleted != 2 {
		t.Errorf("DeleteOlderThan() deleted %d entries, want 2", deleted)
	}

	// Verify the recent entry still exists.
	_, err = repo.FindByID(ctx, "hist-recent")
	if err != nil {
		t.Errorf("FindByID(hist-recent) error = %v, want nil (entry should exist)", err)
	}

	// Verify old entries are deleted.
	_, err = repo.FindByID(ctx, "hist-old-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID(hist-old-1) error = %v, want ErrNotFound", err)
	}

	_, err = repo.FindByID(ctx, "hist-old-2")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID(hist-old-2) error = %v, want ErrNotFound", err)
	}
}

func TestHistoryRepository_NullHandling(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewHistoryRepository(db)
	ctx := context.Background()

	// Create entry with NULL request_id and error.
	entryWithNulls := &repository.HistoryEntry{
		ID:              "hist-nulls",
		RequestID:       "", // Will be NULL in database
		ExecutedAt:      time.Now().Format(time.RFC3339),
		StatusCode:      200,
		Status:          "200 OK",
		ResponseTimeMs:  100,
		ResponseHeaders: `{}`,
		ResponseBody:    "test",
		Error:           "", // Will be NULL in database
	}

	if err := repo.Save(ctx, entryWithNulls); err != nil {
		t.Fatalf("Save() with nulls error = %v", err)
	}

	// Retrieve and verify NULL fields are handled correctly.
	got, err := repo.FindByID(ctx, "hist-nulls")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if got.RequestID != "" {
		t.Errorf("FindByID() RequestID = %q, want empty string", got.RequestID)
	}

	if got.Error != "" {
		t.Errorf("FindByID() Error = %q, want empty string", got.Error)
	}
}
