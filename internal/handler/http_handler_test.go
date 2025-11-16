package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
)

// ==================== Mock Service ====================

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateTeam(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockService) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*domain.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockService) CreatePR(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	args := m.Called(ctx, prID, prName, authorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequestWithReviewers), args.Error(1)
}

func (m *MockService) MergePR(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequestWithReviewers), args.Error(1)
}

func (m *MockService) ReassignReviewer(ctx context.Context, prID, oldUserID uuid.UUID) (*domain.PullRequestWithReviewers, uuid.UUID, error) {
	args := m.Called(ctx, prID, oldUserID)
	if args.Get(0) == nil {
		return nil, uuid.Nil, args.Error(2)
	}
	return args.Get(0).(*domain.PullRequestWithReviewers), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockService) GetUserReviews(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequestShort), args.Error(1)
}

func (m *MockService) DeactivateTeam(ctx context.Context, teamName string) (int, error) {
	args := m.Called(ctx, teamName)
	return args.Int(0), args.Error(1)
}

// ==================== Tests ====================

func TestHealthCheck(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
}

// ==================== Team Tests ====================

func TestCreateTeam_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	userID := uuid.New()
	team := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{
				"user_id":   userID.String(),
				"username":  "Alice",
				"is_active": true,
			},
		},
	}

	mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("*domain.Team")).Return(nil)

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateTeam_InvalidJSON(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("POST", "/team/add", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTeam_TeamExists(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	userID := uuid.New()
	team := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{
				"user_id":   userID.String(),
				"username":  "Alice",
				"is_active": true,
			},
		},
	}

	mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("*domain.Team")).Return(errors.New("TEAM_EXISTS"))

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTeam_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	expectedTeam := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{
				UserID:   uuid.New(),
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	mockService.On("GetTeam", mock.Anything, "backend").Return(expectedTeam, nil)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Team
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "backend", response.TeamName)
	assert.Len(t, response.Members, 1)
	mockService.AssertExpectations(t)
}

func TestGetTeam_NotFound(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	mockService.On("GetTeam", mock.Anything, "backend").Return(nil, errors.New("TEAM_NOT_FOUND"))

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTeam_MissingParameter(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("GET", "/team/get", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== User Tests ====================

func TestSetUserActive_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "admin-secret")
	router := handler.SetupRouter()

	userID := uuid.New()
	requestBody := map[string]interface{}{
		"user_id":   userID.String(),
		"is_active": false,
	}

	expectedUser := &domain.User{
		UserID:   userID,
		Username: "Alice",
		IsActive: false,
		TeamName: "backend",
	}

	mockService.On("SetUserActive", mock.Anything, userID, false).Return(expectedUser, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", "admin-secret")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSetUserActive_Unauthorized(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "admin-secret")
	router := handler.SetupRouter()

	userID := uuid.New()
	requestBody := map[string]interface{}{
		"user_id":   userID.String(),
		"is_active": false,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", "wrong-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetUserActive_InvalidUUID(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "admin-secret")
	router := handler.SetupRouter()

	requestBody := map[string]interface{}{
		"user_id":   "invalid-uuid",
		"is_active": false,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", "admin-secret")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Pull Request Tests ====================

func TestCreatePR_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	prID := uuid.New()
	authorID := uuid.New()
	reviewer1 := uuid.New()
	reviewer2 := uuid.New()

	requestBody := map[string]interface{}{
		"pull_request_id":   prID.String(),
		"pull_request_name": "Add new feature",
		"author_id":         authorID.String(),
	}

	expectedPR := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		PullRequestName:   "Add new feature",
		AuthorID:          authorID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{reviewer1, reviewer2},
	}

	mockService.On("CreatePR", mock.Anything, prID, "Add new feature", authorID).Return(expectedPR, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		PR domain.PullRequestWithReviewers `json:"pr"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, prID, response.PR.PullRequestID)
	assert.Len(t, response.PR.AssignedReviewers, 2)
	mockService.AssertExpectations(t)
}

func TestCreatePR_InvalidJSON(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMergePR_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	prID := uuid.New()
	requestBody := map[string]interface{}{
		"pull_request_id": prID.String(),
	}

	expectedPR := &domain.PullRequestWithReviewers{
		PullRequestID: prID,
		Status:        "merged",
	}

	mockService.On("MergePR", mock.Anything, prID).Return(expectedPR, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		PR domain.PullRequestWithReviewers `json:"pr"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "merged", response.PR.Status)
	mockService.AssertExpectations(t)
}

func TestReassignReviewer_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	prID := uuid.New()
	oldReviewerID := uuid.New()
	newReviewerID := uuid.New()

	requestBody := map[string]interface{}{
		"pull_request_id": prID.String(),
		"old_user_id":     oldReviewerID.String(),
	}

	expectedPR := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{newReviewerID},
	}

	mockService.On("ReassignReviewer", mock.Anything, prID, oldReviewerID).Return(expectedPR, newReviewerID, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetUserReviews_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	userID := uuid.New()
	expectedPRs := []domain.PullRequestShort{
		{
			PullRequestID:   uuid.New(),
			PullRequestName: "PR 1",
			Status:          "open",
		},
		{
			PullRequestID:   uuid.New(),
			PullRequestName: "PR 2",
			Status:          "merged",
		},
	}

	mockService.On("GetUserReviews", mock.Anything, userID).Return(expectedPRs, nil)

	req := httptest.NewRequest("GET", "/users/getReview?user_id="+userID.String(), http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		UserID       uuid.UUID                 `json:"user_id"`
		PullRequests []domain.PullRequestShort `json:"pull_requests"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Len(t, response.PullRequests, 2)
	mockService.AssertExpectations(t)
}

func TestGetUserReviews_InvalidUUID(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("GET", "/users/getReview?user_id=invalid-uuid", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserReviews_MissingParameter(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	req := httptest.NewRequest("GET", "/users/getReview", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStatistics_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService, "test-token")
	router := handler.SetupRouter()

	expectedStats := &domain.Statistics{
		PRStats: domain.PRStats{
			TotalOpen:         5,
			TotalMerged:       10,
			TotalPRs:          15,
			AvgMergeTimeHours: 2.5,
		},
		UserStats: []domain.UserAssignmentStats{
			{
				UserID:            uuid.New(),
				Username:          "Alice",
				TeamName:          "backend",
				TotalAssignments:  10,
				OpenAssignments:   3,
				MergedAssignments: 7,
			},
		},
		TotalUsers:  15,
		TotalTeams:  3,
		ActiveUsers: 12,
	}

	mockService.On("GetStatistics", mock.Anything).Return(expectedStats, nil)

	req := httptest.NewRequest("GET", "/stats", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Statistics
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, 15, response.PRStats.TotalPRs)
	assert.Equal(t, 15, response.TotalUsers)
	mockService.AssertExpectations(t)
}

func (m *MockService) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Statistics), args.Error(1)
}
