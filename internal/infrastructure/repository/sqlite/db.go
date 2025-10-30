// Package sqlite provides SQLite database implementations of repository interfaces.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // SQLite driver
)

// Config holds database configuration options.
type Config struct {
	// Path is the file path to the SQLite database.
	// Use ":memory:" for in-memory databases (useful for testing).
	Path string

	// MigrationsPath is the directory containing migration SQL files.
	// If empty, migrations are not run automatically.
	MigrationsPath string
}

// DefaultConfig returns the default database configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Path:           filepath.Join(homeDir, ".local", "share", "curly", "curly.db"),
		MigrationsPath: "migrations",
	}
}

// Open opens a connection to the SQLite database and applies performance optimizations.
// It also runs migrations if MigrationsPath is specified.
func Open(config *Config) (*sql.DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure the database directory exists (unless using in-memory database).
	if config.Path != ":memory:" {
		dbDir := filepath.Dir(config.Path)
		if err := os.MkdirAll(dbDir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Open the database connection.
	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply performance pragmas.
	if err := applyPragmas(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to apply pragmas: %w", err)
	}

	// Run migrations if path is specified.
	if config.MigrationsPath != "" {
		if err := runMigrations(db, config.MigrationsPath); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return db, nil
}

// applyPragmas configures SQLite for optimal performance.
func applyPragmas(db *sql.DB) error {
	pragmas := []string{
		// Enable Write-Ahead Logging for better concurrency.
		"PRAGMA journal_mode = WAL",
		// Normal synchronous mode is safe with WAL and much faster.
		"PRAGMA synchronous = NORMAL",
		// 64MB cache size for better performance.
		"PRAGMA cache_size = -64000",
		// Enable foreign key constraints.
		"PRAGMA foreign_keys = ON",
		// Reduce memory usage for temp tables.
		"PRAGMA temp_store = MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}
