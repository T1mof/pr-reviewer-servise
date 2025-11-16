package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
	"github.com/T1mof/pr-reviewer-service/internal/repository"
)

type ReviewerService struct {
	repo      repository.RepositoryInterface
	validator *domain.Validator
	rand      *rand.Rand
}

func NewReviewerService(repo repository.RepositoryInterface) *ReviewerService {
	return &ReviewerService{
		repo:      repo,
		validator: domain.NewValidator(),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ========================================
// Team Methods
// ========================================

func (s *ReviewerService) CreateTeam(ctx context.Context, team *domain.Team) error {
	if err := s.validator.ValidateTeam(team); err != nil {
		slog.Warn("Team validation failed", "team_name", team.TeamName, "error", err)
		return fmt.Errorf("validation error: %w", err)
	}

	exists, err := s.repo.TeamExists(ctx, team.TeamName)
	if err != nil {
		slog.Error("Failed to check team existence", "team_name", team.TeamName, "error", err)
		return err
	}
	if exists {
		slog.Warn("Team already exists", "team_name", team.TeamName)
		return errors.New("TEAM_EXISTS")
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		slog.Error("Failed to create team", "team_name", team.TeamName, "error", err)
		return err
	}

	slog.Info("Team created", "team_name", team.TeamName, "members_count", len(team.Members))
	return nil
}

func (s *ReviewerService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	if teamName == "" {
		return nil, errors.New("team_name cannot be empty")
	}

	team, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		slog.Error("Failed to get team", "team_name", teamName, "error", err)
		return nil, err
	}

	return team, nil
}

// ========================================
// User Methods
// ========================================

func (s *ReviewerService) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*domain.User, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user_id cannot be nil UUID")
	}

	err := s.repo.SetUserActive(ctx, userID, isActive)
	if err != nil {
		slog.Error("Failed to set user active", "user_id", userID, "is_active", isActive, "error", err)
		return nil, fmt.Errorf("failed to set user active: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user", "user_id", userID, "error", err)
		return nil, err
	}

	slog.Info("User active status updated", "user_id", userID, "is_active", isActive)
	return user, nil
}

// ========================================
// PullRequest Methods
// ========================================

func (s *ReviewerService) CreatePR(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	pr := &domain.PullRequest{
		PullRequestID:   prID,
		PullRequestName: prName,
		AuthorID:        authorID,
		Status:          domain.StatusOpen,
	}

	if err := s.validator.ValidatePullRequest(pr); err != nil {
		slog.Warn("PR validation failed", "pr_id", prID, "error", err)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	exists, err := s.repo.PRExists(ctx, prID)
	if err != nil {
		slog.Error("Failed to check PR existence", "pr_id", prID, "error", err)
		return nil, err
	}
	if exists {
		slog.Warn("PR already exists", "pr_id", prID)
		return nil, errors.New("PR_EXISTS")
	}

	author, err := s.repo.GetUserByID(ctx, authorID)
	if err != nil {
		slog.Error("Failed to get author", "author_id", authorID, "error", err)
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	members, err := s.repo.GetTeamMembers(ctx, author.TeamName)
	if err != nil {
		slog.Error("Failed to get team members", "team_name", author.TeamName, "error", err)
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	reviewers := s.selectReviewers(members, authorID, 2)
	slog.Info("Reviewers selected", "pr_id", prID, "count", len(reviewers))

	if err := s.validator.ValidateReviewersCount(reviewers); err != nil {
		slog.Warn("Reviewers count validation failed", "pr_id", prID, "count", len(reviewers), "error", err)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	err = s.repo.CreatePR(ctx, pr, reviewers)
	if err != nil {
		slog.Error("Failed to create PR", "pr_id", prID, "error", err)
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return s.repo.GetPRByID(ctx, prID)
}

func (s *ReviewerService) MergePR(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	if prID == uuid.Nil {
		return nil, errors.New("pull_request_id cannot be nil UUID")
	}

	pr, err := s.repo.GetPRByID(ctx, prID)
	if err != nil {
		slog.Error("Failed to get PR", "pr_id", prID, "error", err)
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == domain.StatusMerged {
		slog.Info("PR already merged", "pr_id", prID)
		return pr, nil
	}

	now := time.Now()
	err = s.repo.UpdatePRStatus(ctx, prID, domain.StatusMerged, &now)
	if err != nil {
		slog.Error("Failed to merge PR", "pr_id", prID, "error", err)
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	slog.Info("PR merged", "pr_id", prID)
	return s.repo.GetPRByID(ctx, prID)
}

func (s *ReviewerService) ReassignReviewer(ctx context.Context, prID, oldUserID uuid.UUID) (*domain.PullRequestWithReviewers, uuid.UUID, error) {
	if prID == uuid.Nil || oldUserID == uuid.Nil {
		return nil, uuid.Nil, errors.New("IDs cannot be nil UUID")
	}

	pr, err := s.repo.GetPRByID(ctx, prID)
	if err != nil {
		slog.Error("Failed to get PR", "pr_id", prID, "error", err)
		return nil, uuid.Nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == domain.StatusMerged {
		slog.Warn("Cannot reassign on merged PR", "pr_id", prID)
		return nil, uuid.Nil, errors.New("PR_MERGED")
	}

	isAssigned := false
	for _, r := range pr.AssignedReviewers {
		if r == oldUserID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		slog.Warn("Reviewer not assigned", "pr_id", prID, "reviewer", oldUserID)
		return nil, uuid.Nil, errors.New("NOT_ASSIGNED")
	}

	oldUser, err := s.repo.GetUserByID(ctx, oldUserID)
	if err != nil {
		slog.Error("Failed to get old user", "user_id", oldUserID, "error", err)
		return nil, uuid.Nil, fmt.Errorf("failed to get old user: %w", err)
	}

	members, err := s.repo.GetTeamMembers(ctx, oldUser.TeamName)
	if err != nil {
		slog.Error("Failed to get team members", "team_name", oldUser.TeamName, "error", err)
		return nil, uuid.Nil, fmt.Errorf("failed to get team members: %w", err)
	}

	excludeIDs := make(map[uuid.UUID]bool)
	excludeIDs[pr.AuthorID] = true
	for _, r := range pr.AssignedReviewers {
		excludeIDs[r] = true
	}

	var candidates []domain.User
	for _, m := range members {
		if !excludeIDs[m.UserID] && m.IsActive {
			candidates = append(candidates, m)
		}
	}

	if len(candidates) == 0 {
		slog.Warn("No candidates for reassignment", "pr_id", prID, "team", oldUser.TeamName)
		return nil, uuid.Nil, errors.New("NO_CANDIDATE")
	}

	newReviewer := candidates[s.rand.Intn(len(candidates))]
	slog.Info("New reviewer selected", "pr_id", prID, "old", oldUserID, "new", newReviewer.UserID)

	err = s.repo.ReplaceReviewer(ctx, prID, oldUserID, newReviewer.UserID)
	if err != nil {
		slog.Error("Failed to replace reviewer", "pr_id", prID, "error", err)
		return nil, uuid.Nil, fmt.Errorf("failed to replace reviewer: %w", err)
	}

	updatedPR, err := s.repo.GetPRByID(ctx, prID)
	if err != nil {
		slog.Error("Failed to get updated PR", "pr_id", prID, "error", err)
		return nil, uuid.Nil, fmt.Errorf("failed to get updated PR: %w", err)
	}

	return updatedPR, newReviewer.UserID, nil
}

func (s *ReviewerService) GetUserReviews(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user_id cannot be nil UUID")
	}

	prs, err := s.repo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user reviews", "user_id", userID, "error", err)
		return nil, err
	}

	return prs, nil
}

// GetStatistics возвращает общую статистику сервиса.
func (s *ReviewerService) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	prStats, err := s.repo.GetPRStats(ctx)
	if err != nil {
		slog.Error("Failed to get PR stats", "error", err)
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}

	userStats, err := s.repo.GetUserAssignmentStats(ctx)
	if err != nil {
		slog.Error("Failed to get user assignment stats", "error", err)
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	totalUsers, err := s.repo.GetTotalUsers(ctx)
	if err != nil {
		slog.Error("Failed to get total users", "error", err)
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	totalTeams, err := s.repo.GetTotalTeams(ctx)
	if err != nil {
		slog.Error("Failed to get total teams", "error", err)
		return nil, fmt.Errorf("failed to get total teams: %w", err)
	}

	activeUsers, err := s.repo.GetActiveUsers(ctx)
	if err != nil {
		slog.Error("Failed to get active users", "error", err)
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	stats := &domain.Statistics{
		PRStats:     *prStats,
		UserStats:   userStats,
		TotalUsers:  totalUsers,
		TotalTeams:  totalTeams,
		ActiveUsers: activeUsers,
	}

	slog.Info("Statistics retrieved",
		"total_prs", stats.PRStats.TotalPRs,
		"total_users", stats.TotalUsers,
		"total_teams", stats.TotalTeams,
	)

	return stats, nil
}

// ========================================
// Helper Methods
// ========================================

func (s *ReviewerService) selectReviewers(members []domain.User, excludeID uuid.UUID, maxCount int) []uuid.UUID {
	var candidates []domain.User
	for _, m := range members {
		if m.UserID != excludeID && m.IsActive {
			candidates = append(candidates, m)
		}
	}

	count := minInt(maxCount, len(candidates))
	if count == 0 {
		slog.Warn("No active candidates", "exclude_id", excludeID, "total", len(members))
		return []uuid.UUID{}
	}

	s.rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	result := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		result[i] = candidates[i].UserID
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
