package domain

import (
	"testing"
	"time"
)

// TestNewRequest tests the NewRequest constructor
func TestNewRequest(t *testing.T) {
	req := NewRequest()

	if req.ID == "" {
		t.Error("expected ID to be generated")
	}
	if req.Method != "GET" {
		t.Errorf("expected default method to be 'GET', got '%s'", req.Method)
	}
	if req.Headers == nil {
		t.Error("expected Headers map to be initialized")
	}
	if req.QueryParams == nil {
		t.Error("expected QueryParams map to be initialized")
	}
	if req.AuthConfig == nil {
		t.Error("expected AuthConfig to be initialized")
	}
	if req.AuthConfig.Type() != "none" {
		t.Errorf("expected default auth type to be 'none', got '%s'", req.AuthConfig.Type())
	}
	if req.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if req.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

// TestNewRequestWithMethodAndURL tests the convenience constructor
func TestNewRequestWithMethodAndURL(t *testing.T) {
	req := NewRequestWithMethodAndURL("POST", "https://api.example.com/users")

	if req.Method != "POST" {
		t.Errorf("expected method to be 'POST', got '%s'", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("expected URL to be set, got '%s'", req.URL)
	}
	if req.ID == "" {
		t.Error("expected ID to be generated")
	}
}

// TestValidateMethod tests the ValidateMethod function
func TestValidateMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr error
	}{
		{"GET", "GET", nil},
		{"POST", "POST", nil},
		{"PUT", "PUT", nil},
		{"PATCH", "PATCH", nil},
		{"DELETE", "DELETE", nil},
		{"HEAD", "HEAD", nil},
		{"OPTIONS", "OPTIONS", nil},
		{"lowercase get", "get", nil}, // Should be case-insensitive
		{"lowercase post", "post", nil},
		{"empty method", "", ErrEmptyMethod},
		{"whitespace method", "   ", ErrEmptyMethod},
		{"invalid method", "INVALID", ErrInvalidMethod},
		{"unsupported method", "TRACE", ErrInvalidMethod},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			req.Method = tt.method

			err := req.ValidateMethod()
			if err != tt.wantErr {
				t.Errorf("ValidateMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateURL tests the ValidateURL function
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"valid http", "http://example.com", nil},
		{"valid https", "https://example.com", nil},
		{"valid with path", "https://api.example.com/v1/users", nil},
		{"valid with query", "https://example.com/search?q=test", nil},
		{"valid with port", "https://example.com:8080/api", nil},
		{"valid localhost", "http://localhost:3000", nil},
		{"valid IP", "http://192.168.1.1", nil},
		{"empty url", "", ErrEmptyURL},
		{"whitespace url", "   ", ErrEmptyURL},
		{"invalid scheme ftp", "ftp://example.com", ErrInvalidURL},
		{"invalid scheme file", "file:///path/to/file", ErrInvalidURL},
		{"no scheme", "example.com", ErrInvalidURL},
		{"malformed url", "://invalid", ErrInvalidURL},
		{"no host", "https://", ErrInvalidURL},
		{"no host with path", "https:///path", ErrInvalidURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			req.URL = tt.url

			err := req.ValidateURL()
			if err != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateHeaders tests the ValidateHeaders function
func TestValidateHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		wantErr error
	}{
		{
			name:    "valid headers",
			headers: map[string]string{"Content-Type": "application/json", "Accept": "application/json"},
			wantErr: nil,
		},
		{
			name:    "empty headers map",
			headers: map[string]string{},
			wantErr: nil,
		},
		{
			name:    "nil headers map",
			headers: nil,
			wantErr: nil,
		},
		{
			name:    "header with empty name",
			headers: map[string]string{"": "value"},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name:    "header with whitespace name",
			headers: map[string]string{"   ": "value"},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name:    "header with colon in name",
			headers: map[string]string{"Invalid:Header": "value"},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name:    "header with newline in name",
			headers: map[string]string{"Invalid\nHeader": "value"},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name:    "header with carriage return in name",
			headers: map[string]string{"Invalid\rHeader": "value"},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name:    "valid header with empty value",
			headers: map[string]string{"X-Custom-Header": ""},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			req.Headers = tt.headers

			err := req.ValidateHeaders()
			if err != tt.wantErr {
				t.Errorf("ValidateHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateQueryParams tests the ValidateQueryParams function
func TestValidateQueryParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		wantErr     error
	}{
		{
			name:        "valid params",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			wantErr:     nil,
		},
		{
			name:        "empty params map",
			queryParams: map[string]string{},
			wantErr:     nil,
		},
		{
			name:        "nil params map",
			queryParams: nil,
			wantErr:     nil,
		},
		{
			name:        "param with empty name",
			queryParams: map[string]string{"": "value"},
			wantErr:     ErrInvalidQueryParam,
		},
		{
			name:        "param with whitespace name",
			queryParams: map[string]string{"   ": "value"},
			wantErr:     ErrInvalidQueryParam,
		},
		{
			name:        "valid param with empty value",
			queryParams: map[string]string{"filter": ""},
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			req.QueryParams = tt.queryParams

			err := req.ValidateQueryParams()
			if err != tt.wantErr {
				t.Errorf("ValidateQueryParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidate tests the overall Validate function
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Request)
		wantErr error
	}{
		{
			name: "fully valid request",
			setup: func(r *Request) {
				r.Method = "POST"
				r.URL = "https://api.example.com/users"
				r.Headers = map[string]string{"Content-Type": "application/json"}
				r.QueryParams = map[string]string{"version": "v1"}
				r.Body = `{"name": "test"}`
				r.AuthConfig = NewBearerAuth("token123")
			},
			wantErr: nil,
		},
		{
			name: "minimal valid request",
			setup: func(r *Request) {
				r.Method = "GET"
				r.URL = "https://example.com"
			},
			wantErr: nil,
		},
		{
			name: "invalid method",
			setup: func(r *Request) {
				r.Method = "INVALID"
				r.URL = "https://example.com"
			},
			wantErr: ErrInvalidMethod,
		},
		{
			name: "invalid url",
			setup: func(r *Request) {
				r.Method = "GET"
				r.URL = "not-a-url"
			},
			wantErr: ErrInvalidURL,
		},
		{
			name: "invalid header",
			setup: func(r *Request) {
				r.Method = "GET"
				r.URL = "https://example.com"
				r.Headers = map[string]string{"Invalid:Header": "value"}
			},
			wantErr: ErrInvalidHeaderName,
		},
		{
			name: "invalid query param",
			setup: func(r *Request) {
				r.Method = "GET"
				r.URL = "https://example.com"
				r.QueryParams = map[string]string{"": "value"}
			},
			wantErr: ErrInvalidQueryParam,
		},
		{
			name: "invalid auth config",
			setup: func(r *Request) {
				r.Method = "GET"
				r.URL = "https://example.com"
				r.AuthConfig = NewBasicAuth("", "")
			},
			wantErr: ErrMissingUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			tt.setup(req)

			err := req.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSetHeader tests the SetHeader method
func TestSetHeader(t *testing.T) {
	req := NewRequest()
	originalTime := req.UpdatedAt

	// Small delay to ensure time difference
	time.Sleep(2 * time.Millisecond)

	t.Run("set new header", func(t *testing.T) {
		req.SetHeader("Content-Type", "application/json")

		if req.Headers["Content-Type"] != "application/json" {
			t.Errorf("expected header to be set, got '%s'", req.Headers["Content-Type"])
		}
		if !req.UpdatedAt.After(originalTime) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("update existing header", func(t *testing.T) {
		req.SetHeader("Content-Type", "text/plain")

		if req.Headers["Content-Type"] != "text/plain" {
			t.Errorf("expected header to be updated, got '%s'", req.Headers["Content-Type"])
		}
	})

	t.Run("remove header with empty value", func(t *testing.T) {
		req.SetHeader("Content-Type", "")

		if _, exists := req.Headers["Content-Type"]; exists {
			t.Error("expected header to be removed")
		}
	})
}

// TestSetQueryParam tests the SetQueryParam method
func TestSetQueryParam(t *testing.T) {
	req := NewRequest()
	originalTime := req.UpdatedAt

	// Small delay to ensure time difference
	time.Sleep(2 * time.Millisecond)

	t.Run("set new query param", func(t *testing.T) {
		req.SetQueryParam("page", "1")

		if req.QueryParams["page"] != "1" {
			t.Errorf("expected query param to be set, got '%s'", req.QueryParams["page"])
		}
		if !req.UpdatedAt.After(originalTime) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("update existing query param", func(t *testing.T) {
		req.SetQueryParam("page", "2")

		if req.QueryParams["page"] != "2" {
			t.Errorf("expected query param to be updated, got '%s'", req.QueryParams["page"])
		}
	})

	t.Run("remove query param with empty value", func(t *testing.T) {
		req.SetQueryParam("page", "")

		if _, exists := req.QueryParams["page"]; exists {
			t.Error("expected query param to be removed")
		}
	})
}

// TestSetAuth tests the SetAuth method
func TestSetAuth(t *testing.T) {
	req := NewRequest()
	originalTime := req.UpdatedAt

	// Small delay to ensure time difference
	time.Sleep(2 * time.Millisecond)

	auth := NewBearerAuth("token123")
	req.SetAuth(auth)

	if req.AuthConfig != auth {
		t.Error("expected auth config to be set")
	}
	if !req.UpdatedAt.After(originalTime) {
		t.Error("expected UpdatedAt to be updated")
	}
}

// TestClone tests the Clone method
func TestClone(t *testing.T) {
	original := NewRequest()
	original.ID = "test-id"
	original.Name = "Test Request"
	original.Method = "POST"
	original.URL = "https://api.example.com/users"
	original.Headers = map[string]string{"Content-Type": "application/json"}
	original.QueryParams = map[string]string{"version": "v1"}
	original.Body = `{"name": "test"}`
	original.AuthConfig = NewBearerAuth("token123")

	clone := original.Clone()

	// Test that all fields are copied
	if clone.ID != original.ID {
		t.Error("ID not copied correctly")
	}
	if clone.Name != original.Name {
		t.Error("Name not copied correctly")
	}
	if clone.Method != original.Method {
		t.Error("Method not copied correctly")
	}
	if clone.URL != original.URL {
		t.Error("URL not copied correctly")
	}
	if clone.Body != original.Body {
		t.Error("Body not copied correctly")
	}
	if clone.AuthConfig != original.AuthConfig {
		t.Error("AuthConfig not copied correctly")
	}

	// Test that maps are deep copied
	clone.Headers["X-Custom"] = "value"
	if _, exists := original.Headers["X-Custom"]; exists {
		t.Error("modifying clone's headers affected original")
	}

	clone.QueryParams["new"] = "param"
	if _, exists := original.QueryParams["new"]; exists {
		t.Error("modifying clone's query params affected original")
	}

	// Test that original map values are preserved
	if clone.Headers["Content-Type"] != "application/json" {
		t.Error("header not copied correctly")
	}
	if clone.QueryParams["version"] != "v1" {
		t.Error("query param not copied correctly")
	}
}

// TestIsBodyAllowed tests the IsBodyAllowed method
func TestIsBodyAllowed(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		allowed bool
	}{
		{"POST allows body", "POST", true},
		{"PUT allows body", "PUT", true},
		{"PATCH allows body", "PATCH", true},
		{"GET does not allow body", "GET", false},
		{"DELETE does not allow body", "DELETE", false},
		{"HEAD does not allow body", "HEAD", false},
		{"OPTIONS does not allow body", "OPTIONS", false},
		{"lowercase post", "post", true},
		{"lowercase get", "get", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest()
			req.Method = tt.method

			if req.IsBodyAllowed() != tt.allowed {
				t.Errorf("IsBodyAllowed() = %v, want %v", req.IsBodyAllowed(), tt.allowed)
			}
		})
	}
}

// TestSupportedMethods verifies the list of supported methods
func TestSupportedMethods(t *testing.T) {
	expected := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	if len(SupportedMethods) != len(expected) {
		t.Errorf("expected %d supported methods, got %d", len(expected), len(SupportedMethods))
	}

	for _, method := range expected {
		found := false
		for _, supported := range SupportedMethods {
			if supported == method {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected method '%s' to be in SupportedMethods", method)
		}
	}
}

// TestRequestInitialization tests that requests are properly initialized with nil maps
func TestRequestInitialization(t *testing.T) {
	t.Run("NewRequest initializes maps", func(t *testing.T) {
		req := NewRequest()

		// Should not panic when adding to maps
		req.Headers["test"] = "value"
		req.QueryParams["test"] = "value"

		if req.Headers["test"] != "value" {
			t.Error("failed to add to initialized Headers map")
		}
		if req.QueryParams["test"] != "value" {
			t.Error("failed to add to initialized QueryParams map")
		}
	})

	t.Run("SetHeader initializes map if nil", func(t *testing.T) {
		req := &Request{}
		req.SetHeader("test", "value")

		if req.Headers == nil {
			t.Error("SetHeader should initialize Headers map")
		}
		if req.Headers["test"] != "value" {
			t.Error("header not set correctly")
		}
	})

	t.Run("SetQueryParam initializes map if nil", func(t *testing.T) {
		req := &Request{}
		req.SetQueryParam("test", "value")

		if req.QueryParams == nil {
			t.Error("SetQueryParam should initialize QueryParams map")
		}
		if req.QueryParams["test"] != "value" {
			t.Error("query param not set correctly")
		}
	})
}

// BenchmarkValidateURL benchmarks URL validation
func BenchmarkValidateURL(b *testing.B) {
	req := NewRequest()
	req.URL = "https://api.example.com/v1/users?page=1&limit=10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.ValidateURL()
	}
}

// BenchmarkValidate benchmarks full request validation
func BenchmarkValidate(b *testing.B) {
	req := NewRequest()
	req.Method = "POST"
	req.URL = "https://api.example.com/users"
	req.Headers = map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": "Bearer token123",
	}
	req.QueryParams = map[string]string{"version": "v1"}
	req.Body = `{"name": "test"}`
	req.AuthConfig = NewBearerAuth("token123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

// BenchmarkClone benchmarks request cloning
func BenchmarkClone(b *testing.B) {
	req := NewRequest()
	req.Method = "POST"
	req.URL = "https://api.example.com/users"
	req.Headers = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	req.QueryParams = map[string]string{"version": "v1"}
	req.Body = `{"name": "test"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Clone()
	}
}
