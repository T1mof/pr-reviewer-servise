package service

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
)

// ========================================
// Mock Repository
// ========================================

type MockRepository struct {
	mock.Mock
}

// TeamRepository methods.
func (m *MockRepository) CreateTeam(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockRepository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

// UserRepository methods.
func (m *MockRepository) UpsertUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockRepository) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockRepository) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

// PullRequestRepository methods.
func (m *MockRepository) CreatePR(ctx context.Context, pr *domain.PullRequest, reviewers []uuid.UUID) error {
	args := m.Called(ctx, pr, reviewers)
	return args.Error(0)
}

func (m *MockRepository) GetPRByID(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequestWithReviewers), args.Error(1)
}

func (m *MockRepository) PRExists(ctx context.Context, prID uuid.UUID) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) UpdatePRStatus(ctx context.Context, prID uuid.UUID, status string, mergedAt *time.Time) error {
	args := m.Called(ctx, prID, status, mergedAt)
	return args.Error(0)
}

func (m *MockRepository) GetReviewersByPR(ctx context.Context, prID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID uuid.UUID) error {
	args := m.Called(ctx, prID, oldUserID, newUserID)
	return args.Error(0)
}

func (m *MockRepository) GetPRsByReviewer(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequestShort), args.Error(1)
}

// ========================================
// Tests
// ========================================

func TestCreateTeam_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{
				UserID:   uuid.New(),
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	mockRepo.On("TeamExists", mock.Anything, "backend").Return(false, nil)
	mockRepo.On("CreateTeam", mock.Anything, team).Return(nil)

	err := service.CreateTeam(context.Background(), team)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{
				UserID:   uuid.New(),
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	mockRepo.On("TeamExists", mock.Anything, "backend").Return(true, nil)

	err := service.CreateTeam(context.Background(), team)

	assert.EqualError(t, err, "TEAM_EXISTS")
	mockRepo.AssertExpectations(t)
}

func TestCreateTeam_EmptyTeamName(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team := &domain.Team{
		TeamName: "",
		Members: []domain.TeamMember{
			{
				UserID:   uuid.New(),
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	err := service.CreateTeam(context.Background(), team)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestCreateTeam_EmptyMembers(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team := &domain.Team{
		TeamName: "backend",
		Members:  []domain.TeamMember{},
	}

	err := service.CreateTeam(context.Background(), team)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestGetTeam_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

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

	mockRepo.On("GetTeamByName", mock.Anything, "backend").Return(expectedTeam, nil)

	team, err := service.GetTeam(context.Background(), "backend")

	assert.NoError(t, err)
	assert.Equal(t, "backend", team.TeamName)
	assert.Len(t, team.Members, 1)
	mockRepo.AssertExpectations(t)
}

func TestSetUserActive_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	userID := uuid.New()
	expectedUser := &domain.User{
		UserID:   userID,
		Username: "Alice",
		IsActive: false,
		TeamName: "backend",
	}

	mockRepo.On("SetUserActive", mock.Anything, userID, false).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

	user, err := service.SetUserActive(context.Background(), userID, false)

	assert.NoError(t, err)
	assert.Equal(t, false, user.IsActive)
	assert.Equal(t, "Alice", user.Username)
	mockRepo.AssertExpectations(t)
}

func TestSetUserActive_NilUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	user, err := service.SetUserActive(context.Background(), uuid.Nil, true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user_id cannot be nil UUID")
}

func TestSelectReviewers_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &ReviewerService{
		repo:      mockRepo,
		validator: domain.NewValidator(),
		rand:      rand.New(rand.NewSource(1)),
	}

	members := []domain.User{
		{UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Username: "Alice", IsActive: true},
		{UserID: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"), Username: "Bob", IsActive: true},
		{UserID: uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), Username: "Charlie", IsActive: false},
		{UserID: uuid.MustParse("7c9e6679-7425-40de-944b-e07fc1f90ae7"), Username: "Dave", IsActive: true},
	}

	excludeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	reviewers := service.selectReviewers(members, excludeID, 2)

	assert.Len(t, reviewers, 2)
	assert.NotContains(t, reviewers, excludeID)

	charlieID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	assert.NotContains(t, reviewers, charlieID)
}

func TestSelectReviewers_NoActiveCandidates(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &ReviewerService{
		repo:      mockRepo,
		validator: domain.NewValidator(),
		rand:      rand.New(rand.NewSource(1)),
	}

	members := []domain.User{
		{UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Username: "Alice", IsActive: true},
		{UserID: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"), Username: "Bob", IsActive: false},
	}

	excludeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	reviewers := service.selectReviewers(members, excludeID, 2)

	assert.Len(t, reviewers, 0)
}

func TestSelectReviewers_OnlyOneCandidate(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &ReviewerService{
		repo:      mockRepo,
		validator: domain.NewValidator(),
		rand:      rand.New(rand.NewSource(1)),
	}

	members := []domain.User{
		{UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Username: "Alice", IsActive: true},
		{UserID: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"), Username: "Bob", IsActive: true},
	}

	excludeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	reviewers := service.selectReviewers(members, excludeID, 2)

	assert.Len(t, reviewers, 1)
	assert.Contains(t, reviewers, uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"))
}

func TestCreatePR_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reviewer1 := uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	reviewer2 := uuid.MustParse("7c9e6679-7425-40de-944b-e07fc1f90ae7")

	author := &domain.User{
		UserID:   authorID,
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	members := []domain.User{
		{UserID: authorID, Username: "Alice", IsActive: true, TeamName: "backend"},
		{UserID: reviewer1, Username: "Bob", IsActive: true, TeamName: "backend"},
		{UserID: reviewer2, Username: "Dave", IsActive: true, TeamName: "backend"},
	}

	expectedPR := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		PullRequestName:   "Add new feature",
		AuthorID:          authorID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{reviewer1, reviewer2},
	}

	mockRepo.On("PRExists", mock.Anything, prID).Return(false, nil)
	mockRepo.On("GetUserByID", mock.Anything, authorID).Return(author, nil)
	mockRepo.On("GetTeamMembers", mock.Anything, "backend").Return(members, nil)
	mockRepo.On("CreatePR", mock.Anything, mock.AnythingOfType("*domain.PullRequest"), mock.AnythingOfType("[]uuid.UUID")).Return(nil)
	mockRepo.On("GetPRByID", mock.Anything, prID).Return(expectedPR, nil)

	pr, err := service.CreatePR(context.Background(), prID, "Add new feature", authorID)

	assert.NoError(t, err)
	assert.Equal(t, prID, pr.PullRequestID)
	assert.Equal(t, "open", pr.Status)
	assert.Len(t, pr.AssignedReviewers, 2)
	mockRepo.AssertExpectations(t)
}

func TestMergePR_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	existingPR := &domain.PullRequestWithReviewers{
		PullRequestID:   prID,
		PullRequestName: "Add new feature",
		Status:          "open",
	}

	mergedPR := &domain.PullRequestWithReviewers{
		PullRequestID:   prID,
		PullRequestName: "Add new feature",
		Status:          "merged",
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(existingPR, nil).Once()
	mockRepo.On("UpdatePRStatus", mock.Anything, prID, "merged", mock.AnythingOfType("*time.Time")).Return(nil)
	mockRepo.On("GetPRByID", mock.Anything, prID).Return(mergedPR, nil).Once()

	pr, err := service.MergePR(context.Background(), prID)

	assert.NoError(t, err)
	assert.Equal(t, "merged", pr.Status)
	mockRepo.AssertExpectations(t)
}

func TestMergePR_AlreadyMerged(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	mergedPR := &domain.PullRequestWithReviewers{
		PullRequestID:   prID,
		PullRequestName: "Add new feature",
		Status:          "merged",
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(mergedPR, nil)

	pr, err := service.MergePR(context.Background(), prID)

	assert.NoError(t, err)
	assert.Equal(t, "merged", pr.Status)
	mockRepo.AssertExpectations(t)
}

func TestGetUserReviews_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

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

	mockRepo.On("GetPRsByReviewer", mock.Anything, userID).Return(expectedPRs, nil)

	prs, err := service.GetUserReviews(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, prs, 2)
	mockRepo.AssertExpectations(t)
}

// ========== Error Handling Tests ==========

func TestCreateTeam_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: uuid.New(), Username: "Alice", IsActive: true},
		},
	}

	mockRepo.On("TeamExists", mock.Anything, "backend").Return(false, nil)
	mockRepo.On("CreateTeam", mock.Anything, team).Return(errors.New("database error"))

	err := service.CreateTeam(context.Background(), team)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTeam_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	mockRepo.On("GetTeamByName", mock.Anything, "backend").Return(nil, errors.New("TEAM_NOT_FOUND"))

	team, err := service.GetTeam(context.Background(), "backend")

	assert.Error(t, err)
	assert.Nil(t, team)
	mockRepo.AssertExpectations(t)
}

func TestGetTeam_EmptyName(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	team, err := service.GetTeam(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Contains(t, err.Error(), "team_name cannot be empty")
}

func TestSetUserActive_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	userID := uuid.New()
	mockRepo.On("SetUserActive", mock.Anything, userID, true).Return(errors.New("USER_NOT_FOUND"))

	user, err := service.SetUserActive(context.Background(), userID, true)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// ========== CreatePR Edge Cases ==========

func TestCreatePR_AlreadyExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.New()

	mockRepo.On("PRExists", mock.Anything, prID).Return(true, nil)

	pr, err := service.CreatePR(context.Background(), prID, "Feature", authorID)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "PR_EXISTS")
	mockRepo.AssertExpectations(t)
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.New()

	mockRepo.On("PRExists", mock.Anything, prID).Return(false, nil)
	mockRepo.On("GetUserByID", mock.Anything, authorID).Return(nil, errors.New("USER_NOT_FOUND"))

	pr, err := service.CreatePR(context.Background(), prID, "Feature", authorID)

	assert.Error(t, err)
	assert.Nil(t, pr)
	mockRepo.AssertExpectations(t)
}

func TestCreatePR_NotEnoughReviewers(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.New()

	author := &domain.User{
		UserID:   authorID,
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	members := []domain.User{
		{UserID: authorID, Username: "Alice", IsActive: true, TeamName: "backend"},
	}

	mockRepo.On("PRExists", mock.Anything, prID).Return(false, nil)
	mockRepo.On("GetUserByID", mock.Anything, authorID).Return(author, nil)
	mockRepo.On("GetTeamMembers", mock.Anything, "backend").Return(members, nil)

	pr, err := service.CreatePR(context.Background(), prID, "Feature", authorID)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "validation error")
	mockRepo.AssertExpectations(t)
}

func TestCreatePR_EmptyName(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.New()

	pr, err := service.CreatePR(context.Background(), prID, "", authorID)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "validation error")
}

func TestCreatePR_NilUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	pr, err := service.CreatePR(context.Background(), uuid.Nil, "Feature", uuid.New())

	assert.Error(t, err)
	assert.Nil(t, pr)
}

// ========== MergePR Edge Cases ==========

func TestMergePR_NilUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	pr, err := service.MergePR(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "pull_request_id cannot be nil UUID")
}

func TestMergePR_PRNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	mockRepo.On("GetPRByID", mock.Anything, prID).Return(nil, errors.New("PR_NOT_FOUND"))

	pr, err := service.MergePR(context.Background(), prID)

	assert.Error(t, err)
	assert.Nil(t, pr)
	mockRepo.AssertExpectations(t)
}

// ========== ReassignReviewer Tests ==========

func TestReassignReviewer_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &ReviewerService{
		repo:      mockRepo,
		validator: domain.NewValidator(),
		rand:      rand.New(rand.NewSource(1)),
	}

	prID := uuid.New()
	authorID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	oldReviewerID := uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	newReviewerID := uuid.MustParse("7c9e6679-7425-40de-944b-e07fc1f90ae7")

	pr := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		PullRequestName:   "Feature",
		AuthorID:          authorID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{oldReviewerID},
	}

	oldUser := &domain.User{
		UserID:   oldReviewerID,
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	members := []domain.User{
		{UserID: authorID, Username: "Alice", IsActive: true, TeamName: "backend"},
		{UserID: oldReviewerID, Username: "Bob", IsActive: true, TeamName: "backend"},
		{UserID: newReviewerID, Username: "Dave", IsActive: true, TeamName: "backend"},
	}

	updatedPR := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		PullRequestName:   "Feature",
		AuthorID:          authorID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{newReviewerID},
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(pr, nil).Once()
	mockRepo.On("GetUserByID", mock.Anything, oldReviewerID).Return(oldUser, nil)
	mockRepo.On("GetTeamMembers", mock.Anything, "backend").Return(members, nil)
	mockRepo.On("ReplaceReviewer", mock.Anything, prID, oldReviewerID, newReviewerID).Return(nil)
	mockRepo.On("GetPRByID", mock.Anything, prID).Return(updatedPR, nil).Once()

	resultPR, newID, err := service.ReassignReviewer(context.Background(), prID, oldReviewerID)

	assert.NoError(t, err)
	assert.Equal(t, newReviewerID, newID)
	assert.Len(t, resultPR.AssignedReviewers, 1)
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_PRMerged(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	reviewerID := uuid.New()

	pr := &domain.PullRequestWithReviewers{
		PullRequestID: prID,
		Status:        "merged",
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(pr, nil)

	resultPR, newID, err := service.ReassignReviewer(context.Background(), prID, reviewerID)

	assert.Error(t, err)
	assert.Nil(t, resultPR)
	assert.Equal(t, uuid.Nil, newID)
	assert.Contains(t, err.Error(), "PR_MERGED")
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	reviewerID := uuid.New()
	otherReviewerID := uuid.New()

	pr := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{otherReviewerID},
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(pr, nil)

	resultPR, newID, err := service.ReassignReviewer(context.Background(), prID, reviewerID)

	assert.Error(t, err)
	assert.Nil(t, resultPR)
	assert.Equal(t, uuid.Nil, newID)
	assert.Contains(t, err.Error(), "NOT_ASSIGNED")
	mockRepo.AssertExpectations(t)
}

func TestReassignReviewer_NoCandidates(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prID := uuid.New()
	authorID := uuid.New()
	reviewerID := uuid.New()

	pr := &domain.PullRequestWithReviewers{
		PullRequestID:     prID,
		AuthorID:          authorID,
		Status:            "open",
		AssignedReviewers: []uuid.UUID{reviewerID},
	}

	user := &domain.User{
		UserID:   reviewerID,
		TeamName: "backend",
	}

	members := []domain.User{
		{UserID: authorID, Username: "Alice", IsActive: true, TeamName: "backend"},
		{UserID: reviewerID, Username: "Bob", IsActive: true, TeamName: "backend"},
	}

	mockRepo.On("GetPRByID", mock.Anything, prID).Return(pr, nil)
	mockRepo.On("GetUserByID", mock.Anything, reviewerID).Return(user, nil)
	mockRepo.On("GetTeamMembers", mock.Anything, "backend").Return(members, nil)

	resultPR, newID, err := service.ReassignReviewer(context.Background(), prID, reviewerID)

	assert.Error(t, err)
	assert.Nil(t, resultPR)
	assert.Equal(t, uuid.Nil, newID)
	assert.Contains(t, err.Error(), "NO_CANDIDATE")
	mockRepo.AssertExpectations(t)
}

func TestGetUserReviews_NilUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	prs, err := service.GetUserReviews(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, prs)
	assert.Contains(t, err.Error(), "user_id cannot be nil UUID")
}

func TestGetUserReviews_EmptyResult(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReviewerService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetPRsByReviewer", mock.Anything, userID).Return([]domain.PullRequestShort{}, nil)

	prs, err := service.GetUserReviews(context.Background(), userID)

	assert.NoError(t, err)
	assert.Empty(t, prs)
	mockRepo.AssertExpectations(t)
}

func (m *MockRepository) GetUserAssignmentStats(ctx context.Context) ([]domain.UserAssignmentStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserAssignmentStats), args.Error(1)
}

func (m *MockRepository) GetPRStats(ctx context.Context) (*domain.PRStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PRStats), args.Error(1)
}

func (m *MockRepository) GetTotalUsers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetTotalTeams(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetActiveUsers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
