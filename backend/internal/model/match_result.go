package model

import (
	"time"

	"github.com/google/uuid"
)

// MatchResult represents a participant's result in a match
type MatchResult struct {
	ID             uuid.UUID `json:"id"`
	MatchID        uuid.UUID `json:"match_id"`
	ParticipantID  uuid.UUID `json:"participant_id"`
	Position       *int      `json:"position,omitempty"`
	Points         float64   `json:"points"`
	FastestLap     bool      `json:"fastest_lap"`
	DNF            bool      `json:"dnf"`
	DNFReason      *string   `json:"dnf_reason,omitempty"`
	SprintPosition *int      `json:"sprint_position,omitempty"`
	SprintPoints   float64   `json:"sprint_points"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Joined fields for display
	ParticipantName *string `json:"participant_name,omitempty"`
	TeamName        *string `json:"team_name,omitempty"`
}

// CreateMatchResultRequest represents a request to create/update a match result
type CreateMatchResultRequest struct {
	ParticipantID  uuid.UUID `json:"participant_id" validate:"required"`
	Position       *int      `json:"position,omitempty"`
	Points         float64   `json:"points"`
	FastestLap     bool      `json:"fastest_lap"`
	DNF            bool      `json:"dnf"`
	DNFReason      *string   `json:"dnf_reason,omitempty"`
	SprintPosition *int      `json:"sprint_position,omitempty"`
	SprintPoints   float64   `json:"sprint_points"`
}

// BulkUpdateResultsRequest represents a request to update multiple results at once
type BulkUpdateResultsRequest struct {
	Results []CreateMatchResultRequest `json:"results" validate:"required"`
}

// ListMatchResultsResponse represents the response for listing match results
type ListMatchResultsResponse struct {
	Results []*MatchResult `json:"results"`
	Total   int            `json:"total"`
}
