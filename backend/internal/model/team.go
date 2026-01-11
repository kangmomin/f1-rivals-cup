package model

import (
	"time"

	"github.com/google/uuid"
)

// Team represents a team within a league
type Team struct {
	ID         uuid.UUID `json:"id"`
	LeagueID   uuid.UUID `json:"league_id"`
	Name       string    `json:"name"`
	Color      *string   `json:"color,omitempty"`
	IsOfficial bool      `json:"is_official"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateTeamRequest represents a request to create a team
type CreateTeamRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=100"`
	Color      *string `json:"color,omitempty"`
	IsOfficial bool    `json:"is_official"`
}

// UpdateTeamRequest represents a request to update a team
type UpdateTeamRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Color *string `json:"color,omitempty"`
}

// ListTeamsResponse represents a response containing a list of teams
type ListTeamsResponse struct {
	Teams []*Team `json:"teams"`
	Total int     `json:"total"`
}
