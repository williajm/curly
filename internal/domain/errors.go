// Package domain provides core business models and validation logic for curly.
//
// It defines the fundamental types (Request, Response, AuthConfig) and their
// behaviors without dependencies on external frameworks or infrastructure.
package domain

import "errors"

// Sentinel errors for request validation
var (
	// ErrInvalidURL indicates the provided URL is malformed or has an unsupported scheme
	ErrInvalidURL = errors.New("invalid URL format")

	// ErrInvalidMethod indicates the HTTP method is not supported
	ErrInvalidMethod = errors.New("invalid HTTP method")

	// ErrEmptyURL indicates no URL was provided
	ErrEmptyURL = errors.New("URL cannot be empty")

	// ErrEmptyMethod indicates no HTTP method was provided
	ErrEmptyMethod = errors.New("HTTP method cannot be empty")

	// ErrInvalidHeaderName indicates a header name contains invalid characters
	ErrInvalidHeaderName = errors.New("invalid header name")

	// ErrInvalidQueryParam indicates a query parameter name is invalid
	ErrInvalidQueryParam = errors.New("invalid query parameter name")
)

// Sentinel errors for authentication configuration
var (
	// ErrUnsupportedAuth indicates the authentication type is not recognized
	ErrUnsupportedAuth = errors.New("unsupported authentication type")

	// ErrInvalidAuthConfig indicates the authentication configuration is incomplete or malformed
	ErrInvalidAuthConfig = errors.New("invalid authentication configuration")

	// ErrMissingUsername indicates username is required but not provided
	ErrMissingUsername = errors.New("username is required for basic auth")

	// ErrMissingPassword indicates password is required but not provided
	ErrMissingPassword = errors.New("password is required for basic auth")

	// ErrMissingToken indicates bearer token is required but not provided
	ErrMissingToken = errors.New("token is required for bearer auth")

	// ErrMissingAPIKey indicates API key is required but not provided
	ErrMissingAPIKey = errors.New("API key is required")

	// ErrMissingAPIKeyName indicates API key name is required but not provided
	ErrMissingAPIKeyName = errors.New("API key name is required")

	// ErrInvalidAPIKeyLocation indicates the API key location is not supported
	ErrInvalidAPIKeyLocation = errors.New("invalid API key location (must be 'header' or 'query')")
)
