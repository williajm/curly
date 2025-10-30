package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/williajm/curly/internal/domain"
)

// TestNewClient verifies that NewClient creates a client with proper configuration.
func TestNewClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client := NewClient(nil)
		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			Timeout:      5 * time.Second,
			MaxRedirects: 5,
		}
		client := NewClient(config)
		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
	})
}

// TestDefaultConfig verifies default configuration values.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout to be 30s, got %v", config.Timeout)
	}

	if config.MaxRedirects != 10 {
		t.Errorf("expected MaxRedirects to be 10, got %d", config.MaxRedirects)
	}

	if config.FollowRedirects != true {
		t.Error("expected FollowRedirects to be true")
	}

	if config.InsecureSkipTLS != false {
		t.Error("expected InsecureSkipTLS to be false")
	}
}

// TestExecute_AllHTTPMethods tests all supported HTTP methods.
func TestExecute_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create test server that echoes the method
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected method %s, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Method: %s", method)
			}))
			defer server.Close()

			// Create client and request
			client := NewClient(nil)
			req := domain.NewRequestWithMethodAndURL(method, server.URL)

			// Execute request
			ctx := context.Background()
			resp, err := client.Execute(ctx, req)

			// Verify response
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			// HEAD requests don't return a body per HTTP spec
			if method == "HEAD" {
				if resp.Body != "" {
					t.Errorf("expected empty body for HEAD request, got %q", resp.Body)
				}
			} else {
				expectedBody := fmt.Sprintf("Method: %s", method)
				if resp.Body != expectedBody {
					t.Errorf("expected body %q, got %q", expectedBody, resp.Body)
				}
			}
		})
	}
}

// TestExecute_BasicAuth tests HTTP Basic Authentication.
func TestExecute_BasicAuth(t *testing.T) {
	expectedUsername := "testuser"
	expectedPassword := "testpass"

	// Create test server that validates basic auth
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "No auth provided")
			return
		}

		if username != expectedUsername || password != expectedPassword {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid credentials")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Authenticated")
	}))
	defer server.Close()

	tests := []struct {
		name           string
		auth           domain.AuthConfig
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid credentials",
			auth:           domain.NewBasicAuth(expectedUsername, expectedPassword),
			expectedStatus: http.StatusOK,
			expectedBody:   "Authenticated",
		},
		{
			name:           "invalid credentials",
			auth:           domain.NewBasicAuth("wrong", "wrong"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid credentials",
		},
		{
			name:           "no auth",
			auth:           domain.NewNoAuth(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "No auth provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(nil)
			req := domain.NewRequestWithMethodAndURL("GET", server.URL)
			req.SetAuth(tt.auth)

			ctx := context.Background()
			resp, err := client.Execute(ctx, req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.Body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, resp.Body)
			}
		})
	}
}

// TestExecute_BearerAuth tests Bearer Token Authentication.
func TestExecute_BearerAuth(t *testing.T) {
	expectedToken := "secret-token-12345"

	// Create test server that validates bearer token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "No token provided")
			return
		}

		expectedAuth := fmt.Sprintf("Bearer %s", expectedToken)
		if authHeader != expectedAuth {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid token")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Authenticated")
	}))
	defer server.Close()

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			token:          expectedToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "Authenticated",
		},
		{
			name:           "invalid token",
			token:          "wrong-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(nil)
			req := domain.NewRequestWithMethodAndURL("GET", server.URL)
			req.SetAuth(domain.NewBearerAuth(tt.token))

			ctx := context.Background()
			resp, err := client.Execute(ctx, req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.Body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, resp.Body)
			}
		})
	}
}

// TestExecute_APIKeyAuth tests API Key Authentication.
func TestExecute_APIKeyAuth(t *testing.T) {
	expectedKey := "X-API-Key"
	expectedValue := "secret-api-key"

	t.Run("header location", func(t *testing.T) {
		// Create test server that validates API key in header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(expectedKey)
			if apiKey != expectedValue {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Invalid API key")
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Authenticated")
		}))
		defer server.Close()

		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)
		req.SetAuth(domain.NewAPIKeyAuth(expectedKey, expectedValue, domain.APIKeyLocationHeader))

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("query location", func(t *testing.T) {
		// Create test server that validates API key in query
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.URL.Query().Get("api_key")
			if apiKey != expectedValue {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Invalid API key")
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Authenticated")
		}))
		defer server.Close()

		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)
		req.SetAuth(domain.NewAPIKeyAuth("api_key", expectedValue, domain.APIKeyLocationQuery))

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

// TestExecute_Headers tests custom header handling.
func TestExecute_Headers(t *testing.T) {
	// Create test server that echoes headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the custom headers as JSON
		headers := make(map[string]string)
		headers["X-Custom-Header"] = r.Header.Get("X-Custom-Header")
		headers["User-Agent"] = r.Header.Get("User-Agent")
		headers["Accept"] = r.Header.Get("Accept")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(headers)
	}))
	defer server.Close()

	client := NewClient(nil)
	req := domain.NewRequestWithMethodAndURL("GET", server.URL)
	req.SetHeader("X-Custom-Header", "CustomValue")
	req.SetHeader("User-Agent", "Curly/1.0")
	req.SetHeader("Accept", "application/json")

	ctx := context.Background()
	resp, err := client.Execute(ctx, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify response headers were received
	var receivedHeaders map[string]string
	if err := json.Unmarshal([]byte(resp.Body), &receivedHeaders); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if receivedHeaders["X-Custom-Header"] != "CustomValue" {
		t.Errorf("expected X-Custom-Header to be 'CustomValue', got %q", receivedHeaders["X-Custom-Header"])
	}
}

// TestExecute_QueryParams tests query parameter handling.
func TestExecute_QueryParams(t *testing.T) {
	// Create test server that echoes query parameters
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]string)
		for key := range r.URL.Query() {
			params[key] = r.URL.Query().Get(key)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(params)
	}))
	defer server.Close()

	t.Run("with query params", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)
		req.SetQueryParam("name", "test")
		req.SetQueryParam("page", "1")
		req.SetQueryParam("limit", "10")

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var params map[string]string
		if err := json.Unmarshal([]byte(resp.Body), &params); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if params["name"] != "test" {
			t.Errorf("expected name=test, got %q", params["name"])
		}
		if params["page"] != "1" {
			t.Errorf("expected page=1, got %q", params["page"])
		}
	})

	t.Run("merges with existing URL params", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL+"?existing=value")
		req.SetQueryParam("new", "param")

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var params map[string]string
		if err := json.Unmarshal([]byte(resp.Body), &params); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if params["existing"] != "value" {
			t.Errorf("expected existing=value, got %q", params["existing"])
		}
		if params["new"] != "param" {
			t.Errorf("expected new=param, got %q", params["new"])
		}
	})
}

// TestExecute_RequestBody tests request body handling.
func TestExecute_RequestBody(t *testing.T) {
	// Create test server that echoes request body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(body)
	}))
	defer server.Close()

	tests := []struct {
		name        string
		method      string
		body        string
		contentType string
	}{
		{
			name:        "POST with JSON",
			method:      "POST",
			body:        `{"name":"test","value":123}`,
			contentType: "application/json",
		},
		{
			name:        "PUT with JSON",
			method:      "PUT",
			body:        `{"id":1,"status":"updated"}`,
			contentType: "application/json",
		},
		{
			name:        "PATCH with JSON",
			method:      "PATCH",
			body:        `{"status":"active"}`,
			contentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(nil)
			req := domain.NewRequestWithMethodAndURL(tt.method, server.URL)
			req.Body = tt.body
			req.SetHeader("Content-Type", tt.contentType)

			ctx := context.Background()
			resp, err := client.Execute(ctx, req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Body != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, resp.Body)
			}
		})
	}
}

// TestExecute_Timeout tests timeout handling.
func TestExecute_Timeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("request completes before timeout", func(t *testing.T) {
		config := &Config{
			Timeout:      500 * time.Millisecond,
			MaxRedirects: 10,
		}
		client := NewClient(config)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("request times out", func(t *testing.T) {
		config := &Config{
			Timeout:      50 * time.Millisecond,
			MaxRedirects: 10,
		}
		client := NewClient(config)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}

		if !strings.Contains(err.Error(), "timeout") {
			t.Errorf("expected timeout error, got: %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected context canceled error, got nil")
		}

		if !strings.Contains(err.Error(), "cancel") {
			t.Errorf("expected context canceled error, got: %v", err)
		}
	})
}

// TestExecute_Redirects tests redirect handling.
func TestExecute_Redirects(t *testing.T) {
	redirectCount := 0
	var server *httptest.Server

	// Create test server that redirects
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			redirectCount++
			if redirectCount <= 3 {
				http.Redirect(w, r, "/redirect", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}

		if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Success")
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("follows redirects", func(t *testing.T) {
		redirectCount = 0
		config := &Config{
			Timeout:         5 * time.Second,
			MaxRedirects:    10,
			FollowRedirects: true,
		}
		client := NewClient(config)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL+"/redirect")

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		if resp.Body != "Success" {
			t.Errorf("expected body 'Success', got %q", resp.Body)
		}
	})

	t.Run("stops at max redirects", func(t *testing.T) {
		redirectCount = 0
		config := &Config{
			Timeout:         5 * time.Second,
			MaxRedirects:    2,
			FollowRedirects: true,
		}
		client := NewClient(config)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL+"/redirect")

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected redirect error, got nil")
		}

		if !strings.Contains(err.Error(), "redirect") {
			t.Errorf("expected redirect error, got: %v", err)
		}
	})

	t.Run("does not follow redirects when disabled", func(t *testing.T) {
		redirectCount = 0
		config := &Config{
			Timeout:         5 * time.Second,
			MaxRedirects:    10,
			FollowRedirects: false,
		}
		client := NewClient(config)
		req := domain.NewRequestWithMethodAndURL("GET", server.URL+"/redirect")

		ctx := context.Background()
		resp, err := client.Execute(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusFound {
			t.Errorf("expected status 302, got %d", resp.StatusCode)
		}
	})
}

// TestExecute_ErrorHandling tests error scenarios.
func TestExecute_ErrorHandling(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", "not-a-valid-url")

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("INVALID", server.URL)

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("server not reachable", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", "http://localhost:1")

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected connection error, got nil")
		}
	})

	t.Run("DNS lookup failure", func(t *testing.T) {
		client := NewClient(nil)
		req := domain.NewRequestWithMethodAndURL("GET", "http://invalid-domain-that-does-not-exist-12345.com")

		ctx := context.Background()
		_, err := client.Execute(ctx, req)

		if err == nil {
			t.Fatal("expected DNS error, got nil")
		}
	})
}

// TestExecute_ResponseMetadata tests response metadata collection.
func TestExecute_ResponseMetadata(t *testing.T) {
	responseBody := "Test response body"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Custom-Header", "CustomValue")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	client := NewClient(nil)
	req := domain.NewRequestWithMethodAndURL("GET", server.URL)

	ctx := context.Background()
	startTime := time.Now()
	resp, err := client.Execute(ctx, req)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify headers
	if resp.GetHeader("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got %q", resp.GetHeader("Content-Type"))
	}

	if resp.GetHeader("X-Custom-Header") != "CustomValue" {
		t.Errorf("expected X-Custom-Header 'CustomValue', got %q", resp.GetHeader("X-Custom-Header"))
	}

	// Verify body
	if resp.Body != responseBody {
		t.Errorf("expected body %q, got %q", responseBody, resp.Body)
	}

	// Verify content length
	expectedLength := int64(len(responseBody))
	if resp.ContentLength != expectedLength {
		t.Errorf("expected ContentLength %d, got %d", expectedLength, resp.ContentLength)
	}

	// Verify timing
	if resp.Duration <= 0 {
		t.Error("expected Duration to be positive")
	}

	if resp.Duration > elapsed+100*time.Millisecond {
		t.Errorf("Duration %v should not exceed actual elapsed time %v", resp.Duration, elapsed)
	}

	// Verify timestamp
	if resp.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}

	// Verify request ID
	if resp.RequestID != req.ID {
		t.Errorf("expected RequestID %q, got %q", req.ID, resp.RequestID)
	}
}

// TestExecute_StatusCodes tests various HTTP status codes.
func TestExecute_StatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		statusText string
	}{
		{"OK", http.StatusOK, "200 OK"},
		{"Created", http.StatusCreated, "201 Created"},
		{"BadRequest", http.StatusBadRequest, "400 Bad Request"},
		{"Unauthorized", http.StatusUnauthorized, "401 Unauthorized"},
		{"NotFound", http.StatusNotFound, "404 Not Found"},
		{"InternalServerError", http.StatusInternalServerError, "500 Internal Server Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := NewClient(nil)
			req := domain.NewRequestWithMethodAndURL("GET", server.URL)

			ctx := context.Background()
			resp, err := client.Execute(ctx, req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, resp.StatusCode)
			}

			if resp.Status != tc.statusText {
				t.Errorf("expected status %q, got %q", tc.statusText, resp.Status)
			}
		})
	}
}
