package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateDB_Success(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	err = MigrateDB(db)
	require.NoError(t, err)

	// Verify schema_migrations table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should have 1 migration applied")

	// Verify requests table exists
	_, err = db.Exec("SELECT * FROM requests LIMIT 1")
	require.NoError(t, err)

	// Verify history table exists
	_, err = db.Exec("SELECT * FROM history LIMIT 1")
	require.NoError(t, err)

	// Verify indexes exist
	var indexCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='index'
		AND name LIKE 'idx_%'
	`).Scan(&indexCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, indexCount, 4, "should have at least 4 indexes")
}

func TestMigrateDB_Idempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Run migrations first time
	err = MigrateDB(db)
	require.NoError(t, err)

	// Run migrations second time (should be idempotent)
	err = MigrateDB(db)
	require.NoError(t, err)

	// Verify only one migration record exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should still have only 1 migration applied")
}

func TestCreateMigrationsTable(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migrations table
	err = createMigrationsTable(db)
	require.NoError(t, err)

	// Verify table exists and has correct columns
	rows, err := db.Query("PRAGMA table_info(schema_migrations)")
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		require.NoError(t, err)
		columns[name] = true
	}

	assert.True(t, columns["version"], "should have version column")
	assert.True(t, columns["name"], "should have name column")
	assert.True(t, columns["applied_at"], "should have applied_at column")
}

func TestLoadMigrations_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	migrations, err := loadMigrations(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, migrations, "should return empty list for empty directory")
}

func TestLoadMigrations_NonexistentDirectory(t *testing.T) {
	migrations, err := loadMigrations("/nonexistent/path")
	require.NoError(t, err)
	assert.Empty(t, migrations, "should return empty list for nonexistent directory")
}

func TestLoadMigrations_ValidFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test migration files
	migrations := []struct {
		filename string
		content  string
	}{
		{
			filename: "001_initial_schema.sql",
			content:  "CREATE TABLE test1 (id INTEGER);",
		},
		{
			filename: "002_add_users.sql",
			content:  "CREATE TABLE users (id INTEGER, name TEXT);",
		},
		{
			filename: "003_add_indexes.sql",
			content:  "CREATE INDEX idx_users_name ON users(name);",
		},
	}

	for _, m := range migrations {
		path := filepath.Join(tmpDir, m.filename)
		err := os.WriteFile(path, []byte(m.content), 0644)
		require.NoError(t, err)
	}

	// Load migrations
	loaded, err := loadMigrations(tmpDir)
	require.NoError(t, err)
	assert.Len(t, loaded, 3, "should load all migration files")

	// Verify migrations are sorted by version
	assert.Equal(t, 1, loaded[0].Version)
	assert.Equal(t, 2, loaded[1].Version)
	assert.Equal(t, 3, loaded[2].Version)

	// Verify content is loaded
	assert.Contains(t, loaded[0].SQL, "CREATE TABLE test1")
	assert.Contains(t, loaded[1].SQL, "CREATE TABLE users")
	assert.Contains(t, loaded[2].SQL, "CREATE INDEX")
}

func TestLoadMigrations_SkipsNonSQLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create migration files and non-SQL files
	files := []struct {
		filename string
		content  string
	}{
		{"001_schema.sql", "CREATE TABLE test (id INTEGER);"},
		{"002_data.sql", "INSERT INTO test VALUES (1);"},
		{"README.md", "# Migrations"},
		{"script.sh", "#!/bin/bash"},
		{".gitkeep", ""},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.filename)
		err := os.WriteFile(path, []byte(f.content), 0644)
		require.NoError(t, err)
	}

	// Load migrations
	loaded, err := loadMigrations(tmpDir)
	require.NoError(t, err)
	assert.Len(t, loaded, 2, "should load only .sql files")
}

func TestLoadMigrations_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a migration file and a subdirectory
	migrationPath := filepath.Join(tmpDir, "001_schema.sql")
	err := os.WriteFile(migrationPath, []byte("CREATE TABLE test (id INTEGER);"), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(tmpDir, "subdirectory")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Load migrations
	loaded, err := loadMigrations(tmpDir)
	require.NoError(t, err)
	assert.Len(t, loaded, 1, "should load only files, not directories")
}

func TestParseMigrationFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "042_add_feature.sql")

	content := "CREATE TABLE feature (id INTEGER, name TEXT);"
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	migration, err := parseMigrationFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, 42, migration.Version)
	assert.Equal(t, "add_feature", migration.Name)
	assert.Equal(t, content, migration.SQL)
}

func TestParseMigrationFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "no underscore",
			filename: "001schema.sql",
		},
		{
			name:     "no version",
			filename: "schema.sql",
		},
		{
			name:     "invalid version",
			filename: "abc_schema.sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte("CREATE TABLE test (id INTEGER);"), 0644)
			require.NoError(t, err)

			_, err = parseMigrationFile(filePath)
			assert.Error(t, err)
		})
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migrations table
	err = createMigrationsTable(db)
	require.NoError(t, err)

	// Insert some applied migrations
	_, err = db.Exec("INSERT INTO schema_migrations (version, name) VALUES (1, 'initial'), (3, 'feature'), (5, 'bugfix')")
	require.NoError(t, err)

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	require.NoError(t, err)

	assert.True(t, applied[1], "version 1 should be applied")
	assert.False(t, applied[2], "version 2 should not be applied")
	assert.True(t, applied[3], "version 3 should be applied")
	assert.False(t, applied[4], "version 4 should not be applied")
	assert.True(t, applied[5], "version 5 should be applied")
}

func TestApplyMigration_Success(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migrations table
	err = createMigrationsTable(db)
	require.NoError(t, err)

	// Apply a migration
	migration := Migration{
		Version: 1,
		Name:    "test_migration",
		SQL:     "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);",
	}

	err = applyMigration(db, migration)
	require.NoError(t, err)

	// Verify migration was applied
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify table was created
	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'test')")
	require.NoError(t, err)
}

func TestApplyMigration_InvalidSQL(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migrations table
	err = createMigrationsTable(db)
	require.NoError(t, err)

	// Try to apply a migration with invalid SQL
	migration := Migration{
		Version: 1,
		Name:    "bad_migration",
		SQL:     "INVALID SQL SYNTAX HERE;",
	}

	err = applyMigration(db, migration)
	assert.Error(t, err)

	// Verify migration was NOT recorded (transaction rollback)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "migration should not be recorded when SQL fails")
}

func TestRunMigrations_AppliesOnlyPending(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migration files
	migrations := []struct {
		filename string
		content  string
	}{
		{"001_first.sql", "CREATE TABLE table1 (id INTEGER);"},
		{"002_second.sql", "CREATE TABLE table2 (id INTEGER);"},
		{"003_third.sql", "CREATE TABLE table3 (id INTEGER);"},
	}

	for _, m := range migrations {
		path := filepath.Join(tmpDir, m.filename)
		err := os.WriteFile(path, []byte(m.content), 0644)
		require.NoError(t, err)
	}

	// Run migrations first time
	err = runMigrations(db, tmpDir)
	require.NoError(t, err)

	// Verify all migrations were applied
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify all tables exist
	_, err = db.Exec("SELECT * FROM table1 LIMIT 1")
	require.NoError(t, err)
	_, err = db.Exec("SELECT * FROM table2 LIMIT 1")
	require.NoError(t, err)
	_, err = db.Exec("SELECT * FROM table3 LIMIT 1")
	require.NoError(t, err)

	// Add a new migration file
	newMigration := filepath.Join(tmpDir, "004_fourth.sql")
	err = os.WriteFile(newMigration, []byte("CREATE TABLE table4 (id INTEGER);"), 0644)
	require.NoError(t, err)

	// Run migrations again
	err = runMigrations(db, tmpDir)
	require.NoError(t, err)

	// Verify only the new migration was applied
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 4, count)

	// Verify new table exists
	_, err = db.Exec("SELECT * FROM table4 LIMIT 1")
	require.NoError(t, err)
}

func TestMigrateDB_ForeignKeyConstraints(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Apply pragmas to enable foreign keys before migrations
	err = applyPragmas(db)
	require.NoError(t, err)

	// Run migrations
	err = MigrateDB(db)
	require.NoError(t, err)

	// Verify foreign key constraints are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled, "foreign keys should be enabled")

	// Test foreign key constraint by trying to insert invalid reference
	// First insert should succeed (valid request)
	_, err = db.Exec(`
		INSERT INTO requests (id, name, method, url, headers, query_params, body, auth_type, auth_config, created_at, updated_at)
		VALUES ('req-1', 'Test', 'GET', 'http://example.com', '{}', '{}', '', 'none', '{}', datetime('now'), datetime('now'))
	`)
	require.NoError(t, err)

	// This should succeed (valid foreign key)
	_, err = db.Exec(`
		INSERT INTO history (id, request_id, executed_at, status_code, status, response_time_ms, response_headers, response_body, error)
		VALUES ('hist-1', 'req-1', datetime('now'), 200, '200 OK', 100, '{}', '{}', '')
	`)
	require.NoError(t, err)

	// Test cascade delete - delete request should delete history
	_, err = db.Exec("DELETE FROM requests WHERE id = 'req-1'")
	require.NoError(t, err)

	// History should be deleted due to CASCADE
	var historyCount int
	err = db.QueryRow("SELECT COUNT(*) FROM history WHERE id = 'hist-1'").Scan(&historyCount)
	require.NoError(t, err)
	assert.Equal(t, 0, historyCount, "history should be deleted via CASCADE")
}
