package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/williajm/curly/internal/infrastructure/repository"
)

// HistoryRepository implements repository.HistoryRepository using SQLite.
type HistoryRepository struct {
	db *sql.DB
}

// NewHistoryRepository creates a new SQLite-backed history repository.
func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// Save persists a history entry to the database.
func (r *HistoryRepository) Save(ctx context.Context, entry *repository.HistoryEntry) error {
	if entry == nil {
		return fmt.Errorf("history entry cannot be nil")
	}

	query := `
		INSERT INTO history (id, request_id, executed_at, status_code, status, response_time_ms, response_headers, response_body, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		nullString(entry.RequestID),
		entry.ExecutedAt,
		entry.StatusCode,
		entry.Status,
		entry.ResponseTimeMs,
		entry.ResponseHeaders,
		entry.ResponseBody,
		nullString(entry.Error),
	)

	if err != nil {
		return fmt.Errorf("failed to save history entry: %w", err)
	}

	return nil
}

// FindByID retrieves a history entry by its ID.
func (r *HistoryRepository) FindByID(ctx context.Context, id string) (*repository.HistoryEntry, error) {
	query := `
		SELECT id, request_id, executed_at, status_code, status, response_time_ms, response_headers, response_body, error
		FROM history
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	entry := &repository.HistoryEntry{}
	var requestID, errorMsg sql.NullString

	err := row.Scan(
		&entry.ID,
		&requestID,
		&entry.ExecutedAt,
		&entry.StatusCode,
		&entry.Status,
		&entry.ResponseTimeMs,
		&entry.ResponseHeaders,
		&entry.ResponseBody,
		&errorMsg,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan history entry: %w", err)
	}

	entry.RequestID = requestID.String
	entry.Error = errorMsg.String

	return entry, nil
}

// FindAll retrieves all history entries ordered by executed_at descending.
func (r *HistoryRepository) FindAll(ctx context.Context, limit int) ([]*repository.HistoryEntry, error) {
	query := `
		SELECT id, request_id, executed_at, status_code, status, response_time_ms, response_headers, response_body, error
		FROM history
		ORDER BY executed_at DESC
	`

	// Add LIMIT clause if specified
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	return scanHistoryEntries(rows)
}

// FindByRequestID retrieves all history entries for a specific request.
func (r *HistoryRepository) FindByRequestID(ctx context.Context, requestID string, limit int) ([]*repository.HistoryEntry, error) {
	query := `
		SELECT id, request_id, executed_at, status_code, status, response_time_ms, response_headers, response_body, error
		FROM history
		WHERE request_id = ?
		ORDER BY executed_at DESC
	`

	// Add LIMIT clause if specified
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history by request ID: %w", err)
	}
	defer rows.Close()

	return scanHistoryEntries(rows)
}

// Delete removes a history entry from the database.
func (r *HistoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM history WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete history entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteOlderThan removes all history entries older than the specified timestamp.
func (r *HistoryRepository) DeleteOlderThan(ctx context.Context, timestamp string) (int64, error) {
	query := `DELETE FROM history WHERE executed_at < ?`

	result, err := r.db.ExecContext(ctx, query, timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old history entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// scanHistoryEntries is a helper function to scan multiple rows into HistoryEntry structs.
func scanHistoryEntries(rows *sql.Rows) ([]*repository.HistoryEntry, error) {
	var entries []*repository.HistoryEntry

	for rows.Next() {
		entry := &repository.HistoryEntry{}
		var requestID, errorMsg sql.NullString

		err := rows.Scan(
			&entry.ID,
			&requestID,
			&entry.ExecutedAt,
			&entry.StatusCode,
			&entry.Status,
			&entry.ResponseTimeMs,
			&entry.ResponseHeaders,
			&entry.ResponseBody,
			&errorMsg,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan history entry: %w", err)
		}

		entry.RequestID = requestID.String
		entry.Error = errorMsg.String

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// nullString converts a string to sql.NullString, setting Valid to false if the string is empty.
func nullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
