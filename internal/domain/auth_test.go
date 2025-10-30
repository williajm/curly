package domain

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

// Test constants for repeated test values.
const (
	testValue = "value"
)

// TestNoAuth tests the NoAuth implementation.
func TestNoAuth(t *testing.T) {
	auth := NewNoAuth()

	t.Run("Type", func(t *testing.T) {
		if auth.Type() != "none" {
			t.Errorf("expected type 'none', got '%s'", auth.Type())
		}
	})

	t.Run("Validate", func(t *testing.T) {
		if err := auth.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Apply", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		if err := auth.Apply(req); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// Should not add any headers.
		if req.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got '%s'", req.Header.Get("Authorization"))
		}
	})
}

// TestBasicAuth tests the BasicAuth implementation.
func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{
			name:     "valid credentials",
			username: "user",
			password: "pass",
			wantErr:  nil,
		},
		{
			name:     "empty username",
			username: "",
			password: "pass",
			wantErr:  ErrMissingUsername,
		},
		{
			name:     "whitespace username",
			username: "   ",
			password: "pass",
			wantErr:  ErrMissingUsername,
		},
		{
			name:     "empty password",
			username: "user",
			password: "",
			wantErr:  ErrMissingPassword,
		},
		{
			name:     "both empty",
			username: "",
			password: "",
			wantErr:  ErrMissingUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewBasicAuth(tt.username, tt.password)

			// Test Type.
			if auth.Type() != "basic" {
				t.Errorf("expected type 'basic', got '%s'", auth.Type())
			}

			// Test Validate.
			err := auth.Validate()
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test Apply.
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			err = auth.Apply(req)
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If valid, check the Authorization header.
			if tt.wantErr == nil {
				authHeader := req.Header.Get("Authorization")
				if authHeader == "" {
					t.Error("expected Authorization header to be set")
				}
				if len(authHeader) < 6 || authHeader[:6] != "Basic " {
					t.Errorf("expected Authorization header to start with 'Basic ', got '%s'", authHeader)
				}
			}
		})
	}
}

// TestBasicAuthEncoding tests that BasicAuth properly encodes credentials.
func TestBasicAuthEncoding(t *testing.T) {
	auth := NewBasicAuth("testuser", "testpass")
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	err := auth.Apply(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	// The base64 encoding of "testuser:testpass" is "dGVzdHVzZXI6dGVzdHBhc3M=".
	expected := "Basic dGVzdHVzZXI6dGVzdHBhc3M="
	if authHeader != expected {
		t.Errorf("expected '%s', got '%s'", expected, authHeader)
	}
}

// TestBearerAuth tests the BearerAuth implementation.
func TestBearerAuth(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token",
			token:   "my-secret-token",
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrMissingToken,
		},
		{
			name:    "whitespace token",
			token:   "   ",
			wantErr: ErrMissingToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewBearerAuth(tt.token)

			// Test Type.
			if auth.Type() != "bearer" {
				t.Errorf("expected type 'bearer', got '%s'", auth.Type())
			}

			// Test Validate.
			err := auth.Validate()
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test Apply.
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			err = auth.Apply(req)
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If valid, check the Authorization header.
			if tt.wantErr == nil {
				authHeader := req.Header.Get("Authorization")
				expected := "Bearer " + tt.token
				if authHeader != expected {
					t.Errorf("expected '%s', got '%s'", expected, authHeader)
				}
			}
		})
	}
}

// TestAPIKeyAuth tests the APIKeyAuth implementation.
func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		location APIKeyLocation
		wantErr  error
	}{
		{
			name:     "valid header auth",
			key:      "X-API-Key",
			value:    "secret123",
			location: APIKeyLocationHeader,
			wantErr:  nil,
		},
		{
			name:     "valid query auth",
			key:      "api_key",
			value:    "secret123",
			location: APIKeyLocationQuery,
			wantErr:  nil,
		},
		{
			name:     "empty key",
			key:      "",
			value:    "secret123",
			location: APIKeyLocationHeader,
			wantErr:  ErrMissingAPIKeyName,
		},
		{
			name:     "whitespace key",
			key:      "   ",
			value:    "secret123",
			location: APIKeyLocationHeader,
			wantErr:  ErrMissingAPIKeyName,
		},
		{
			name:     "empty value",
			key:      "X-API-Key",
			value:    "",
			location: APIKeyLocationHeader,
			wantErr:  ErrMissingAPIKey,
		},
		{
			name:     "whitespace value",
			key:      "X-API-Key",
			value:    "   ",
			location: APIKeyLocationHeader,
			wantErr:  ErrMissingAPIKey,
		},
		{
			name:     "invalid location",
			key:      "X-API-Key",
			value:    "secret123",
			location: "invalid",
			wantErr:  ErrInvalidAPIKeyLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAPIKeyAuth(tt.key, tt.value, tt.location)

			// Test Type.
			if auth.Type() != "apikey" {
				t.Errorf("expected type 'apikey', got '%s'", auth.Type())
			}

			// Test Validate.
			err := auth.Validate()
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test Apply.
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			err = auth.Apply(req)
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If valid, check that the key is in the right location.
			if tt.wantErr == nil {
				switch tt.location {
				case APIKeyLocationHeader:
					headerValue := req.Header.Get(tt.key)
					if headerValue != tt.value {
						t.Errorf("expected header '%s' to be '%s', got '%s'", tt.key, tt.value, headerValue)
					}
				case APIKeyLocationQuery:
					queryValue := req.URL.Query().Get(tt.key)
					if queryValue != tt.value {
						t.Errorf("expected query param '%s' to be '%s', got '%s'", tt.key, tt.value, queryValue)
					}
				}
			}
		})
	}
}

// TestAPIKeyAuthHeader tests that API key is correctly added to headers.
func TestAPIKeyAuthHeader(t *testing.T) {
	auth := NewAPIKeyAuth("X-API-Key", "my-secret-key", APIKeyLocationHeader)
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	err := auth.Apply(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Header.Get("X-API-Key") != "my-secret-key" {
		t.Errorf("expected X-API-Key header to be 'my-secret-key', got '%s'", req.Header.Get("X-API-Key"))
	}
}

// TestAPIKeyAuthQuery tests that API key is correctly added to query parameters.
func TestAPIKeyAuthQuery(t *testing.T) {
	auth := NewAPIKeyAuth("api_key", "my-secret-key", APIKeyLocationQuery)
	req, _ := http.NewRequest("GET", "http://example.com/path?existing=value", nil)

	err := auth.Apply(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := req.URL.Query()
	if query.Get("api_key") != "my-secret-key" {
		t.Errorf("expected api_key query param to be 'my-secret-key', got '%s'", query.Get("api_key"))
	}

	// Ensure existing query params are preserved.
	if query.Get("existing") != testValue {
		t.Errorf("expected existing query param to be preserved, got '%s'", query.Get("existing"))
	}
}

// TestAuthConfigInterface ensures all auth types implement the AuthConfig interface.
func TestAuthConfigInterface(_ *testing.T) {
	var _ AuthConfig = (*NoAuth)(nil)
	var _ AuthConfig = (*BasicAuth)(nil)
	var _ AuthConfig = (*BearerAuth)(nil)
	var _ AuthConfig = (*APIKeyAuth)(nil)
}

// TestAPIKeyAuthPreservesExistingURL tests that applying API key auth preserves the URL structure.
func TestAPIKeyAuthPreservesExistingURL(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		key         string
		value       string
		wantPath    string
	}{
		{
			name:        "simple path",
			originalURL: "http://example.com/api/users",
			key:         "api_key",
			value:       "secret",
			wantPath:    "/api/users",
		},
		{
			name:        "with fragment",
			originalURL: "http://example.com/page#section",
			key:         "token",
			value:       "abc123",
			wantPath:    "/page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAPIKeyAuth(tt.key, tt.value, APIKeyLocationQuery)
			req, _ := http.NewRequest("GET", tt.originalURL, nil)

			err := auth.Apply(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.URL.Path != tt.wantPath {
				t.Errorf("expected path '%s', got '%s'", tt.wantPath, req.URL.Path)
			}

			if req.URL.Query().Get(tt.key) != tt.value {
				t.Errorf("expected query param '%s' to be '%s', got '%s'", tt.key, tt.value, req.URL.Query().Get(tt.key))
			}
		})
	}
}

// TestAPIKeyLocationConstants verifies the location constants.
func TestAPIKeyLocationConstants(t *testing.T) {
	if APIKeyLocationHeader != "header" {
		t.Errorf("expected APIKeyLocationHeader to be 'header', got '%s'", APIKeyLocationHeader)
	}
	if APIKeyLocationQuery != "query" {
		t.Errorf("expected APIKeyLocationQuery to be 'query', got '%s'", APIKeyLocationQuery)
	}
}

// BenchmarkBasicAuthApply benchmarks the BasicAuth Apply method.
func BenchmarkBasicAuthApply(b *testing.B) {
	auth := NewBasicAuth("user", "password")
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.Apply(req)
	}
}

// BenchmarkBearerAuthApply benchmarks the BearerAuth Apply method.
func BenchmarkBearerAuthApply(b *testing.B) {
	auth := NewBearerAuth("my-token-12345")
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.Apply(req)
	}
}

// BenchmarkAPIKeyAuthHeaderApply benchmarks the APIKeyAuth Apply method with header location.
func BenchmarkAPIKeyAuthHeaderApply(b *testing.B) {
	auth := NewAPIKeyAuth("X-API-Key", "secret-key", APIKeyLocationHeader)
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.Apply(req)
	}
}

// BenchmarkAPIKeyAuthQueryApply benchmarks the APIKeyAuth Apply method with query location.
func BenchmarkAPIKeyAuthQueryApply(b *testing.B) {
	auth := NewAPIKeyAuth("api_key", "secret-key", APIKeyLocationQuery)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		_ = auth.Apply(req)
	}
}

// TestAuthWithSpecialCharacters tests auth implementations with special characters.
func TestAuthWithSpecialCharacters(t *testing.T) {
	t.Run("BasicAuth with special chars", func(t *testing.T) {
		auth := NewBasicAuth("user@example.com", "p@ssw0rd!#$%")
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		err := auth.Apply(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header to be set")
		}
	})

	t.Run("BearerAuth with special chars", func(t *testing.T) {
		auth := NewBearerAuth("tok.en-with_special/chars+123")
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		err := auth.Apply(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if !contains(authHeader, "tok.en-with_special/chars+123") {
			t.Errorf("expected token in Authorization header, got '%s'", authHeader)
		}
	})

	t.Run("APIKeyAuth with special chars in value", func(t *testing.T) {
		auth := NewAPIKeyAuth("X-API-Key", "key!@#$%^&*()", APIKeyLocationHeader)
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		err := auth.Apply(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		keyValue := req.Header.Get("X-API-Key")
		if keyValue != "key!@#$%^&*()" {
			t.Errorf("expected special chars in API key, got '%s'", keyValue)
		}
	})

	t.Run("APIKeyAuth query with URL-unsafe chars", func(t *testing.T) {
		auth := NewAPIKeyAuth("key", "value with spaces & symbols=test", APIKeyLocationQuery)
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		err := auth.Apply(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The URL encoding should handle special characters.
		rawQuery := req.URL.RawQuery
		if rawQuery == "" {
			t.Error("expected query string to be set")
		}

		// Decode and verify.
		parsed, _ := url.ParseQuery(rawQuery)
		if parsed.Get("key") != "value with spaces & symbols=test" {
			t.Errorf("expected decoded value to match, got '%s'", parsed.Get("key"))
		}
	})
}
