package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/williajm/curly/internal/domain"
)

// AuthService handles authentication configuration and application to requests.
// It provides factory methods to create auth configurations and validation.
type AuthService struct {
	logger *slog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(logger *slog.Logger) *AuthService {
	if logger == nil {
		logger = slog.Default()
	}

	return &AuthService{
		logger: logger,
	}
}

// CreateAuth creates an authentication configuration from the provided type and credentials.
// Supported types: "none", "basic", "bearer", "apikey".
// Returns an error if the auth type is not supported or credentials are invalid.
func (s *AuthService) CreateAuth(authType string, credentials map[string]string) (domain.AuthConfig, error) {
	authType = strings.ToLower(strings.TrimSpace(authType))

	s.logger.Debug("creating auth config", "type", authType)

	switch authType {
	case "none", "":
		return domain.NewNoAuth(), nil

	case "basic":
		username, hasUsername := credentials["username"]
		password, hasPassword := credentials["password"]

		if !hasUsername || !hasPassword {
			return nil, fmt.Errorf("basic auth requires 'username' and 'password' credentials")
		}

		auth := domain.NewBasicAuth(username, password)
		if err := auth.Validate(); err != nil {
			s.logger.Warn("basic auth validation failed", "error", err)
			return nil, fmt.Errorf("invalid basic auth credentials: %w", err)
		}

		s.logger.Debug("basic auth created successfully", "username", username)
		return auth, nil

	case "bearer":
		token, hasToken := credentials["token"]

		if !hasToken {
			return nil, fmt.Errorf("bearer auth requires 'token' credential")
		}

		auth := domain.NewBearerAuth(token)
		if err := auth.Validate(); err != nil {
			s.logger.Warn("bearer auth validation failed", "error", err)
			return nil, fmt.Errorf("invalid bearer token: %w", err)
		}

		s.logger.Debug("bearer auth created successfully")
		return auth, nil

	case "apikey":
		key, hasKey := credentials["key"]
		value, hasValue := credentials["value"]
		location, hasLocation := credentials["location"]

		if !hasKey || !hasValue || !hasLocation {
			return nil, fmt.Errorf("api key auth requires 'key', 'value', and 'location' credentials")
		}

		// Parse location.
		var apiKeyLocation domain.APIKeyLocation
		switch strings.ToLower(location) {
		case "header":
			apiKeyLocation = domain.APIKeyLocationHeader
		case "query":
			apiKeyLocation = domain.APIKeyLocationQuery
		default:
			return nil, fmt.Errorf("invalid api key location: %s (must be 'header' or 'query')", location)
		}

		auth := domain.NewAPIKeyAuth(key, value, apiKeyLocation)
		if err := auth.Validate(); err != nil {
			s.logger.Warn("api key auth validation failed", "error", err)
			return nil, fmt.Errorf("invalid api key credentials: %w", err)
		}

		s.logger.Debug("api key auth created successfully",
			"key", key,
			"location", location,
		)
		return auth, nil

	default:
		s.logger.Warn("unsupported auth type", "type", authType)
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}
}

// ValidateAuth validates an authentication configuration.
// Returns an error if the configuration is invalid or incomplete.
func (s *AuthService) ValidateAuth(auth domain.AuthConfig) error {
	if auth == nil {
		return fmt.Errorf("auth config cannot be nil")
	}

	s.logger.Debug("validating auth config", "type", auth.Type())

	if err := auth.Validate(); err != nil {
		s.logger.Warn("auth validation failed",
			"type", auth.Type(),
			"error", err,
		)
		return fmt.Errorf("invalid %s auth: %w", auth.Type(), err)
	}

	s.logger.Debug("auth config validated successfully", "type", auth.Type())
	return nil
}

// ApplyAuth applies an authentication configuration to an HTTP request.
// This delegates to the AuthConfig.Apply method after validation.
func (s *AuthService) ApplyAuth(auth domain.AuthConfig, req *http.Request) error {
	if auth == nil {
		return fmt.Errorf("auth config cannot be nil")
	}

	if req == nil {
		return fmt.Errorf("http request cannot be nil")
	}

	s.logger.Debug("applying auth to request",
		"type", auth.Type(),
		"url", req.URL.String(),
	)

	// Validate before applying.
	if err := auth.Validate(); err != nil {
		s.logger.Warn("cannot apply invalid auth",
			"type", auth.Type(),
			"error", err,
		)
		return fmt.Errorf("invalid auth config: %w", err)
	}

	// Apply auth to request.
	if err := auth.Apply(req); err != nil {
		s.logger.Error("failed to apply auth to request",
			"type", auth.Type(),
			"error", err,
		)
		return fmt.Errorf("failed to apply %s auth: %w", auth.Type(), err)
	}

	s.logger.Debug("auth applied successfully",
		"type", auth.Type(),
	)

	return nil
}

// SupportedTypes returns a list of supported authentication types.
func (s *AuthService) SupportedTypes() []string {
	return []string{
		"none",
		"basic",
		"bearer",
		"apikey",
	}
}
