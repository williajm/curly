package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	UI       UIConfig       `mapstructure:"ui"`
	History  HistoryConfig  `mapstructure:"history"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	Timeout            time.Duration `mapstructure:"timeout"`
	MaxRedirects       int           `mapstructure:"max_redirects"`
	FollowRedirects    bool          `mapstructure:"follow_redirects"`
	InsecureSkipTLS    bool          `mapstructure:"insecure_skip_tls"`
}

// UIConfig holds UI preferences
type UIConfig struct {
	Theme              string `mapstructure:"theme"`
	SyntaxHighlighting bool   `mapstructure:"syntax_highlighting"`
	ShowResponseTime   bool   `mapstructure:"show_response_time"`
	DefaultTab         string `mapstructure:"default_tab"`
}

// HistoryConfig holds history management settings
type HistoryConfig struct {
	MaxEntries       int  `mapstructure:"max_entries"`
	AutoCleanup      bool `mapstructure:"auto_cleanup"`
	CleanupAfterDays int  `mapstructure:"cleanup_after_days"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Level   string `mapstructure:"level"`
}

// Load loads configuration from file, environment variables, and defaults
// It returns the merged configuration and any error encountered
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file path if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config locations
		configDir, err := getConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config directory: %w", err)
		}

		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(configDir)
		v.AddConfigPath(".") // Also look in current directory
	}

	// Read environment variables with CURLY_ prefix
	v.SetEnvPrefix("CURLY")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional, only error on read errors (not file not found)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand paths
	if err := expandPaths(&cfg); err != nil {
		return nil, fmt.Errorf("failed to expand paths: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	homeDir, _ := os.UserHomeDir()

	// Database defaults
	v.SetDefault("database.path", filepath.Join(homeDir, ".local", "share", "curly", "curly.db"))

	// HTTP defaults
	v.SetDefault("http.timeout", "30s")
	v.SetDefault("http.max_redirects", 10)
	v.SetDefault("http.follow_redirects", true)
	v.SetDefault("http.insecure_skip_tls", false)

	// UI defaults
	v.SetDefault("ui.theme", "dark")
	v.SetDefault("ui.syntax_highlighting", true)
	v.SetDefault("ui.show_response_time", true)
	v.SetDefault("ui.default_tab", "request")

	// History defaults
	v.SetDefault("history.max_entries", 1000)
	v.SetDefault("history.auto_cleanup", true)
	v.SetDefault("history.cleanup_after_days", 90)

	// Logging defaults
	v.SetDefault("logging.enabled", true)
	v.SetDefault("logging.path", filepath.Join(homeDir, ".cache", "curly", "curly.log"))
	v.SetDefault("logging.level", "info")
}

// expandPaths expands ~ and environment variables in file paths
func expandPaths(cfg *Config) error {
	var err error

	cfg.Database.Path, err = expandPath(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to expand database path: %w", err)
	}

	cfg.Logging.Path, err = expandPath(cfg.Logging.Path)
	if err != nil {
		return fmt.Errorf("failed to expand logging path: %w", err)
	}

	return nil
}

// expandPath expands ~ to home directory and environment variables
func expandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand ~
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	return path, nil
}

// getConfigDir returns the configuration directory following XDG Base Directory spec
func getConfigDir() (string, error) {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "curly"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "curly"), nil
}

// EnsureDirectories creates necessary directories for the application
func EnsureDirectories(cfg *Config) error {
	// Create database directory
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Create log directory
	if cfg.Logging.Enabled {
		logDir := filepath.Dir(cfg.Logging.Path)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Create config directory
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}
