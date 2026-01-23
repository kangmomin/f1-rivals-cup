package model

import (
	"time"

	"github.com/google/uuid"
)

type MatchStatus string

const (
	MatchStatusUpcoming   MatchStatus = "upcoming"
	MatchStatusInProgress MatchStatus = "in_progress"
	MatchStatusCompleted  MatchStatus = "completed"
	MatchStatusCancelled  MatchStatus = "cancelled"
)

// Match represents a league match/race schedule
type Match struct {
	ID           uuid.UUID   `json:"id"`
	LeagueID     uuid.UUID   `json:"league_id"`
	Round        int         `json:"round"`
	Track        string      `json:"track"`
	MatchDate    string      `json:"match_date"`
	MatchTime    *string     `json:"match_time,omitempty"`
	HasSprint    bool        `json:"has_sprint"`
	SprintDate   *string     `json:"sprint_date,omitempty"`
	SprintTime   *string     `json:"sprint_time,omitempty"`
	SprintStatus MatchStatus `json:"sprint_status"`
	Status       MatchStatus `json:"status"`
	Description  *string     `json:"description,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// CreateMatchRequest represents a request to create a match
type CreateMatchRequest struct {
	Round       int     `json:"round" validate:"required,min=1"`
	Track       string  `json:"track" validate:"required"`
	MatchDate   string  `json:"match_date" validate:"required"`
	MatchTime   *string `json:"match_time,omitempty"`
	HasSprint   bool    `json:"has_sprint"`
	SprintDate  *string `json:"sprint_date,omitempty"`
	SprintTime  *string `json:"sprint_time,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateMatchRequest represents a request to update a match
type UpdateMatchRequest struct {
	Round        *int         `json:"round,omitempty"`
	Track        *string      `json:"track,omitempty"`
	MatchDate    *string      `json:"match_date,omitempty"`
	MatchTime    *string      `json:"match_time,omitempty"`
	HasSprint    *bool        `json:"has_sprint,omitempty"`
	SprintDate   *string      `json:"sprint_date,omitempty"`
	SprintTime   *string      `json:"sprint_time,omitempty"`
	SprintStatus *MatchStatus `json:"sprint_status,omitempty"`
	Status       *MatchStatus `json:"status,omitempty"`
	Description  *string      `json:"description,omitempty"`
}

// ListMatchesResponse represents the response for listing matches
type ListMatchesResponse struct {
	Matches []*Match `json:"matches"`
	Total   int      `json:"total"`
}
