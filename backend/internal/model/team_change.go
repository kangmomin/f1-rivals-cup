package model

import (
	"time"

	"github.com/google/uuid"
)

// TeamChangeRequestStatus represents the status of a team change request
type TeamChangeRequestStatus string

const (
	TeamChangeStatusPending  TeamChangeRequestStatus = "pending"
	TeamChangeStatusApproved TeamChangeRequestStatus = "approved"
	TeamChangeStatusRejected TeamChangeRequestStatus = "rejected"
)

// TeamChangeRequest represents a team change request
type TeamChangeRequest struct {
	ID                uuid.UUID               `json:"id"`
	ParticipantID     uuid.UUID               `json:"participant_id"`
	CurrentTeamName   *string                 `json:"current_team_name,omitempty"`
	RequestedTeamName string                  `json:"requested_team_name"`
	Status            TeamChangeRequestStatus `json:"status"`
	Reason            *string                 `json:"reason,omitempty"`
	ReviewedBy        *uuid.UUID              `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time              `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`

	// Joined fields
	ParticipantName *string    `json:"participant_name,omitempty"`
	LeagueID        *uuid.UUID `json:"league_id,omitempty"`
	ReviewerName    *string    `json:"reviewer_name,omitempty"`
}

// ParticipantTeamHistory represents a record of team membership
type ParticipantTeamHistory struct {
	ID              uuid.UUID  `json:"id"`
	ParticipantID   uuid.UUID  `json:"participant_id"`
	TeamName        string     `json:"team_name"`
	EffectiveFrom   time.Time  `json:"effective_from"`
	EffectiveUntil  *time.Time `json:"effective_until,omitempty"`
	ChangeRequestID *uuid.UUID `json:"change_request_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateTeamChangeRequest represents a request to create a team change request
type CreateTeamChangeRequest struct {
	RequestedTeamName string  `json:"requested_team_name" validate:"required,max=100"`
	Reason            *string `json:"reason,omitempty"`
}

// ReviewTeamChangeRequest represents a request to review (approve/reject) a team change request
type ReviewTeamChangeRequest struct {
	Status TeamChangeRequestStatus `json:"status" validate:"required"`
	Reason *string                 `json:"reason,omitempty"`
}

// TeamChangeRequestListResponse represents the response for listing team change requests
type TeamChangeRequestListResponse struct {
	Requests []*TeamChangeRequest `json:"requests"`
	Total    int                  `json:"total"`
}
