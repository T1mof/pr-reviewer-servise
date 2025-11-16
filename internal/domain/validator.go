package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// Валидация UUID.
func (v *Validator) ValidateUUID(id string, fieldName string) (uuid.UUID, error) {
	if strings.TrimSpace(id) == "" {
		return uuid.Nil, fmt.Errorf("%s cannot be empty", fieldName)
	}

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a valid UUID: %w", fieldName, err)
	}

	if parsedUUID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%s cannot be nil UUID", fieldName)
	}

	return parsedUUID, nil
}

// Валидация статуса PR.
func (v *Validator) ValidatePRStatus(status string) error {
	if status != StatusOpen && status != StatusMerged {
		return fmt.Errorf("invalid PR status: %s, must be OPEN or MERGED", status)
	}
	return nil
}

// Валидация Team.
func (v *Validator) ValidateTeam(team *Team) error {
	if strings.TrimSpace(team.TeamName) == "" {
		return errors.New("team_name cannot be empty")
	}
	if len(team.TeamName) > 255 {
		return errors.New("team_name too long (max 255 characters)")
	}
	if len(team.Members) == 0 {
		return errors.New("team must have at least one member")
	}

	seen := make(map[uuid.UUID]bool)
	for _, member := range team.Members {
		if err := v.ValidateTeamMember(&member); err != nil {
			return err
		}
		if seen[member.UserID] {
			return fmt.Errorf("duplicate user_id in team: %s", member.UserID)
		}
		seen[member.UserID] = true
	}

	return nil
}

// Валидация TeamMember.
func (v *Validator) ValidateTeamMember(member *TeamMember) error {
	if member.UserID == uuid.Nil {
		return errors.New("user_id cannot be nil UUID")
	}
	if strings.TrimSpace(member.Username) == "" {
		return errors.New("username cannot be empty")
	}
	if len(member.Username) > 255 {
		return errors.New("username too long (max 255 characters)")
	}
	return nil
}

// Валидация PullRequest.
func (v *Validator) ValidatePullRequest(pr *PullRequest) error {
	if pr.PullRequestID == uuid.Nil {
		return errors.New("pull_request_id cannot be nil UUID")
	}
	if strings.TrimSpace(pr.PullRequestName) == "" {
		return errors.New("pull_request_name cannot be empty")
	}
	if len(pr.PullRequestName) > 255 {
		return errors.New("pull_request_name too long (max 255 characters)")
	}
	if pr.AuthorID == uuid.Nil {
		return errors.New("author_id cannot be nil UUID")
	}

	return v.ValidatePRStatus(pr.Status)
}

func (v *Validator) ValidateReviewersCount(reviewers []uuid.UUID) error {
	if len(reviewers) < 2 {
		return errors.New("not enough reviewers: minimum 2 required")
	}
	if len(reviewers) > 2 {
		return errors.New("cannot assign more than 2 reviewers")
	}
	return nil
}
