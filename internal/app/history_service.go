package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/williajm/curly/internal/infrastructure/repository"
)

// HistoryService manages request execution history.
// It provides operations to retrieve, save, and cleanup history entries.
type HistoryService struct {
	repo   repository.HistoryRepository
	logger *slog.Logger
}

// NewHistoryService creates a new HistoryService with the provided dependencies.
// The repository is required and must not be nil.
func NewHistoryService(repo repository.HistoryRepository, logger *slog.Logger) *HistoryService {
	if repo == nil {
		panic("history repository cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &HistoryService{
		repo:   repo,
		logger: logger,
	}
}

// GetHistory retrieves all history entries with optional pagination.
// If limit is 0, all entries are returned.
// Results are ordered by executed_at descending (newest first).
func (s *HistoryService) GetHistory(ctx context.Context, limit int) ([]*repository.HistoryEntry, error) {
	s.logger.Debug("retrieving history", "limit", limit)

	entries, err := s.repo.FindAll(ctx, limit)
	if err != nil {
		s.logger.Error("failed to retrieve history",
			"limit", limit,
			"error", err,
		)
		return nil, fmt.Errorf("failed to retrieve history: %w", err)
	}

	s.logger.Debug("history retrieved successfully",
		"count", len(entries),
		"limit", limit,
	)

	return entries, nil
}

// GetRequestHistory retrieves all history entries for a specific request.
// If limit is 0, all entries for the request are returned.
// Results are ordered by executed_at descending (newest first).
func (s *HistoryService) GetRequestHistory(ctx context.Context, requestID string, limit int) ([]*repository.HistoryEntry, error) {
	s.logger.Debug("retrieving request history",
		"request_id", requestID,
		"limit", limit,
	)

	entries, err := s.repo.FindByRequestID(ctx, requestID, limit)
	if err != nil {
		s.logger.Error("failed to retrieve request history",
			"request_id", requestID,
			"limit", limit,
			"error", err,
		)
		return nil, fmt.Errorf("failed to retrieve request history: %w", err)
	}

	s.logger.Debug("request history retrieved successfully",
		"request_id", requestID,
		"count", len(entries),
		"limit", limit,
	)

	return entries, nil
}

// DeleteHistory removes a history entry by ID.
// Returns an error if the entry is not found.
func (s *HistoryService) DeleteHistory(ctx context.Context, id string) error {
	s.logger.Info("deleting history entry", "history_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete history entry",
			"history_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to delete history entry: %w", err)
	}

	s.logger.Info("history entry deleted successfully", "history_id", id)
	return nil
}

// CleanupOldHistory deletes all history entries older than the specified number of days.
// Returns the number of entries deleted.
func (s *HistoryService) CleanupOldHistory(ctx context.Context, daysToKeep int) (int64, error) {
	if daysToKeep < 0 {
		return 0, fmt.Errorf("daysToKeep must be non-negative, got: %d", daysToKeep)
	}

	// Calculate the cutoff timestamp.
	cutoffTime := time.Now().UTC().AddDate(0, 0, -daysToKeep)
	cutoffTimestamp := cutoffTime.Format(time.RFC3339)

	s.logger.Info("cleaning up old history",
		"days_to_keep", daysToKeep,
		"cutoff_timestamp", cutoffTimestamp,
	)

	count, err := s.repo.DeleteOlderThan(ctx, cutoffTimestamp)
	if err != nil {
		s.logger.Error("failed to cleanup old history",
			"days_to_keep", daysToKeep,
			"cutoff_timestamp", cutoffTimestamp,
			"error", err,
		)
		return 0, fmt.Errorf("failed to cleanup old history: %w", err)
	}

	s.logger.Info("old history cleaned up successfully",
		"days_to_keep", daysToKeep,
		"deleted_count", count,
	)

	return count, nil
}

// SaveExecution saves a request execution result to history.
// This is typically called after executing a request to record its outcome.
func (s *HistoryService) SaveExecution(ctx context.Context, entry *repository.HistoryEntry) error {
	if entry == nil {
		return fmt.Errorf("history entry cannot be nil")
	}

	s.logger.Debug("saving execution to history",
		"history_id", entry.ID,
		"request_id", entry.RequestID,
	)

	if err := s.repo.Save(ctx, entry); err != nil {
		s.logger.Error("failed to save execution to history",
			"history_id", entry.ID,
			"request_id", entry.RequestID,
			"error", err,
		)
		return fmt.Errorf("failed to save execution to history: %w", err)
	}

	s.logger.Debug("execution saved to history successfully",
		"history_id", entry.ID,
		"request_id", entry.RequestID,
		"status_code", entry.StatusCode,
	)

	return nil
}
