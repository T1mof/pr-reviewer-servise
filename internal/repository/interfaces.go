package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error)
	SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error
}

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *domain.PullRequest, reviewers []uuid.UUID) error
	GetPRByID(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error)
	PRExists(ctx context.Context, prID uuid.UUID) (bool, error)
	UpdatePRStatus(ctx context.Context, prID uuid.UUID, status string, mergedAt *time.Time) error
	GetReviewersByPR(ctx context.Context, prID uuid.UUID) ([]uuid.UUID, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID uuid.UUID) error
	GetPRsByReviewer(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error)
}

type StatsRepository interface {
	GetUserAssignmentStats(ctx context.Context) ([]domain.UserAssignmentStats, error)
	GetPRStats(ctx context.Context) (*domain.PRStats, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetTotalTeams(ctx context.Context) (int, error)
	GetActiveUsers(ctx context.Context) (int, error)
}

// RepositoryInterface объединяет все интерфейсы.
type RepositoryInterface interface {
	TeamRepository
	UserRepository
	PullRequestRepository
	StatsRepository
}
