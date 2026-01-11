package model

import (
	"time"

	"github.com/google/uuid"
)

// LeagueStatus represents the status of a league
type LeagueStatus string

const (
	LeagueStatusDraft      LeagueStatus = "draft"
	LeagueStatusOpen       LeagueStatus = "open"
	LeagueStatusInProgress LeagueStatus = "in_progress"
	LeagueStatusCompleted  LeagueStatus = "completed"
	LeagueStatusCancelled  LeagueStatus = "cancelled"
)

// League represents a league in the system
type League struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	Status      LeagueStatus `json:"status"`
	Season      int          `json:"season"`
	CreatedBy   uuid.UUID    `json:"created_by"`
	StartDate   *time.Time   `json:"start_date,omitempty"`
	EndDate     *time.Time   `json:"end_date,omitempty"`
	MatchTime   *string      `json:"match_time,omitempty"`
	Rules       *string      `json:"rules,omitempty"`
	Settings    *string      `json:"settings,omitempty"`
	ContactInfo *string      `json:"contact_info,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CreateLeagueRequest represents a request to create a league
type CreateLeagueRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty"`
	Season      int     `json:"season" validate:"min=1"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	MatchTime   *string `json:"match_time,omitempty"`
	Rules       *string `json:"rules,omitempty"`
	Settings    *string `json:"settings,omitempty"`
	ContactInfo *string `json:"contact_info,omitempty"`
}

// UpdateLeagueRequest represents a request to update a league
type UpdateLeagueRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Season      *int    `json:"season,omitempty" validate:"omitempty,min=1"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	MatchTime   *string `json:"match_time,omitempty"`
	Rules       *string `json:"rules,omitempty"`
	Settings    *string `json:"settings,omitempty"`
	ContactInfo *string `json:"contact_info,omitempty"`
}

// ListLeaguesResponse represents the response for listing leagues
type ListLeaguesResponse struct {
	Leagues    []*League `json:"leagues"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
