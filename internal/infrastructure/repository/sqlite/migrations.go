package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a single database migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// runMigrations executes all pending migrations in order.
// It tracks which migrations have been applied using a schema_migrations table.
func runMigrations(db *sql.DB, migrationsPath string) error {
	// Create migrations tracking table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migration files
	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

// createMigrationsTable creates the table that tracks applied migrations.
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := db.Exec(query)
	return err
}

// loadMigrations reads all migration files from the specified directory.
// Migration files should be named: NNN_description.sql (e.g., 001_initial_schema.sql)
func loadMigrations(migrationsPath string) ([]Migration, error) {
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// No migrations directory, return empty list (not an error)
		return []Migration{}, nil
	}

	// Read directory contents
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		migration, err := parseMigrationFile(filepath.Join(migrationsPath, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file and extracts version, name, and SQL.
// Expected format: NNN_description.sql
func parseMigrationFile(filePath string) (Migration, error) {
	filename := filepath.Base(filePath)

	// Extract version from filename (before first underscore)
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s (expected: NNN_description.sql)", filename)
	}

	var version int
	_, err := fmt.Sscanf(parts[0], "%d", &version)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version number in filename: %s", filename)
	}

	// Extract name (remove .sql extension)
	name := strings.TrimSuffix(parts[1], ".sql")

	// Read SQL content
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read migration file: %w", err)
	}

	return Migration{
		Version: version,
		Name:    name,
		SQL:     string(sqlBytes),
	}, nil
}

// getAppliedMigrations returns a set of migration versions that have been applied.
func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	query := `SELECT version FROM schema_migrations`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration executes a single migration within a transaction.
func applyMigration(db *sql.DB, migration Migration) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safe to call even after Commit

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in schema_migrations table
	recordQuery := `INSERT INTO schema_migrations (version, name) VALUES (?, ?)`
	if _, err := tx.Exec(recordQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// embeddedMigrations contains the schema migrations embedded in the code.
var embeddedMigrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		SQL: `
-- Requests table
CREATE TABLE IF NOT EXISTS requests (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	method TEXT NOT NULL,
	url TEXT NOT NULL,
	headers TEXT,
	query_params TEXT,
	body TEXT,
	auth_type TEXT,
	auth_config TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- History table
CREATE TABLE IF NOT EXISTS history (
	id TEXT PRIMARY KEY,
	request_id TEXT,
	executed_at TIMESTAMP NOT NULL,
	status_code INTEGER,
	status TEXT,
	response_time_ms INTEGER,
	response_headers TEXT,
	response_body TEXT,
	error TEXT,
	FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_updated_at ON requests(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_executed_at ON history(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_request_id ON history(request_id);
		`,
	},
}

// MigrateDB runs embedded migrations on the database.
// This is the recommended way to initialize the database schema.
func MigrateDB(db *sql.DB) error {
	// Create migrations tracking table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range embeddedMigrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}
