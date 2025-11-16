package domain

import "github.com/google/uuid"

// UserAssignmentStats статистика назначений по пользователю.
type UserAssignmentStats struct {
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	Username          string    `json:"username" db:"username"`
	TeamName          string    `json:"team_name" db:"team_name"`
	TotalAssignments  int       `json:"total_assignments" db:"total_assignments"`
	OpenAssignments   int       `json:"open_assignments" db:"open_assignments"`
	MergedAssignments int       `json:"merged_assignments" db:"merged_assignments"`
}

// PRStats общая статистика по PR.
type PRStats struct {
	TotalOpen         int     `json:"total_open" db:"total_open"`
	TotalMerged       int     `json:"total_merged" db:"total_merged"`
	TotalPRs          int     `json:"total_prs" db:"total_prs"`
	AvgMergeTimeHours float64 `json:"avg_merge_time_hours" db:"avg_merge_time_hours"`
}

// Statistics общая статистика сервиса.
type Statistics struct {
	PRStats     PRStats               `json:"pr_stats"`
	UserStats   []UserAssignmentStats `json:"user_stats"`
	TotalUsers  int                   `json:"total_users"`
	TotalTeams  int                   `json:"total_teams"`
	ActiveUsers int                   `json:"active_users"`
}
