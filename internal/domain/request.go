package domain

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTP method constants.
const (
	// MethodGet represents the HTTP GET method.
	MethodGet = "GET"

	// MethodPost represents the HTTP POST method.
	MethodPost = "POST"

	// MethodPut represents the HTTP PUT method.
	MethodPut = "PUT"

	// MethodPatch represents the HTTP PATCH method.
	MethodPatch = "PATCH"

	// MethodDelete represents the HTTP DELETE method.
	MethodDelete = "DELETE"

	// MethodHead represents the HTTP HEAD method.
	MethodHead = "HEAD"

	// MethodOptions represents the HTTP OPTIONS method.
	MethodOptions = "OPTIONS"
)

// SupportedMethods lists all HTTP methods supported by curly.
var SupportedMethods = []string{
	MethodGet,
	MethodPost,
	MethodPut,
	MethodPatch,
	MethodDelete,
	MethodHead,
	MethodOptions,
}

// Request represents an HTTP request configuration.
// It contains all the information needed to construct and execute an HTTP request.
type Request struct {
	// ID is a unique identifier for this request.
	ID string

	// Name is a human-readable name for this request.
	Name string

	// Method is the HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS).
	Method string

	// URL is the full URL for the request.
	URL string

	// Headers are custom HTTP headers to include with the request.
	// Keys are header names, values are header values.
	Headers map[string]string

	// QueryParams are query parameters to append to the URL.
	// Keys are parameter names, values are parameter values.
	QueryParams map[string]string

	// Body is the request body content.
	// For JSON requests, this should be the JSON string.
	Body string

	// AuthConfig is the authentication configuration for this request.
	// If nil, no authentication is applied.
	AuthConfig AuthConfig

	// CreatedAt is the timestamp when this request was created.
	CreatedAt time.Time

	// UpdatedAt is the timestamp when this request was last modified.
	UpdatedAt time.Time
}

// NewRequest creates a new Request with default values.
// It generates a unique ID and sets timestamps.
func NewRequest() *Request {
	now := time.Now()
	return &Request{
		ID:          uuid.New().String(),
		Method:      MethodGet,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		AuthConfig:  NewNoAuth(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewRequestWithMethodAndURL creates a new Request with the specified method and URL.
func NewRequestWithMethodAndURL(method, requestURL string) *Request {
	req := NewRequest()
	req.Method = method
	req.URL = requestURL
	return req
}

// Validate checks if the request configuration is valid.
// It validates the method, URL, headers, and authentication.
func (r *Request) Validate() error {
	// Validate method.
	if err := r.ValidateMethod(); err != nil {
		return err
	}

	// Validate URL.
	if err := r.ValidateURL(); err != nil {
		return err
	}

	// Validate headers.
	if err := r.ValidateHeaders(); err != nil {
		return err
	}

	// Validate query parameters.
	if err := r.ValidateQueryParams(); err != nil {
		return err
	}

	// Validate auth config if present.
	if r.AuthConfig != nil {
		if err := r.AuthConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateMethod checks if the HTTP method is supported.
func (r *Request) ValidateMethod() error {
	if strings.TrimSpace(r.Method) == "" {
		return ErrEmptyMethod
	}

	method := strings.ToUpper(r.Method)
	for _, supported := range SupportedMethods {
		if method == supported {
			return nil
		}
	}

	return ErrInvalidMethod
}

// ValidateURL checks if the URL is valid and has a supported scheme.
func (r *Request) ValidateURL() error {
	if strings.TrimSpace(r.URL) == "" {
		return ErrEmptyURL
	}

	parsed, err := url.Parse(r.URL)
	if err != nil {
		return ErrInvalidURL
	}

	// Ensure the URL has a scheme (http or https).
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}

	// Ensure the URL has a host.
	if parsed.Host == "" {
		return ErrInvalidURL
	}

	return nil
}

// ValidateHeaders checks if all header names are valid.
// Header names cannot be empty or contain invalid characters.
func (r *Request) ValidateHeaders() error {
	for name := range r.Headers {
		if strings.TrimSpace(name) == "" {
			return ErrInvalidHeaderName
		}
		// Basic validation: header names should not contain colons, spaces, or newlines.
		if strings.ContainsAny(name, ":\n\r") {
			return ErrInvalidHeaderName
		}
	}
	return nil
}

// ValidateQueryParams checks if all query parameter names are valid.
func (r *Request) ValidateQueryParams() error {
	for name := range r.QueryParams {
		if strings.TrimSpace(name) == "" {
			return ErrInvalidQueryParam
		}
	}
	return nil
}

// SetHeader sets a header value. If the value is empty, the header is removed.
func (r *Request) SetHeader(name, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	if value == "" {
		delete(r.Headers, name)
	} else {
		r.Headers[name] = value
	}
	r.UpdatedAt = time.Now()
}

// SetQueryParam sets a query parameter. If the value is empty, the parameter is removed.
func (r *Request) SetQueryParam(name, value string) {
	if r.QueryParams == nil {
		r.QueryParams = make(map[string]string)
	}
	if value == "" {
		delete(r.QueryParams, name)
	} else {
		r.QueryParams[name] = value
	}
	r.UpdatedAt = time.Now()
}

// SetAuth sets the authentication configuration for this request.
func (r *Request) SetAuth(auth AuthConfig) {
	r.AuthConfig = auth
	r.UpdatedAt = time.Now()
}

// Clone creates a deep copy of the request.
// Useful for modifying a request without affecting the original.
func (r *Request) Clone() *Request {
	clone := &Request{
		ID:          r.ID,
		Name:        r.Name,
		Method:      r.Method,
		URL:         r.URL,
		Body:        r.Body,
		AuthConfig:  r.AuthConfig,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}

	// Deep copy maps.
	for k, v := range r.Headers {
		clone.Headers[k] = v
	}
	for k, v := range r.QueryParams {
		clone.QueryParams[k] = v
	}

	return clone
}

// IsBodyAllowed returns true if the HTTP method allows a request body.
func (r *Request) IsBodyAllowed() bool {
	method := strings.ToUpper(r.Method)
	switch method {
	case MethodPost, MethodPut, MethodPatch:
		return true
	default:
		return false
	}
}
