package app

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/williajm/curly/internal/domain"
)

func TestNewAuthService(t *testing.T) {
	logger := slog.Default()

	service := NewAuthService(logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestNewAuthService_NilLogger(t *testing.T) {
	service := NewAuthService(nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestCreateAuth_NoAuth(t *testing.T) {
	service := NewAuthService(slog.Default())

	tests := []struct {
		name     string
		authType string
	}{
		{"empty string", ""},
		{"none", "none"},
		{"NONE uppercase", "NONE"},
		{"None mixed case", "None"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := service.CreateAuth(tt.authType, nil)

			assert.NoError(t, err)
			assert.NotNil(t, auth)
			assert.Equal(t, "none", auth.Type())
		})
	}
}

func TestCreateAuth_BasicAuth_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}

	auth, err := service.CreateAuth("basic", credentials)

	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "basic", auth.Type())

	// Verify it's actually a BasicAuth
	basicAuth, ok := auth.(*domain.BasicAuth)
	assert.True(t, ok)
	assert.Equal(t, "testuser", basicAuth.Username)
	assert.Equal(t, "testpass", basicAuth.Password)
}

func TestCreateAuth_BasicAuth_MissingUsername(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"password": "testpass",
	}

	auth, err := service.CreateAuth("basic", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "username")
}

func TestCreateAuth_BasicAuth_MissingPassword(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"username": "testuser",
	}

	auth, err := service.CreateAuth("basic", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "password")
}

func TestCreateAuth_BasicAuth_EmptyUsername(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"username": "   ",
		"password": "testpass",
	}

	auth, err := service.CreateAuth("basic", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "invalid basic auth credentials")
}

func TestCreateAuth_BearerAuth_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"token": "my-secret-token",
	}

	auth, err := service.CreateAuth("bearer", credentials)

	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "bearer", auth.Type())

	// Verify it's actually a BearerAuth
	bearerAuth, ok := auth.(*domain.BearerAuth)
	assert.True(t, ok)
	assert.Equal(t, "my-secret-token", bearerAuth.Token)
}

func TestCreateAuth_BearerAuth_MissingToken(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{}

	auth, err := service.CreateAuth("bearer", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "token")
}

func TestCreateAuth_BearerAuth_EmptyToken(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"token": "   ",
	}

	auth, err := service.CreateAuth("bearer", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "invalid bearer token")
}

func TestCreateAuth_APIKeyAuth_Header_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"key":      "X-API-Key",
		"value":    "my-api-key",
		"location": "header",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "apikey", auth.Type())

	// Verify it's actually an APIKeyAuth
	apiKeyAuth, ok := auth.(*domain.APIKeyAuth)
	assert.True(t, ok)
	assert.Equal(t, "X-API-Key", apiKeyAuth.Key)
	assert.Equal(t, "my-api-key", apiKeyAuth.Value)
	assert.Equal(t, domain.APIKeyLocationHeader, apiKeyAuth.Location)
}

func TestCreateAuth_APIKeyAuth_Query_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"key":      "api_key",
		"value":    "my-api-key",
		"location": "query",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "apikey", auth.Type())

	// Verify it's actually an APIKeyAuth
	apiKeyAuth, ok := auth.(*domain.APIKeyAuth)
	assert.True(t, ok)
	assert.Equal(t, "api_key", apiKeyAuth.Key)
	assert.Equal(t, "my-api-key", apiKeyAuth.Value)
	assert.Equal(t, domain.APIKeyLocationQuery, apiKeyAuth.Location)
}

func TestCreateAuth_APIKeyAuth_MissingKey(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"value":    "my-api-key",
		"location": "header",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "key")
}

func TestCreateAuth_APIKeyAuth_MissingValue(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"key":      "X-API-Key",
		"location": "header",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "value")
}

func TestCreateAuth_APIKeyAuth_MissingLocation(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"key":   "X-API-Key",
		"value": "my-api-key",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "location")
}

func TestCreateAuth_APIKeyAuth_InvalidLocation(t *testing.T) {
	service := NewAuthService(slog.Default())

	credentials := map[string]string{
		"key":      "X-API-Key",
		"value":    "my-api-key",
		"location": "body",
	}

	auth, err := service.CreateAuth("apikey", credentials)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "invalid api key location")
}

func TestCreateAuth_UnsupportedType(t *testing.T) {
	service := NewAuthService(slog.Default())

	auth, err := service.CreateAuth("oauth", nil)

	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "unsupported auth type")
}

func TestValidateAuth_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	tests := []struct {
		name string
		auth domain.AuthConfig
	}{
		{
			"NoAuth",
			domain.NewNoAuth(),
		},
		{
			"BasicAuth",
			domain.NewBasicAuth("user", "pass"),
		},
		{
			"BearerAuth",
			domain.NewBearerAuth("token"),
		},
		{
			"APIKeyAuth Header",
			domain.NewAPIKeyAuth("X-API-Key", "key", domain.APIKeyLocationHeader),
		},
		{
			"APIKeyAuth Query",
			domain.NewAPIKeyAuth("api_key", "key", domain.APIKeyLocationQuery),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAuth(tt.auth)
			assert.NoError(t, err)
		})
	}
}

func TestValidateAuth_NilAuth(t *testing.T) {
	service := NewAuthService(slog.Default())

	err := service.ValidateAuth(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth config cannot be nil")
}

func TestValidateAuth_InvalidAuth(t *testing.T) {
	service := NewAuthService(slog.Default())

	tests := []struct {
		name string
		auth domain.AuthConfig
	}{
		{
			"BasicAuth empty username",
			domain.NewBasicAuth("", "pass"),
		},
		{
			"BasicAuth empty password",
			domain.NewBasicAuth("user", ""),
		},
		{
			"BearerAuth empty token",
			domain.NewBearerAuth(""),
		},
		{
			"APIKeyAuth empty key",
			domain.NewAPIKeyAuth("", "value", domain.APIKeyLocationHeader),
		},
		{
			"APIKeyAuth empty value",
			domain.NewAPIKeyAuth("key", "", domain.APIKeyLocationHeader),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAuth(tt.auth)
			assert.Error(t, err)
		})
	}
}

func TestApplyAuth_Success(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
	auth := domain.NewBasicAuth("user", "pass")

	err := service.ApplyAuth(auth, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, req.Header.Get("Authorization"))
	assert.Contains(t, req.Header.Get("Authorization"), "Basic")
}

func TestApplyAuth_NilAuth(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)

	err := service.ApplyAuth(nil, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth config cannot be nil")
}

func TestApplyAuth_NilRequest(t *testing.T) {
	service := NewAuthService(slog.Default())

	auth := domain.NewBasicAuth("user", "pass")

	err := service.ApplyAuth(auth, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http request cannot be nil")
}

func TestApplyAuth_InvalidAuth(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
	auth := domain.NewBasicAuth("", "pass") // Invalid: empty username

	err := service.ApplyAuth(auth, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth config")
}

func TestApplyAuth_BearerToken(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
	auth := domain.NewBearerAuth("my-token")

	err := service.ApplyAuth(auth, req)

	assert.NoError(t, err)
	assert.Equal(t, "Bearer my-token", req.Header.Get("Authorization"))
}

func TestApplyAuth_APIKeyHeader(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
	auth := domain.NewAPIKeyAuth("X-API-Key", "my-key", domain.APIKeyLocationHeader)

	err := service.ApplyAuth(auth, req)

	assert.NoError(t, err)
	assert.Equal(t, "my-key", req.Header.Get("X-API-Key"))
}

func TestApplyAuth_APIKeyQuery(t *testing.T) {
	service := NewAuthService(slog.Default())

	req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
	auth := domain.NewAPIKeyAuth("api_key", "my-key", domain.APIKeyLocationQuery)

	err := service.ApplyAuth(auth, req)

	assert.NoError(t, err)
	assert.Equal(t, "my-key", req.URL.Query().Get("api_key"))
}

func TestSupportedTypes(t *testing.T) {
	service := NewAuthService(slog.Default())

	types := service.SupportedTypes()

	assert.NotNil(t, types)
	assert.Len(t, types, 4)
	assert.Contains(t, types, "none")
	assert.Contains(t, types, "basic")
	assert.Contains(t, types, "bearer")
	assert.Contains(t, types, "apikey")
}
