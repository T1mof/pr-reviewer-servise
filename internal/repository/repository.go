package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ========================================
// TeamRepository Methods
// ========================================

func (r *Repository) CreateTeam(ctx context.Context, team *domain.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Failed to rollback transaction", "error", err)
		}
	}()

	var teamID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO teams (team_name) 
		VALUES ($1) 
		RETURNING team_id
	`, team.TeamName).Scan(&teamID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return errors.New("TEAM_EXISTS")
		}
		return fmt.Errorf("failed to insert team: %w", err)
	}

	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, team_id, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET
				username = EXCLUDED.username,
				team_id = EXCLUDED.team_id,
				is_active = EXCLUDED.is_active,
				updated_at = NOW()
		`, member.UserID, member.Username, teamID, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Team created in DB", "team_name", team.TeamName, "team_id", teamID)
	return nil
}

func (r *Repository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team
	var teamID uuid.UUID

	err := r.db.QueryRowContext(ctx, `
		SELECT team_id, team_name 
		FROM teams 
		WHERE team_name = $1
	`, teamName).Scan(&teamID, &team.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("TEAM_NOT_FOUND")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_id = $1
		ORDER BY username
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	return &team, nil
}

func (r *Repository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)
	`, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}

// ========================================
// UserRepository Methods
// ========================================

func (r *Repository) UpsertUser(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, (SELECT team_id FROM teams WHERE team_name = $3), $4)
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			team_id = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT u.user_id, u.username, t.team_name, u.is_active
		FROM users u
		JOIN teams t ON u.team_id = t.team_id
		WHERE u.user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("USER_NOT_FOUND")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *Repository) GetTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.user_id, u.username, t.team_name, u.is_active
		FROM users u
		JOIN teams t ON u.team_id = t.team_id
		WHERE t.team_name = $1
		ORDER BY u.username
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var members []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		members = append(members, user)
	}
	return members, nil
}

func (r *Repository) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users 
		SET is_active = $1, updated_at = NOW() 
		WHERE user_id = $2
	`, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("USER_NOT_FOUND")
	}

	slog.Info("User active status updated", "user_id", userID, "is_active", isActive)
	return nil
}

// ========================================
// PullRequestRepository Methods
// ========================================

func (r *Repository) CreatePR(ctx context.Context, pr *domain.PullRequest, reviewers []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Failed to rollback transaction", "error", err)
		}
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return errors.New("PR_EXISTS")
		}
		return fmt.Errorf("failed to insert pull request: %w", err)
	}

	for _, reviewerID := range reviewers {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`, pr.PullRequestID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to insert reviewer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Pull request created", "pr_id", pr.PullRequestID, "reviewers_count", len(reviewers))
	return nil
}

func (r *Repository) GetPRByID(ctx context.Context, prID uuid.UUID) (*domain.PullRequestWithReviewers, error) {
	var pr domain.PullRequestWithReviewers

	err := r.db.QueryRowContext(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("PR_NOT_FOUND")
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID uuid.UUID
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, nil
}

func (r *Repository) PRExists(ctx context.Context, prID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}
	return exists, nil
}

func (r *Repository) UpdatePRStatus(ctx context.Context, prID uuid.UUID, status string, mergedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2 
		WHERE pull_request_id = $3
	`, status, mergedAt, prID)
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}

	slog.Info("PR status updated", "pr_id", prID, "status", status)
	return nil
}

func (r *Repository) GetReviewersByPR(ctx context.Context, prID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []uuid.UUID
	for rows.Next() {
		var reviewerID uuid.UUID
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}
	return reviewers, nil
}

func (r *Repository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE pr_reviewers 
		SET user_id = $1 
		WHERE pull_request_id = $2 AND user_id = $3
	`, newUserID, prID, oldUserID)
	if err != nil {
		return fmt.Errorf("failed to replace reviewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("REVIEWER_NOT_FOUND")
	}

	slog.Info("Reviewer replaced", "pr_id", prID, "old_user", oldUserID, "new_user", newUserID)
	return nil
}

func (r *Repository) GetPRsByReviewer(ctx context.Context, userID uuid.UUID) ([]domain.PullRequestShort, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// ========================================
// StatsRepository Methods
// ========================================

// GetUserAssignmentStats возвращает статистику назначений по пользователям.
func (r *Repository) GetUserAssignmentStats(ctx context.Context) ([]domain.UserAssignmentStats, error) {
	query := `
		SELECT 
			u.user_id,
			u.username,
			t.team_name,
			COALESCE(COUNT(pr.pull_request_id), 0) as total_assignments,
			COALESCE(COUNT(CASE WHEN pr_main.status = 'open' THEN 1 END), 0) as open_assignments,
			COALESCE(COUNT(CASE WHEN pr_main.status = 'merged' THEN 1 END), 0) as merged_assignments
		FROM users u
		JOIN teams t ON u.team_id = t.team_id
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
		LEFT JOIN pull_requests pr_main ON pr.pull_request_id = pr_main.pull_request_id
		GROUP BY u.user_id, u.username, t.team_name
		ORDER BY total_assignments DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignment stats: %w", err)
	}
	defer rows.Close()

	var stats []domain.UserAssignmentStats
	for rows.Next() {
		var stat domain.UserAssignmentStats
		err := rows.Scan(
			&stat.UserID,
			&stat.Username,
			&stat.TeamName,
			&stat.TotalAssignments,
			&stat.OpenAssignments,
			&stat.MergedAssignments,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetPRStats возвращает общую статистику по PR.
func (r *Repository) GetPRStats(ctx context.Context) (*domain.PRStats, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'open') as total_open,
			COUNT(*) FILTER (WHERE status = 'merged') as total_merged,
			COUNT(*) as total_prs,
			COALESCE(AVG(EXTRACT(EPOCH FROM (merged_at - created_at))/3600) FILTER (WHERE merged_at IS NOT NULL), 0) as avg_merge_time_hours
		FROM pull_requests
	`

	var stats domain.PRStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalOpen,
		&stats.TotalMerged,
		&stats.TotalPRs,
		&stats.AvgMergeTimeHours,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}

	return &stats, nil
}

// GetTotalUsers возвращает общее количество пользователей.
func (r *Repository) GetTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total users: %w", err)
	}
	return count, nil
}

// GetTotalTeams возвращает общее количество команд.
func (r *Repository) GetTotalTeams(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total teams: %w", err)
	}
	return count, nil
}

// GetActiveUsers возвращает количество активных пользователей.
func (r *Repository) GetActiveUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active users: %w", err)
	}
	return count, nil
}

// ========================================
// Compile-time interface check
// ========================================

var _ RepositoryInterface = (*Repository)(nil)
