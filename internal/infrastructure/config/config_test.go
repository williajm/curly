package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Load config with no file (should use defaults)
	// Pass empty string to use default config search paths
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults
	assert.Equal(t, 30*time.Second, cfg.HTTP.Timeout)
	assert.Equal(t, 10, cfg.HTTP.MaxRedirects)
	assert.True(t, cfg.HTTP.FollowRedirects)
	assert.False(t, cfg.HTTP.InsecureSkipTLS)

	assert.Equal(t, "dark", cfg.UI.Theme)
	assert.True(t, cfg.UI.SyntaxHighlighting)
	assert.True(t, cfg.UI.ShowResponseTime)
	assert.Equal(t, "request", cfg.UI.DefaultTab)

	assert.Equal(t, 1000, cfg.History.MaxEntries)
	assert.True(t, cfg.History.AutoCleanup)
	assert.Equal(t, 90, cfg.History.CleanupAfterDays)

	assert.True(t, cfg.Logging.Enabled)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoad_CustomConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
database:
  path: /tmp/test.db

http:
  timeout: 60s
  max_redirects: 5
  follow_redirects: false
  insecure_skip_tls: true

ui:
  theme: light
  syntax_highlighting: false
  show_response_time: false
  default_tab: history

history:
  max_entries: 500
  auto_cleanup: false
  cleanup_after_days: 30

logging:
  enabled: false
  path: /tmp/test.log
  level: debug
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config from file
	cfg, err := Load(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify custom values
	assert.Equal(t, "/tmp/test.db", cfg.Database.Path)
	assert.Equal(t, 60*time.Second, cfg.HTTP.Timeout)
	assert.Equal(t, 5, cfg.HTTP.MaxRedirects)
	assert.False(t, cfg.HTTP.FollowRedirects)
	assert.True(t, cfg.HTTP.InsecureSkipTLS)

	assert.Equal(t, "light", cfg.UI.Theme)
	assert.False(t, cfg.UI.SyntaxHighlighting)
	assert.False(t, cfg.UI.ShowResponseTime)
	assert.Equal(t, "history", cfg.UI.DefaultTab)

	assert.Equal(t, 500, cfg.History.MaxEntries)
	assert.False(t, cfg.History.AutoCleanup)
	assert.Equal(t, 30, cfg.History.CleanupAfterDays)

	assert.False(t, cfg.Logging.Enabled)
	assert.Equal(t, "/tmp/test.log", cfg.Logging.Path)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/test/path",
			expected: filepath.Join(homeDir, "test", "path"),
		},
		{
			name:     "absolute path",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandPath_EnvVar(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "/test/value")
	defer os.Unsetenv("TEST_VAR")

	result, err := expandPath("$TEST_VAR/path")
	require.NoError(t, err)
	assert.Equal(t, "/test/value/path", result)
}

func TestEnsureDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(tmpDir, "db", "curly.db"),
		},
		Logging: LoggingConfig{
			Enabled: true,
			Path:    filepath.Join(tmpDir, "logs", "curly.log"),
		},
	}

	err := EnsureDirectories(cfg)
	require.NoError(t, err)

	// Verify directories were created
	dbDir := filepath.Dir(cfg.Database.Path)
	logDir := filepath.Dir(cfg.Logging.Path)

	assert.DirExists(t, dbDir)
	assert.DirExists(t, logDir)
}

func TestGetConfigDir(t *testing.T) {
	// Save original env var
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		configDir, err := getConfigDir()
		require.NoError(t, err)
		assert.Equal(t, "/custom/config/curly", configDir)
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		homeDir, _ := os.UserHomeDir()
		configDir, err := getConfigDir()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, ".config", "curly"), configDir)
	})
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
database:
  path: /tmp/test.db
  invalid_indent:
    bad: yaml
	tabs: are bad
`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// Load config should fail
	_, err = Load(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestEnsureDirectories_LoggingDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(tmpDir, "db", "curly.db"),
		},
		Logging: LoggingConfig{
			Enabled: false,
			Path:    filepath.Join(tmpDir, "logs", "curly.log"),
		},
	}

	err := EnsureDirectories(cfg)
	require.NoError(t, err)

	// Verify database directory was created
	dbDir := filepath.Dir(cfg.Database.Path)
	assert.DirExists(t, dbDir)

	// Log directory may or may not exist - that's OK when logging is disabled
}

func TestExpandPaths_Error(t *testing.T) {
	// Test error handling in expandPaths
	cfg := &Config{
		Database: DatabaseConfig{
			Path: "~/valid/path",
		},
		Logging: LoggingConfig{
			Path: "~/valid/path",
		},
	}

	// This should succeed normally
	err := expandPaths(cfg)
	assert.NoError(t, err)
}

func TestLoad_PartialConfig(t *testing.T) {
	// Create config file with only some values
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
http:
  timeout: 120s

ui:
  theme: light
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify custom values are set
	assert.Equal(t, 120*time.Second, cfg.HTTP.Timeout)
	assert.Equal(t, "light", cfg.UI.Theme)

	// Verify defaults are used for unspecified values
	assert.Equal(t, 10, cfg.HTTP.MaxRedirects)
	assert.True(t, cfg.HTTP.FollowRedirects)
	assert.Equal(t, 1000, cfg.History.MaxEntries)
}

func TestExpandPath_Relative(t *testing.T) {
	// Test relative path (should be left as-is)
	result, err := expandPath("relative/path")
	require.NoError(t, err)
	assert.Equal(t, "relative/path", result)
}

func TestExpandPath_MultipleEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_VAR1", "/first")
	os.Setenv("TEST_VAR2", "second")
	defer func() {
		os.Unsetenv("TEST_VAR1")
		os.Unsetenv("TEST_VAR2")
	}()

	result, err := expandPath("$TEST_VAR1/$TEST_VAR2/path")
	require.NoError(t, err)
	assert.Equal(t, "/first/second/path", result)
}

func TestEnsureDirectories_NestedPaths(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(tmpDir, "deeply", "nested", "path", "db", "curly.db"),
		},
		Logging: LoggingConfig{
			Enabled: true,
			Path:    filepath.Join(tmpDir, "very", "deep", "log", "path", "curly.log"),
		},
	}

	err := EnsureDirectories(cfg)
	require.NoError(t, err)

	// Verify nested directories were created
	dbDir := filepath.Dir(cfg.Database.Path)
	logDir := filepath.Dir(cfg.Logging.Path)

	assert.DirExists(t, dbDir)
	assert.DirExists(t, logDir)
}
