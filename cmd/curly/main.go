package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/williajm/curly/internal/app"
	"github.com/williajm/curly/internal/infrastructure/config"
	"github.com/williajm/curly/internal/infrastructure/http"
	"github.com/williajm/curly/internal/infrastructure/repository/sqlite"
	"github.com/williajm/curly/internal/presentation"
	"github.com/williajm/curly/pkg/version"
)

func main() {
	// Command-line flags
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	configFlag := flag.String("config", "", "Path to configuration file")
	dbPathFlag := flag.String("db", "", "Path to SQLite database (overrides config)")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		info := version.Get()
		fmt.Println(info.String())
		os.Exit(0)
	}

	// Initialize and run the application
	if err := run(*configFlag, *dbPathFlag); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run(configPath, dbPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override database path from command line if provided
	if dbPath != "" {
		cfg.Database.Path = dbPath
	}

	// Ensure necessary directories exist
	if err := config.EnsureDirectories(cfg); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Set up logging
	logger, logFile, err := setupLogging(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}
	if logFile != nil {
		defer logFile.Close()
	}
	slog.SetDefault(logger)

	slog.Info("Starting curly",
		"version", version.Get().Version,
		"config_loaded", configPath != "",
		"database_path", cfg.Database.Path,
	)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize database with config
	dbConfig := &sqlite.Config{
		Path:           cfg.Database.Path,
		MigrationsPath: "", // Migrations are embedded in the code
	}

	db, err := sqlite.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		slog.Debug("Closing database connection")
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	// Run embedded migrations
	if err := sqlite.MigrateDB(db); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize repositories
	requestRepo := sqlite.NewRequestRepository(db)
	historyRepo := sqlite.NewHistoryRepository(db)

	// Initialize HTTP client with config
	httpConfig := &http.Config{
		Timeout:         cfg.HTTP.Timeout,
		MaxRedirects:    cfg.HTTP.MaxRedirects,
		FollowRedirects: cfg.HTTP.FollowRedirects,
		InsecureSkipTLS: cfg.HTTP.InsecureSkipTLS,
	}
	httpClient := http.NewClient(httpConfig)

	// Initialize services
	requestService := app.NewRequestService(requestRepo, httpClient, historyRepo, slog.Default())
	historyService := app.NewHistoryService(historyRepo, slog.Default())
	authService := app.NewAuthService(slog.Default())

	// Channel to receive TUI errors
	errChan := make(chan error, 1)

	// Start the TUI in a goroutine
	go func() {
		slog.Info("Starting curly TUI...")
		if err := presentation.RunApp(requestService, historyService, authService); err != nil {
			errChan <- fmt.Errorf("TUI error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either a signal or TUI to exit
	var appErr error
	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down gracefully", "signal", sig)
		// Give the TUI a moment to clean up
		time.Sleep(100 * time.Millisecond)
	case appErr = <-errChan:
		// TUI exited normally or with error
	}

	slog.Info("curly exited", "error", appErr)
	return appErr
}

// setupLogging configures the application logger based on configuration
func setupLogging(cfg *config.Config) (*slog.Logger, *os.File, error) {
	var handler slog.Handler
	var logFile *os.File

	// Parse log level
	level := parseLogLevel(cfg.Logging.Level)

	if cfg.Logging.Enabled && cfg.Logging.Path != "" {
		// Create log file
		var err error
		logFile, err = os.OpenFile(cfg.Logging.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Log to both file and stderr
		multiWriter := io.MultiWriter(logFile, os.Stderr)
		handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		// Log only to stderr
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler), logFile, nil
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
