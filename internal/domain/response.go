package domain

import (
	"time"
)

// Response represents the result of executing an HTTP request.
// It captures the status, headers, body, timing, and other metadata.
type Response struct {
	// StatusCode is the HTTP status code (e.g., 200, 404, 500).
	StatusCode int

	// Status is the HTTP status string (e.g., "200 OK", "404 Not Found").
	Status string

	// Headers are the response headers received from the server.
	// Keys are header names, values are header values.
	Headers map[string]string

	// Body is the response body content as a string.
	Body string

	// ContentLength is the size of the response body in bytes.
	ContentLength int64

	// Duration is how long the request took to complete.
	Duration time.Duration

	// Timestamp is when the response was received.
	Timestamp time.Time

	// RequestID is the ID of the request that generated this response.
	// This links the response back to its originating request.
	RequestID string
}

// NewResponse creates a new Response with default values.
func NewResponse() *Response {
	return &Response{
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}
}

// IsSuccess returns true if the response status code indicates success (2xx).
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsRedirect returns true if the response status code indicates a redirect (3xx).
func (r *Response) IsRedirect() bool {
	return r.StatusCode >= 300 && r.StatusCode < 400
}

// IsClientError returns true if the response status code indicates a client error (4xx).
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the response status code indicates a server error (5xx).
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// IsError returns true if the response status code indicates any error (4xx or 5xx).
func (r *Response) IsError() bool {
	return r.IsClientError() || r.IsServerError()
}

// IsInformational returns true if the response status code is informational (1xx).
func (r *Response) IsInformational() bool {
	return r.StatusCode >= 100 && r.StatusCode < 200
}

// StatusClass returns a string representing the class of the status code.
// Returns one of: "1xx Informational", "2xx Success", "3xx Redirect",.
// "4xx Client Error", "5xx Server Error", or "Unknown".
func (r *Response) StatusClass() string {
	switch {
	case r.IsInformational():
		return "1xx Informational"
	case r.IsSuccess():
		return "2xx Success"
	case r.IsRedirect():
		return "3xx Redirect"
	case r.IsClientError():
		return "4xx Client Error"
	case r.IsServerError():
		return "5xx Server Error"
	default:
		return "Unknown"
	}
}

// GetHeader returns the value of a response header.
// Header names are case-insensitive. Returns empty string if not found.
func (r *Response) GetHeader(name string) string {
	// Try exact match first.
	if value, ok := r.Headers[name]; ok {
		return value
	}

	// Try case-insensitive match.
	for key, value := range r.Headers {
		if equalsFold(key, name) {
			return value
		}
	}

	return ""
}

// HasHeader returns true if the response contains the specified header.
// Header names are case-insensitive.
func (r *Response) HasHeader(name string) bool {
	return r.GetHeader(name) != ""
}

// ContentType returns the Content-Type header value, or empty string if not present.
func (r *Response) ContentType() string {
	return r.GetHeader("Content-Type")
}

// IsJSON returns true if the response Content-Type indicates JSON.
func (r *Response) IsJSON() bool {
	contentType := r.ContentType()
	return contains(contentType, "application/json") || contains(contentType, "application/vnd.api+json")
}

// IsXML returns true if the response Content-Type indicates XML.
func (r *Response) IsXML() bool {
	contentType := r.ContentType()
	return contains(contentType, "application/xml") || contains(contentType, "text/xml")
}

// IsHTML returns true if the response Content-Type indicates HTML.
func (r *Response) IsHTML() bool {
	contentType := r.ContentType()
	return contains(contentType, "text/html")
}

// IsText returns true if the response Content-Type indicates plain text.
func (r *Response) IsText() bool {
	contentType := r.ContentType()
	return contains(contentType, "text/plain")
}

// DurationMillis returns the response duration in milliseconds.
func (r *Response) DurationMillis() int64 {
	return r.Duration.Milliseconds()
}

// DurationSeconds returns the response duration in seconds.
func (r *Response) DurationSeconds() float64 {
	return r.Duration.Seconds()
}

// Helper functions.

// equalsFold compares two strings case-insensitively.
func equalsFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		// Convert to lowercase.
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// contains checks if substr is contained in s (case-insensitive).
func contains(s, substr string) bool {
	// Simple case-insensitive contains check.
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Convert both strings to lowercase for comparison.
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		sLower[i] = c
	}

	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		substrLower[i] = c
	}

	// Search for substring.
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		match := true
		for j := 0; j < len(substrLower); j++ {
			if sLower[i+j] != substrLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
