package app

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/williajm/curly/internal/infrastructure/repository"
)

func TestNewHistoryService(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
	assert.NotNil(t, service.logger)
}

func TestNewHistoryService_PanicsOnNilRepo(t *testing.T) {
	logger := slog.Default()

	assert.Panics(t, func() {
		NewHistoryService(nil, logger)
	})
}

func TestGetHistory_Success(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	expectedEntries := []*repository.HistoryEntry{
		{
			ID:             "entry-1",
			RequestID:      "req-1",
			ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
			StatusCode:     200,
			ResponseTimeMs: 100,
		},
		{
			ID:             "entry-2",
			RequestID:      "req-2",
			ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
			StatusCode:     404,
			ResponseTimeMs: 50,
		},
	}

	repo.On("FindAll", mock.Anything, 10).Return(expectedEntries, nil)

	entries, err := service.GetHistory(context.Background(), 10)

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
	assert.Equal(t, "entry-1", entries[0].ID)
	assert.Equal(t, 200, entries[0].StatusCode)

	repo.AssertExpectations(t)
}

func TestGetHistory_Error(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	repo.On("FindAll", mock.Anything, 10).Return(nil, errors.New("database error"))

	entries, err := service.GetHistory(context.Background(), 10)

	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "failed to retrieve history")

	repo.AssertExpectations(t)
}

func TestGetHistory_NoLimit(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	expectedEntries := []*repository.HistoryEntry{
		{ID: "entry-1", StatusCode: 200},
	}

	repo.On("FindAll", mock.Anything, 0).Return(expectedEntries, nil)

	entries, err := service.GetHistory(context.Background(), 0)

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 1)

	repo.AssertExpectations(t)
}

func TestGetRequestHistory_Success(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	requestID := "req-1"
	expectedEntries := []*repository.HistoryEntry{
		{
			ID:             "entry-1",
			RequestID:      requestID,
			ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
			StatusCode:     200,
			ResponseTimeMs: 100,
		},
		{
			ID:             "entry-2",
			RequestID:      requestID,
			ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
			StatusCode:     200,
			ResponseTimeMs: 120,
		},
	}

	repo.On("FindByRequestID", mock.Anything, requestID, 5).Return(expectedEntries, nil)

	entries, err := service.GetRequestHistory(context.Background(), requestID, 5)

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
	assert.Equal(t, requestID, entries[0].RequestID)

	repo.AssertExpectations(t)
}

func TestGetRequestHistory_Error(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	requestID := "req-1"

	repo.On("FindByRequestID", mock.Anything, requestID, 5).Return(nil, errors.New("database error"))

	entries, err := service.GetRequestHistory(context.Background(), requestID, 5)

	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "failed to retrieve request history")

	repo.AssertExpectations(t)
}

func TestDeleteHistory_Success(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	historyID := "entry-1"

	repo.On("Delete", mock.Anything, historyID).Return(nil)

	err := service.DeleteHistory(context.Background(), historyID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteHistory_Error(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	historyID := "entry-1"

	repo.On("Delete", mock.Anything, historyID).Return(errors.New("not found"))

	err := service.DeleteHistory(context.Background(), historyID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete history entry")

	repo.AssertExpectations(t)
}

func TestCleanupOldHistory_Success(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	daysToKeep := 30

	// Calculate expected cutoff timestamp.
	cutoffTime := time.Now().UTC().AddDate(0, 0, -daysToKeep)
	expectedCutoff := cutoffTime.Format(time.RFC3339)

	repo.On("DeleteOlderThan", mock.Anything, mock.MatchedBy(func(timestamp string) bool {
		// Check that the timestamp is close to our expected cutoff.
		// Allow for small time differences due to test execution time.
		parsedTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return false
		}
		diff := parsedTime.Sub(cutoffTime).Abs()
		return diff < 1*time.Second
	})).Return(int64(5), nil)

	count, err := service.CleanupOldHistory(context.Background(), daysToKeep)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	repo.AssertExpectations(t)

	// Verify the timestamp format is correct.
	parsedTime, parseErr := time.Parse(time.RFC3339, expectedCutoff)
	assert.NoError(t, parseErr)
	assert.NotNil(t, parsedTime)
}

func TestCleanupOldHistory_NegativeDays(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	count, err := service.CleanupOldHistory(context.Background(), -5)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "daysToKeep must be non-negative")
}

func TestCleanupOldHistory_Error(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	repo.On("DeleteOlderThan", mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))

	count, err := service.CleanupOldHistory(context.Background(), 30)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "failed to cleanup old history")

	repo.AssertExpectations(t)
}

func TestCleanupOldHistory_ZeroDays(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	// Zero days means delete everything older than today.
	repo.On("DeleteOlderThan", mock.Anything, mock.Anything).Return(int64(10), nil)

	count, err := service.CleanupOldHistory(context.Background(), 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)

	repo.AssertExpectations(t)
}

func TestSaveExecution_Success(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	entry := &repository.HistoryEntry{
		ID:             "entry-1",
		RequestID:      "req-1",
		ExecutedAt:     time.Now().UTC().Format(time.RFC3339),
		StatusCode:     200,
		Status:         "200 OK",
		ResponseTimeMs: 100,
		ResponseBody:   "test response",
	}

	repo.On("Save", mock.Anything, entry).Return(nil)

	err := service.SaveExecution(context.Background(), entry)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSaveExecution_NilEntry(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	err := service.SaveExecution(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "history entry cannot be nil")
}

func TestSaveExecution_Error(t *testing.T) {
	repo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewHistoryService(repo, logger)

	entry := &repository.HistoryEntry{
		ID:        "entry-1",
		RequestID: "req-1",
	}

	repo.On("Save", mock.Anything, entry).Return(errors.New("database error"))

	err := service.SaveExecution(context.Background(), entry)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save execution to history")

	repo.AssertExpectations(t)
}
