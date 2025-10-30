package domain

import (
	"testing"
	"time"
)

// TestNewResponse tests the NewResponse constructor
func TestNewResponse(t *testing.T) {
	resp := NewResponse()

	if resp.Headers == nil {
		t.Error("expected Headers map to be initialized")
	}
	if resp.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}
}

// TestIsSuccess tests the IsSuccess method
func TestIsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 edge of 2xx", 299, true},
		{"199 not 2xx", 199, false},
		{"300 redirect", 300, false},
		{"400 client error", 400, false},
		{"500 server error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsSuccess(); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsRedirect tests the IsRedirect method
func TestIsRedirect(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"300 Multiple Choices", 300, true},
		{"301 Moved Permanently", 301, true},
		{"302 Found", 302, true},
		{"304 Not Modified", 304, true},
		{"307 Temporary Redirect", 307, true},
		{"308 Permanent Redirect", 308, true},
		{"399 edge of 3xx", 399, true},
		{"299 not 3xx", 299, false},
		{"400 not 3xx", 400, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsRedirect(); got != tt.want {
				t.Errorf("IsRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsClientError tests the IsClientError method
func TestIsClientError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"400 Bad Request", 400, true},
		{"401 Unauthorized", 401, true},
		{"403 Forbidden", 403, true},
		{"404 Not Found", 404, true},
		{"429 Too Many Requests", 429, true},
		{"499 edge of 4xx", 499, true},
		{"399 not 4xx", 399, false},
		{"500 not 4xx", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsClientError(); got != tt.want {
				t.Errorf("IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsServerError tests the IsServerError method
func TestIsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500 Internal Server Error", 500, true},
		{"501 Not Implemented", 501, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"599 edge of 5xx", 599, true},
		{"499 not 5xx", 499, false},
		{"600 not 5xx", 600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsServerError(); got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsError tests the IsError method
func TestIsError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"400 client error", 400, true},
		{"404 client error", 404, true},
		{"500 server error", 500, true},
		{"503 server error", 503, true},
		{"200 success", 200, false},
		{"301 redirect", 301, false},
		{"100 informational", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsError(); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsInformational tests the IsInformational method
func TestIsInformational(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"100 Continue", 100, true},
		{"101 Switching Protocols", 101, true},
		{"102 Processing", 102, true},
		{"199 edge of 1xx", 199, true},
		{"99 not 1xx", 99, false},
		{"200 not 1xx", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.IsInformational(); got != tt.want {
				t.Errorf("IsInformational() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStatusClass tests the StatusClass method
func TestStatusClass(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       string
	}{
		{"100 informational", 100, "1xx Informational"},
		{"200 success", 200, "2xx Success"},
		{"301 redirect", 301, "3xx Redirect"},
		{"404 client error", 404, "4xx Client Error"},
		{"500 server error", 500, "5xx Server Error"},
		{"999 unknown", 999, "Unknown"},
		{"0 unknown", 0, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.StatusCode = tt.statusCode

			if got := resp.StatusClass(); got != tt.want {
				t.Errorf("StatusClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetHeader tests the GetHeader method
func TestGetHeader(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		searchFor  string
		wantValue  string
		wantExists bool
	}{
		{
			name:       "exact match",
			headers:    map[string]string{"Content-Type": "application/json"},
			searchFor:  "Content-Type",
			wantValue:  "application/json",
			wantExists: true,
		},
		{
			name:       "case insensitive match",
			headers:    map[string]string{"Content-Type": "application/json"},
			searchFor:  "content-type",
			wantValue:  "application/json",
			wantExists: true,
		},
		{
			name:       "case insensitive uppercase",
			headers:    map[string]string{"content-type": "text/plain"},
			searchFor:  "CONTENT-TYPE",
			wantValue:  "text/plain",
			wantExists: true,
		},
		{
			name:       "header not found",
			headers:    map[string]string{"Content-Type": "application/json"},
			searchFor:  "Accept",
			wantValue:  "",
			wantExists: false,
		},
		{
			name:       "empty headers",
			headers:    map[string]string{},
			searchFor:  "Content-Type",
			wantValue:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.Headers = tt.headers

			gotValue := resp.GetHeader(tt.searchFor)
			if gotValue != tt.wantValue {
				t.Errorf("GetHeader() = %v, want %v", gotValue, tt.wantValue)
			}

			gotExists := resp.HasHeader(tt.searchFor)
			if gotExists != tt.wantExists {
				t.Errorf("HasHeader() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

// TestContentType tests the ContentType method
func TestContentType(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name:    "json content type",
			headers: map[string]string{"Content-Type": "application/json"},
			want:    "application/json",
		},
		{
			name:    "json with charset",
			headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			want:    "application/json; charset=utf-8",
		},
		{
			name:    "case insensitive",
			headers: map[string]string{"content-type": "text/html"},
			want:    "text/html",
		},
		{
			name:    "no content type",
			headers: map[string]string{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.Headers = tt.headers

			if got := resp.ContentType(); got != tt.want {
				t.Errorf("ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsJSON tests the IsJSON method
func TestIsJSON(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"application/json", "application/json", true},
		{"application/json with charset", "application/json; charset=utf-8", true},
		{"json api", "application/vnd.api+json", true},
		{"uppercase", "APPLICATION/JSON", true},
		{"mixed case", "Application/Json", true},
		{"text/plain", "text/plain", false},
		{"text/html", "text/html", false},
		{"application/xml", "application/xml", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			if tt.contentType != "" {
				resp.Headers = map[string]string{"Content-Type": tt.contentType}
			}

			if got := resp.IsJSON(); got != tt.want {
				t.Errorf("IsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsXML tests the IsXML method
func TestIsXML(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"application/xml", "application/xml", true},
		{"text/xml", "text/xml", true},
		{"application/xml with charset", "application/xml; charset=utf-8", true},
		{"uppercase", "APPLICATION/XML", true},
		{"mixed case", "Text/Xml", true},
		{"application/json", "application/json", false},
		{"text/plain", "text/plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			if tt.contentType != "" {
				resp.Headers = map[string]string{"Content-Type": tt.contentType}
			}

			if got := resp.IsXML(); got != tt.want {
				t.Errorf("IsXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsHTML tests the IsHTML method
func TestIsHTML(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"text/html", "text/html", true},
		{"text/html with charset", "text/html; charset=utf-8", true},
		{"uppercase", "TEXT/HTML", true},
		{"mixed case", "Text/Html", true},
		{"application/json", "application/json", false},
		{"text/plain", "text/plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			if tt.contentType != "" {
				resp.Headers = map[string]string{"Content-Type": tt.contentType}
			}

			if got := resp.IsHTML(); got != tt.want {
				t.Errorf("IsHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsText tests the IsText method
func TestIsText(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"text/plain", "text/plain", true},
		{"text/plain with charset", "text/plain; charset=utf-8", true},
		{"uppercase", "TEXT/PLAIN", true},
		{"mixed case", "Text/Plain", true},
		{"text/html", "text/html", false},
		{"application/json", "application/json", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			if tt.contentType != "" {
				resp.Headers = map[string]string{"Content-Type": tt.contentType}
			}

			if got := resp.IsText(); got != tt.want {
				t.Errorf("IsText() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDurationMillis tests the DurationMillis method
func TestDurationMillis(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     int64
	}{
		{"100ms", 100 * time.Millisecond, 100},
		{"1 second", 1 * time.Second, 1000},
		{"1.5 seconds", 1500 * time.Millisecond, 1500},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.Duration = tt.duration

			if got := resp.DurationMillis(); got != tt.want {
				t.Errorf("DurationMillis() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDurationSeconds tests the DurationSeconds method
func TestDurationSeconds(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     float64
	}{
		{"100ms", 100 * time.Millisecond, 0.1},
		{"1 second", 1 * time.Second, 1.0},
		{"1.5 seconds", 1500 * time.Millisecond, 1.5},
		{"zero", 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse()
			resp.Duration = tt.duration

			if got := resp.DurationSeconds(); got != tt.want {
				t.Errorf("DurationSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEqualsFold tests the equalsFold helper function
func TestEqualsFold(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"exact match", "Content-Type", "Content-Type", true},
		{"case insensitive", "Content-Type", "content-type", true},
		{"case insensitive reverse", "content-type", "CONTENT-TYPE", true},
		{"mixed case", "Content-Type", "CoNtEnT-TyPe", true},
		{"different strings", "Content-Type", "Accept", false},
		{"different lengths", "Content-Type", "Content-Type-Extra", false},
		{"empty strings", "", "", true},
		{"one empty", "test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalsFold(tt.a, tt.b); got != tt.want {
				t.Errorf("equalsFold(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestContains tests the contains helper function
func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact match", "application/json", "application/json", true},
		{"substring at start", "application/json", "application", true},
		{"substring at end", "application/json", "json", true},
		{"substring in middle", "application/json", "tion/j", true},
		{"case insensitive", "Application/JSON", "application/json", true},
		{"case insensitive reverse", "application/json", "APPLICATION/JSON", true},
		{"not found", "application/json", "xml", false},
		{"empty substring", "test", "", true},
		{"empty string", "", "test", false},
		{"both empty", "", "", true},
		{"partial match", "text/html", "html", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// TestResponseWithAllFields tests a response with all fields populated
func TestResponseWithAllFields(t *testing.T) {
	resp := NewResponse()
	resp.StatusCode = 200
	resp.Status = "200 OK"
	resp.Headers = map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": "123",
	}
	resp.Body = `{"message": "success"}`
	resp.ContentLength = 123
	resp.Duration = 250 * time.Millisecond
	resp.RequestID = "test-request-123"

	// Verify all fields
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if resp.Status != "200 OK" {
		t.Errorf("Status = %s, want '200 OK'", resp.Status)
	}
	if !resp.IsSuccess() {
		t.Error("expected IsSuccess() to be true")
	}
	if resp.IsError() {
		t.Error("expected IsError() to be false")
	}
	if resp.ContentType() != "application/json" {
		t.Errorf("ContentType() = %s, want 'application/json'", resp.ContentType())
	}
	if !resp.IsJSON() {
		t.Error("expected IsJSON() to be true")
	}
	if resp.ContentLength != 123 {
		t.Errorf("ContentLength = %d, want 123", resp.ContentLength)
	}
	if resp.DurationMillis() != 250 {
		t.Errorf("DurationMillis() = %d, want 250", resp.DurationMillis())
	}
	if resp.RequestID != "test-request-123" {
		t.Errorf("RequestID = %s, want 'test-request-123'", resp.RequestID)
	}
}

// BenchmarkIsSuccess benchmarks the IsSuccess method
func BenchmarkIsSuccess(b *testing.B) {
	resp := NewResponse()
	resp.StatusCode = 200

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.IsSuccess()
	}
}

// BenchmarkGetHeader benchmarks the GetHeader method
func BenchmarkGetHeader(b *testing.B) {
	resp := NewResponse()
	resp.Headers = map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": "123",
		"Authorization":  "Bearer token",
		"Accept":         "application/json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.GetHeader("Content-Type")
	}
}

// BenchmarkGetHeaderCaseInsensitive benchmarks case-insensitive header lookup
func BenchmarkGetHeaderCaseInsensitive(b *testing.B) {
	resp := NewResponse()
	resp.Headers = map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": "123",
		"Authorization":  "Bearer token",
		"Accept":         "application/json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.GetHeader("content-type")
	}
}

// BenchmarkIsJSON benchmarks the IsJSON method
func BenchmarkIsJSON(b *testing.B) {
	resp := NewResponse()
	resp.Headers = map[string]string{"Content-Type": "application/json; charset=utf-8"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.IsJSON()
	}
}

// BenchmarkStatusClass benchmarks the StatusClass method
func BenchmarkStatusClass(b *testing.B) {
	resp := NewResponse()
	resp.StatusCode = 404

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resp.StatusClass()
	}
}
