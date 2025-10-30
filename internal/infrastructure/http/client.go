// Package http provides HTTP client functionality for executing API requests.
//
// It wraps the standard library net/http with custom configuration options,.
// request/response conversion, timing metrics, and error handling tailored.
// for the curly application.
package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/williajm/curly/internal/domain"
)

// Client executes HTTP requests with configurable timeout and TLS settings.
// It converts domain.Request to HTTP requests, executes them, and returns.
// domain.Response with timing and metadata.
type Client interface {
	// Execute sends the HTTP request and returns the response with timing information.
	// The context can be used for cancellation and timeout control.
	Execute(ctx context.Context, req *domain.Request) (*domain.Response, error)
}

// Config holds HTTP client configuration options.
type Config struct {
	// Timeout is the maximum duration for the entire request (including redirects).
	Timeout time.Duration

	// MaxRedirects is the maximum number of redirects to follow.
	// Set to 0 to disable redirects.
	MaxRedirects int

	// InsecureSkipTLS disables TLS certificate verification.
	// WARNING: This should only be used for testing/development.
	InsecureSkipTLS bool

	// FollowRedirects determines whether to automatically follow redirects.
	FollowRedirects bool

	// DialTimeout is the maximum time to wait for a TCP connection.
	DialTimeout time.Duration

	// TLSHandshakeTimeout is the maximum time to wait for TLS handshake.
	TLSHandshakeTimeout time.Duration

	// ResponseHeaderTimeout is the maximum time to wait for response headers.
	ResponseHeaderTimeout time.Duration

	// KeepAlive specifies the keep-alive period for active network connections.
	KeepAlive time.Duration

	// IdleConnTimeout is the maximum time an idle connection remains open.
	IdleConnTimeout time.Duration
}

// DefaultConfig returns a Config with sensible default values.
// These defaults prioritize reliability and safety over maximum performance.
func DefaultConfig() *Config {
	return &Config{
		Timeout:               30 * time.Second,
		MaxRedirects:          10,
		InsecureSkipTLS:       false,
		FollowRedirects:       true,
		DialTimeout:           10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		KeepAlive:             30 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
}

// httpClient is the concrete implementation of the Client interface.
type httpClient struct {
	client *http.Client
	config *Config
}

// NewClient creates a new HTTP client with the provided configuration.
// If config is nil, DefaultConfig() is used.
func NewClient(config *Config) Client {
	if config == nil {
		config = DefaultConfig()
	}

	// Create custom transport with configured timeouts.
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		IdleConnTimeout:       config.IdleConnTimeout,
		// Disable HTTP/2 for now to keep things simple.
		ForceAttemptHTTP2: false,
	}

	// Configure TLS if needed.
	if config.InsecureSkipTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // #nosec G402 -- Intentionally allow insecure TLS for testing self-signed certificates
		}
	}

	// Configure redirect policy.
	checkRedirect := func(_ *http.Request, via []*http.Request) error {
		if !config.FollowRedirects {
			return http.ErrUseLastResponse
		}
		if len(via) >= config.MaxRedirects {
			return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
		}
		return nil
	}

	return &httpClient{
		client: &http.Client{
			Transport:     transport,
			CheckRedirect: checkRedirect,
			Timeout:       config.Timeout,
		},
		config: config,
	}
}

// Execute converts the domain request to an HTTP request, executes it,.
// and converts the HTTP response back to a domain response with timing metrics.
func (c *httpClient) Execute(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// Validate the request before processing.
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build the HTTP request.
	httpReq, err := c.buildHTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// Apply authentication.
	if req.AuthConfig != nil {
		if err := req.AuthConfig.Apply(httpReq); err != nil {
			return nil, fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	// Execute the request and measure timing.
	startTime := time.Now()
	httpResp, err := c.client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		return nil, c.handleRequestError(err, duration, startTime)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	// Convert HTTP response to domain response.
	resp, err := c.buildDomainResponse(httpResp, duration, startTime, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to process response: %w", err)
	}

	return resp, nil
}

// buildHTTPRequest converts a domain.Request to an *http.Request.
func (c *httpClient) buildHTTPRequest(ctx context.Context, req *domain.Request) (*http.Request, error) {
	// Parse and build URL with query parameters.
	requestURL, err := c.buildURL(req)
	if err != nil {
		return nil, err
	}

	// Create request body reader.
	var bodyReader io.Reader
	if req.Body != "" && req.IsBodyAllowed() {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	// Create HTTP request.
	httpReq, err := http.NewRequestWithContext(ctx, strings.ToUpper(req.Method), requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add custom headers.
	for name, value := range req.Headers {
		httpReq.Header.Set(name, value)
	}

	// Set Content-Length for requests with body.
	if req.Body != "" && req.IsBodyAllowed() {
		httpReq.ContentLength = int64(len(req.Body))
	}

	return httpReq, nil
}

// buildURL constructs the full URL with query parameters merged correctly.
func (c *httpClient) buildURL(req *domain.Request) (string, error) {
	// Parse the base URL.
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// If there are no query parameters, return as-is.
	if len(req.QueryParams) == 0 {
		return parsedURL.String(), nil
	}

	// Merge query parameters from URL and request.
	query := parsedURL.Query()
	for key, value := range req.QueryParams {
		query.Set(key, value)
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// buildDomainResponse converts an *http.Response to a domain.Response.
func (c *httpClient) buildDomainResponse(httpResp *http.Response, duration time.Duration, timestamp time.Time, requestID string) (*domain.Response, error) {
	// Read response body.
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert headers to map.
	headers := make(map[string]string)
	for name, values := range httpResp.Header {
		// Join multiple values with comma (per HTTP spec).
		if len(values) > 0 {
			headers[name] = strings.Join(values, ", ")
		}
	}

	// Build domain response.
	resp := &domain.Response{
		StatusCode:    httpResp.StatusCode,
		Status:        httpResp.Status,
		Headers:       headers,
		Body:          string(bodyBytes),
		ContentLength: httpResp.ContentLength,
		Duration:      duration,
		Timestamp:     timestamp,
		RequestID:     requestID,
	}

	// If ContentLength is -1 (unknown), use actual body length.
	if resp.ContentLength == -1 {
		resp.ContentLength = int64(len(bodyBytes))
	}

	return resp, nil
}

// handleRequestError converts HTTP client errors to user-friendly error messages.
func (c *httpClient) handleRequestError(err error, duration time.Duration, _ time.Time) error {
	// Check for context cancellation.
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("request canceled: %w", err)
	}

	// Check for timeout.
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("request timeout after %v: %w", duration, err)
	}

	// Check for URL errors.
	var urlErr *url.Error
	if ok := errors.As(err, &urlErr); ok {
		// Timeout errors.
		if urlErr.Timeout() {
			return fmt.Errorf("request timeout after %v: %w", duration, urlErr)
		}

		// Temporary errors (can be retried).
		if urlErr.Temporary() {
			return fmt.Errorf("temporary network error: %w", urlErr)
		}

		// DNS errors.
		var dnsErr *net.DNSError
		if ok := errors.As(urlErr.Err, &dnsErr); ok {
			return fmt.Errorf("DNS lookup failed for %s: %w", dnsErr.Name, dnsErr)
		}

		// Connection errors.
		var opErr *net.OpError
		if ok := errors.As(urlErr.Err, &opErr); ok {
			return fmt.Errorf("connection failed: %s: %w", opErr.Op, opErr)
		}
	}

	// Generic network error.
	return fmt.Errorf("request failed: %w", err)
}
