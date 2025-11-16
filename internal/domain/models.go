package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	Username string    `db:"username" json:"username"`
	TeamID   uuid.UUID `db:"team_id" json:"-"`
	TeamName string    `db:"team_name" json:"team_name,omitempty"`
	IsActive bool      `db:"is_active" json:"is_active"`
}

type Team struct {
	TeamID   uuid.UUID    `db:"team_id" json:"-"`
	TeamName string       `db:"team_name" json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsActive bool      `json:"is_active"`
}

type PullRequest struct {
	PullRequestID   uuid.UUID  `db:"pull_request_id" json:"pull_request_id"`
	PullRequestName string     `db:"pull_request_name" json:"pull_request_name"`
	AuthorID        uuid.UUID  `db:"author_id" json:"author_id"`
	Status          string     `db:"status" json:"status"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt,omitempty"`
	MergedAt        *time.Time `db:"merged_at" json:"mergedAt,omitempty"`
}

type PullRequestWithReviewers struct {
	PullRequestID     uuid.UUID   `json:"pull_request_id"`
	PullRequestName   string      `json:"pull_request_name"`
	AuthorID          uuid.UUID   `json:"author_id"`
	Status            string      `json:"status"`
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
	CreatedAt         time.Time   `json:"createdAt"`
	MergedAt          *time.Time  `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   uuid.UUID `json:"pull_request_id"`
	PullRequestName string    `json:"pull_request_name"`
	AuthorID        uuid.UUID `json:"author_id"`
	Status          string    `json:"status"`
}

const (
	StatusOpen   = "open"
	StatusMerged = "merged"
)
