package domain

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// AuthConfig represents an authentication configuration that can be applied to HTTP requests.
// Different authentication mechanisms (Basic, Bearer, API Key) implement this interface.
type AuthConfig interface {
	// Apply applies the authentication configuration to the given HTTP request.
	// Returns an error if the authentication cannot be applied.
	Apply(req *http.Request) error

	// Type returns the authentication type (e.g., "basic", "bearer", "apikey", "none").
	Type() string

	// Validate checks if the authentication configuration is valid and complete.
	Validate() error
}

// NoAuth represents no authentication (the default).
type NoAuth struct{}

// NewNoAuth creates a new NoAuth configuration.
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// Apply does nothing since no authentication is required.
func (a *NoAuth) Apply(req *http.Request) error {
	return nil
}

// Type returns "none".
func (a *NoAuth) Type() string {
	return "none"
}

// Validate always returns nil since no configuration is required.
func (a *NoAuth) Validate() error {
	return nil
}

// BasicAuth represents HTTP Basic Authentication.
// It encodes the username and password in the Authorization header.
type BasicAuth struct {
	Username string
	Password string
}

// NewBasicAuth creates a new BasicAuth configuration.
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

// Apply adds the Basic Authentication header to the request.
func (a *BasicAuth) Apply(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	// Encode username:password in base64
	credentials := fmt.Sprintf("%s:%s", a.Username, a.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))

	return nil
}

// Type returns "basic".
func (a *BasicAuth) Type() string {
	return "basic"
}

// Validate checks if username and password are provided.
func (a *BasicAuth) Validate() error {
	if strings.TrimSpace(a.Username) == "" {
		return ErrMissingUsername
	}
	if a.Password == "" {
		return ErrMissingPassword
	}
	return nil
}

// BearerAuth represents Bearer Token Authentication.
// It adds the token to the Authorization header with the "Bearer" prefix.
type BearerAuth struct {
	Token string
}

// NewBearerAuth creates a new BearerAuth configuration.
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		Token: token,
	}
}

// Apply adds the Bearer Authentication header to the request.
func (a *BearerAuth) Apply(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.Token))
	return nil
}

// Type returns "bearer".
func (a *BearerAuth) Type() string {
	return "bearer"
}

// Validate checks if the token is provided.
func (a *BearerAuth) Validate() error {
	if strings.TrimSpace(a.Token) == "" {
		return ErrMissingToken
	}
	return nil
}

// APIKeyLocation specifies where the API key should be placed.
type APIKeyLocation string

const (
	// APIKeyLocationHeader indicates the API key should be in a header.
	APIKeyLocationHeader APIKeyLocation = "header"

	// APIKeyLocationQuery indicates the API key should be in a query parameter.
	APIKeyLocationQuery APIKeyLocation = "query"
)

// APIKeyAuth represents API Key Authentication.
// The API key can be placed in either a header or query parameter.
type APIKeyAuth struct {
	Key      string         // The name of the header or query parameter
	Value    string         // The API key value
	Location APIKeyLocation // Where to place the key (header or query)
}

// NewAPIKeyAuth creates a new APIKeyAuth configuration.
func NewAPIKeyAuth(key, value string, location APIKeyLocation) *APIKeyAuth {
	return &APIKeyAuth{
		Key:      key,
		Value:    value,
		Location: location,
	}
}

// Apply adds the API key to the request as either a header or query parameter.
func (a *APIKeyAuth) Apply(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	switch a.Location {
	case APIKeyLocationHeader:
		req.Header.Set(a.Key, a.Value)
	case APIKeyLocationQuery:
		q := req.URL.Query()
		q.Set(a.Key, a.Value)
		req.URL.RawQuery = q.Encode()
	default:
		return ErrInvalidAPIKeyLocation
	}

	return nil
}

// Type returns "apikey".
func (a *APIKeyAuth) Type() string {
	return "apikey"
}

// Validate checks if all required fields are provided and valid.
func (a *APIKeyAuth) Validate() error {
	if strings.TrimSpace(a.Key) == "" {
		return ErrMissingAPIKeyName
	}
	if strings.TrimSpace(a.Value) == "" {
		return ErrMissingAPIKey
	}
	if a.Location != APIKeyLocationHeader && a.Location != APIKeyLocationQuery {
		return ErrInvalidAPIKeyLocation
	}
	return nil
}
