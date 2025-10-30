package app

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/williajm/curly/internal/domain"
	"github.com/williajm/curly/internal/infrastructure/repository"
)

// MockRequestRepository is a mock implementation of repository.RequestRepository
type MockRequestRepository struct {
	mock.Mock
}

func (m *MockRequestRepository) Create(ctx context.Context, req *domain.Request) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRequestRepository) FindByID(ctx context.Context, id string) (*domain.Request, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Request), args.Error(1)
}

func (m *MockRequestRepository) FindAll(ctx context.Context) ([]*domain.Request, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Request), args.Error(1)
}

func (m *MockRequestRepository) Update(ctx context.Context, req *domain.Request) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRequestRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockHTTPClient is a mock implementation of http.Client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Execute(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Response), args.Error(1)
}

// MockHistoryRepository is a mock implementation of repository.HistoryRepository
type MockHistoryRepository struct {
	mock.Mock
}

func (m *MockHistoryRepository) Save(ctx context.Context, entry *repository.HistoryEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockHistoryRepository) FindByID(ctx context.Context, id string) (*repository.HistoryEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.HistoryEntry), args.Error(1)
}

func (m *MockHistoryRepository) FindAll(ctx context.Context, limit int) ([]*repository.HistoryEntry, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.HistoryEntry), args.Error(1)
}

func (m *MockHistoryRepository) FindByRequestID(ctx context.Context, requestID string, limit int) ([]*repository.HistoryEntry, error) {
	args := m.Called(ctx, requestID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.HistoryEntry), args.Error(1)
}

func (m *MockHistoryRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHistoryRepository) DeleteOlderThan(ctx context.Context, timestamp string) (int64, error) {
	args := m.Called(ctx, timestamp)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewRequestService(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
	assert.Equal(t, httpClient, service.httpClient)
	assert.Equal(t, historyRepo, service.historyRepo)
	assert.NotNil(t, service.logger)
}

func TestNewRequestService_PanicsOnNilRepo(t *testing.T) {
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	assert.Panics(t, func() {
		NewRequestService(nil, httpClient, historyRepo, logger)
	})
}

func TestNewRequestService_PanicsOnNilHTTPClient(t *testing.T) {
	repo := new(MockRequestRepository)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	assert.Panics(t, func() {
		NewRequestService(repo, nil, historyRepo, logger)
	})
}

func TestNewRequestService_PanicsOnNilHistoryRepo(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	logger := slog.Default()

	assert.Panics(t, func() {
		NewRequestService(repo, httpClient, nil, logger)
	})
}

func TestCreateRequest_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"

	createdReq, err := service.CreateRequest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, createdReq)
	assert.NotEmpty(t, createdReq.ID)
	assert.False(t, createdReq.CreatedAt.IsZero())
	assert.False(t, createdReq.UpdatedAt.IsZero())
}

func TestCreateRequest_InvalidRequest(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	// Request with invalid URL
	req := domain.NewRequestWithMethodAndURL("GET", "not-a-url")

	createdReq, err := service.CreateRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, createdReq)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestExecuteRequest_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")

	expectedResp := &domain.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       "test response",
		Duration:   100 * time.Millisecond,
	}

	httpClient.On("Execute", mock.Anything, req).Return(expectedResp, nil)

	resp, err := service.ExecuteRequest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "test response", resp.Body)

	httpClient.AssertExpectations(t)
}

func TestExecuteRequest_ValidationFails(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	// Request with invalid URL
	req := domain.NewRequestWithMethodAndURL("GET", "invalid-url")

	resp, err := service.ExecuteRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestExecuteRequest_HTTPClientError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")

	httpClient.On("Execute", mock.Anything, req).Return(nil, errors.New("network error"))

	resp, err := service.ExecuteRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to execute request")

	httpClient.AssertExpectations(t)
}

func TestSaveRequest_Create(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"

	// Request doesn't exist, so FindByID returns error
	repo.On("FindByID", mock.Anything, req.ID).Return(nil, errors.New("not found"))
	repo.On("Create", mock.Anything, req).Return(nil)

	err := service.SaveRequest(context.Background(), req)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSaveRequest_Update(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"

	existingReq := req.Clone()

	// Request exists, so FindByID returns it
	repo.On("FindByID", mock.Anything, req.ID).Return(existingReq, nil)
	repo.On("Update", mock.Anything, req).Return(nil)

	err := service.SaveRequest(context.Background(), req)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSaveRequest_InvalidRequest(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	// Invalid request
	req := domain.NewRequestWithMethodAndURL("GET", "invalid-url")

	err := service.SaveRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestLoadRequest_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	expectedReq := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	expectedReq.Name = "Test Request"

	repo.On("FindByID", mock.Anything, expectedReq.ID).Return(expectedReq, nil)

	req, err := service.LoadRequest(context.Background(), expectedReq.ID)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, expectedReq.ID, req.ID)
	assert.Equal(t, expectedReq.Name, req.Name)

	repo.AssertExpectations(t)
}

func TestLoadRequest_NotFound(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	repo.On("FindByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	req, err := service.LoadRequest(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "failed to load request")

	repo.AssertExpectations(t)
}

func TestListRequests_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	expectedReqs := []*domain.Request{
		domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test1"),
		domain.NewRequestWithMethodAndURL("POST", "https://api.example.com/test2"),
	}

	repo.On("FindAll", mock.Anything).Return(expectedReqs, nil)

	reqs, err := service.ListRequests(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, reqs)
	assert.Len(t, reqs, 2)

	repo.AssertExpectations(t)
}

func TestListRequests_Error(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	repo.On("FindAll", mock.Anything).Return(nil, errors.New("database error"))

	reqs, err := service.ListRequests(context.Background())

	assert.Error(t, err)
	assert.Nil(t, reqs)
	assert.Contains(t, err.Error(), "failed to list requests")

	repo.AssertExpectations(t)
}

func TestDeleteRequest_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	requestID := "test-id"

	repo.On("Delete", mock.Anything, requestID).Return(nil)

	err := service.DeleteRequest(context.Background(), requestID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteRequest_Error(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	requestID := "test-id"

	repo.On("Delete", mock.Anything, requestID).Return(errors.New("not found"))

	err := service.DeleteRequest(context.Background(), requestID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete request")

	repo.AssertExpectations(t)
}

func TestExecuteAndSave_Success(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")

	expectedResp := &domain.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       "test response",
		Duration:   100 * time.Millisecond,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}

	httpClient.On("Execute", mock.Anything, req).Return(expectedResp, nil)
	historyRepo.On("Save", mock.Anything, mock.AnythingOfType("*repository.HistoryEntry")).Return(nil)

	resp, err := service.ExecuteAndSave(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	httpClient.AssertExpectations(t)
	historyRepo.AssertExpectations(t)
}

func TestExecuteAndSave_HTTPError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")

	httpClient.On("Execute", mock.Anything, req).Return(nil, errors.New("network error"))
	// History should still be saved even on error
	historyRepo.On("Save", mock.Anything, mock.AnythingOfType("*repository.HistoryEntry")).Return(nil)

	resp, err := service.ExecuteAndSave(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to execute request")

	httpClient.AssertExpectations(t)
	historyRepo.AssertExpectations(t)
}

func TestExecuteAndSave_HistorySaveError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")

	expectedResp := &domain.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       "test response",
		Duration:   100 * time.Millisecond,
		Headers:    map[string]string{},
	}

	httpClient.On("Execute", mock.Anything, req).Return(expectedResp, nil)
	historyRepo.On("Save", mock.Anything, mock.AnythingOfType("*repository.HistoryEntry")).Return(errors.New("db error"))

	// Should still succeed even if history save fails
	resp, err := service.ExecuteAndSave(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	httpClient.AssertExpectations(t)
	historyRepo.AssertExpectations(t)
}

func TestExecuteAndSave_ValidationError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	// Invalid request
	req := domain.NewRequestWithMethodAndURL("GET", "invalid-url")

	resp, err := service.ExecuteAndSave(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestSaveRequest_CreateError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"

	// Request doesn't exist
	repo.On("FindByID", mock.Anything, req.ID).Return(nil, errors.New("not found"))
	// But create fails
	repo.On("Create", mock.Anything, req).Return(errors.New("database error"))

	err := service.SaveRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save request")
	repo.AssertExpectations(t)
}

func TestSaveRequest_UpdateError(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"

	existingReq := req.Clone()

	// Request exists
	repo.On("FindByID", mock.Anything, req.ID).Return(existingReq, nil)
	// But update fails
	repo.On("Update", mock.Anything, req).Return(errors.New("database error"))

	err := service.SaveRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update request")
	repo.AssertExpectations(t)
}

func TestCreateRequest_WithExistingID(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.ID = "existing-id"
	req.Name = "Test Request"

	createdReq, err := service.CreateRequest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, createdReq)
	assert.Equal(t, "existing-id", createdReq.ID, "should preserve existing ID")
}

func TestCreateRequest_WithExistingTimestamp(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)
	logger := slog.Default()

	service := NewRequestService(repo, httpClient, historyRepo, logger)

	existingTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	req := domain.NewRequestWithMethodAndURL("GET", "https://api.example.com/test")
	req.Name = "Test Request"
	req.CreatedAt = existingTime

	createdReq, err := service.CreateRequest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, createdReq)
	assert.Equal(t, existingTime, createdReq.CreatedAt, "should preserve existing CreatedAt")
	assert.False(t, createdReq.UpdatedAt.IsZero(), "should set UpdatedAt")
}

func TestNewRequestService_NilLogger(t *testing.T) {
	repo := new(MockRequestRepository)
	httpClient := new(MockHTTPClient)
	historyRepo := new(MockHistoryRepository)

	// Pass nil logger - should default to slog.Default()
	service := NewRequestService(repo, httpClient, historyRepo, nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger, "should have default logger")
}
