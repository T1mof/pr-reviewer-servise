package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
)

// ServiceInterface определяет методы бизнес-логики.
type ServiceInterface interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*domain.User, error)
	CreatePR(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (*domain.PullRequestWithReviewers, error)
	MergePR(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID uuid.UUID) (*domain.PullRequestWithReviewers, uuid.UUID, error)
	GetUserReviews(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error)
	GetStatistics(ctx context.Context) (*domain.Statistics, error)
}

// Compile-time проверка.
var _ ServiceInterface = (*ReviewerService)(nil)
