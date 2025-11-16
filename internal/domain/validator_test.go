package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateTeam_Success(t *testing.T) {
	validator := NewValidator()

	team := &Team{
		TeamName: "backend",
		Members: []TeamMember{
			{
				UserID:   uuid.New(),
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	err := validator.ValidateTeam(team)
	assert.NoError(t, err)
}

func TestValidateTeam_EmptyName(t *testing.T) {
	validator := NewValidator()

	team := &Team{
		TeamName: "",
		Members: []TeamMember{
			{UserID: uuid.New(), Username: "Alice", IsActive: true},
		},
	}

	err := validator.ValidateTeam(team)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team_name cannot be empty")
}

func TestValidateTeam_EmptyMembers(t *testing.T) {
	validator := NewValidator()

	team := &Team{
		TeamName: "backend",
		Members:  []TeamMember{},
	}

	err := validator.ValidateTeam(team)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team must have at least one member")
}

func TestValidateTeam_MemberWithEmptyUsername(t *testing.T) {
	validator := NewValidator()

	team := &Team{
		TeamName: "backend",
		Members: []TeamMember{
			{
				UserID:   uuid.New(),
				Username: "",
				IsActive: true,
			},
		},
	}

	err := validator.ValidateTeam(team)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
}

func TestValidateTeam_MemberWithNilUUID(t *testing.T) {
	validator := NewValidator()

	team := &Team{
		TeamName: "backend",
		Members: []TeamMember{
			{
				UserID:   uuid.Nil,
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	err := validator.ValidateTeam(team)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id cannot be nil")
}

func TestValidatePullRequest_Success(t *testing.T) {
	validator := NewValidator()

	pr := &PullRequest{
		PullRequestID:   uuid.New(),
		PullRequestName: "Add new feature",
		AuthorID:        uuid.New(),
		Status:          StatusOpen,
	}

	err := validator.ValidatePullRequest(pr)
	assert.NoError(t, err)
}

func TestValidatePullRequest_NilPRID(t *testing.T) {
	validator := NewValidator()

	pr := &PullRequest{
		PullRequestID:   uuid.Nil,
		PullRequestName: "Add new feature",
		AuthorID:        uuid.New(),
		Status:          StatusOpen,
	}

	err := validator.ValidatePullRequest(pr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pull_request_id cannot be nil")
}

func TestValidatePullRequest_EmptyName(t *testing.T) {
	validator := NewValidator()

	pr := &PullRequest{
		PullRequestID:   uuid.New(),
		PullRequestName: "",
		AuthorID:        uuid.New(),
		Status:          StatusOpen,
	}

	err := validator.ValidatePullRequest(pr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pull_request_name cannot be empty")
}

func TestValidatePullRequest_NilAuthorID(t *testing.T) {
	validator := NewValidator()

	pr := &PullRequest{
		PullRequestID:   uuid.New(),
		PullRequestName: "Add new feature",
		AuthorID:        uuid.Nil,
		Status:          StatusOpen,
	}

	err := validator.ValidatePullRequest(pr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "author_id cannot be nil")
}

func TestValidateReviewersCount_Success(t *testing.T) {
	validator := NewValidator()

	reviewers := []uuid.UUID{uuid.New(), uuid.New()}

	err := validator.ValidateReviewersCount(reviewers)
	assert.NoError(t, err)
}

func TestValidateReviewersCount_TooFew_Zero(t *testing.T) {
	validator := NewValidator()

	reviewers := []uuid.UUID{}

	err := validator.ValidateReviewersCount(reviewers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 2 required")
}

func TestValidateReviewersCount_TooFew_One(t *testing.T) {
	validator := NewValidator()

	reviewers := []uuid.UUID{uuid.New()}

	err := validator.ValidateReviewersCount(reviewers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 2 required")
}

func TestValidateReviewersCount_TooMany(t *testing.T) {
	validator := NewValidator()

	reviewers := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	err := validator.ValidateReviewersCount(reviewers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot assign more than 2 reviewers")
}
