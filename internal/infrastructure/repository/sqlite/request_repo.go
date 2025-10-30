package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/williajm/curly/internal/domain"
)

// Common repository errors
var (
	ErrNotFound = errors.New("not found")
)

// RequestRepository implements repository.RequestRepository using SQLite.
type RequestRepository struct {
	db *sql.DB
}

// NewRequestRepository creates a new SQLite-backed request repository.
func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

// Create persists a new request to the database.
func (r *RequestRepository) Create(ctx context.Context, req *domain.Request) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Serialize headers to JSON
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize headers: %w", err)
	}

	// Serialize query params to JSON
	queryParamsJSON, err := json.Marshal(req.QueryParams)
	if err != nil {
		return fmt.Errorf("failed to serialize query params: %w", err)
	}

	// Serialize auth config to JSON
	authConfigJSON, err := serializeAuthConfig(req.AuthConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize auth config: %w", err)
	}

	// Get auth type (handle nil AuthConfig)
	authType := ""
	if req.AuthConfig != nil {
		authType = req.AuthConfig.Type()
	}

	query := `
		INSERT INTO requests (id, name, method, url, headers, query_params, body, auth_type, auth_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		req.ID,
		req.Name,
		req.Method,
		req.URL,
		string(headersJSON),
		string(queryParamsJSON),
		req.Body,
		authType,
		string(authConfigJSON),
		req.CreatedAt.Format(time.RFC3339),
		req.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return nil
}

// FindByID retrieves a request by its ID.
func (r *RequestRepository) FindByID(ctx context.Context, id string) (*domain.Request, error) {
	query := `
		SELECT id, name, method, url, headers, query_params, body, auth_type, auth_config, created_at, updated_at
		FROM requests
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var (
		reqID           string
		name            string
		method          string
		url             string
		headersJSON     string
		queryParamsJSON string
		body            string
		authType        string
		authConfigJSON  string
		createdAt       string
		updatedAt       string
	)

	err := row.Scan(&reqID, &name, &method, &url, &headersJSON, &queryParamsJSON, &body, &authType, &authConfigJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan request: %w", err)
	}

	return buildRequest(reqID, name, method, url, headersJSON, queryParamsJSON, body, authType, authConfigJSON, createdAt, updatedAt)
}

// FindAll retrieves all requests ordered by created_at descending.
func (r *RequestRepository) FindAll(ctx context.Context) ([]*domain.Request, error) {
	query := `
		SELECT id, name, method, url, headers, query_params, body, auth_type, auth_config, created_at, updated_at
		FROM requests
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.Request

	for rows.Next() {
		var (
			reqID           string
			name            string
			method          string
			url             string
			headersJSON     string
			queryParamsJSON string
			body            string
			authType        string
			authConfigJSON  string
			createdAt       string
			updatedAt       string
		)

		err := rows.Scan(&reqID, &name, &method, &url, &headersJSON, &queryParamsJSON, &body, &authType, &authConfigJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		req, err := buildRequest(reqID, name, method, url, headersJSON, queryParamsJSON, body, authType, authConfigJSON, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return requests, nil
}

// Update modifies an existing request.
func (r *RequestRepository) Update(ctx context.Context, req *domain.Request) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Update the timestamp
	req.UpdatedAt = time.Now()

	// Serialize headers to JSON
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize headers: %w", err)
	}

	// Serialize query params to JSON
	queryParamsJSON, err := json.Marshal(req.QueryParams)
	if err != nil {
		return fmt.Errorf("failed to serialize query params: %w", err)
	}

	// Serialize auth config to JSON
	authConfigJSON, err := serializeAuthConfig(req.AuthConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize auth config: %w", err)
	}

	// Get auth type (handle nil AuthConfig)
	authType := ""
	if req.AuthConfig != nil {
		authType = req.AuthConfig.Type()
	}

	query := `
		UPDATE requests
		SET name = ?, method = ?, url = ?, headers = ?, query_params = ?, body = ?, auth_type = ?, auth_config = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		req.Name,
		req.Method,
		req.URL,
		string(headersJSON),
		string(queryParamsJSON),
		req.Body,
		authType,
		string(authConfigJSON),
		req.UpdatedAt.Format(time.RFC3339),
		req.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete removes a request from the database.
func (r *RequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM requests WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// buildRequest constructs a domain.Request from database fields.
func buildRequest(id, name, method, url, headersJSON, queryParamsJSON, body, authType, authConfigJSON, createdAt, updatedAt string) (*domain.Request, error) {
	// Parse headers
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return nil, fmt.Errorf("failed to deserialize headers: %w", err)
	}

	// Parse query params
	var queryParams map[string]string
	if err := json.Unmarshal([]byte(queryParamsJSON), &queryParams); err != nil {
		return nil, fmt.Errorf("failed to deserialize query params: %w", err)
	}

	// Parse auth config
	authConfig, err := deserializeAuthConfig(authType, authConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize auth config: %w", err)
	}

	// Parse timestamps
	createdTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	updatedTime, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &domain.Request{
		ID:          id,
		Name:        name,
		Method:      method,
		URL:         url,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        body,
		AuthConfig:  authConfig,
		CreatedAt:   createdTime,
		UpdatedAt:   updatedTime,
	}, nil
}

// serializeAuthConfig converts an AuthConfig to JSON.
func serializeAuthConfig(auth domain.AuthConfig) ([]byte, error) {
	if auth == nil {
		return json.Marshal(map[string]interface{}{})
	}

	switch a := auth.(type) {
	case *domain.NoAuth:
		return json.Marshal(map[string]interface{}{})
	case *domain.BasicAuth:
		return json.Marshal(map[string]interface{}{
			"username": a.Username,
			"password": a.Password,
		})
	case *domain.BearerAuth:
		return json.Marshal(map[string]interface{}{
			"token": a.Token,
		})
	case *domain.APIKeyAuth:
		return json.Marshal(map[string]interface{}{
			"key":      a.Key,
			"value":    a.Value,
			"location": a.Location,
		})
	default:
		return nil, fmt.Errorf("unsupported auth type: %T", auth)
	}
}

// deserializeAuthConfig reconstructs an AuthConfig from JSON.
func deserializeAuthConfig(authType, configJSON string) (domain.AuthConfig, error) {
	switch authType {
	case "", "none":
		return domain.NewNoAuth(), nil
	case "basic":
		var config struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal basic auth: %w", err)
		}
		return domain.NewBasicAuth(config.Username, config.Password), nil
	case "bearer":
		var config struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bearer auth: %w", err)
		}
		return domain.NewBearerAuth(config.Token), nil
	case "apikey":
		var config struct {
			Key      string                  `json:"key"`
			Value    string                  `json:"value"`
			Location domain.APIKeyLocation   `json:"location"`
		}
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal apikey auth: %w", err)
		}
		return domain.NewAPIKeyAuth(config.Key, config.Value, config.Location), nil
	default:
		return nil, fmt.Errorf("unknown auth type: %s", authType)
	}
}
