package model

import (
	"time"

	"github.com/google/uuid"
)

// TeamChangeActionType represents the type of action performed on a team change request
type TeamChangeActionType string

const (
	TeamChangeActionCreate  TeamChangeActionType = "CREATE"
	TeamChangeActionApprove TeamChangeActionType = "APPROVE"
	TeamChangeActionReject  TeamChangeActionType = "REJECT"
	TeamChangeActionCancel  TeamChangeActionType = "CANCEL"
)

// TeamChangeActivityLog represents an audit log entry for team change actions
type TeamChangeActivityLog struct {
	ID            uuid.UUID            `json:"id"`
	ActorID       uuid.UUID            `json:"actor_id"`
	RequestID     uuid.UUID            `json:"request_id"`
	ParticipantID uuid.UUID            `json:"participant_id"`
	ActionType    TeamChangeActionType `json:"action_type"`
	Details       map[string]any       `json:"details"`
	CreatedAt     time.Time            `json:"created_at"`

	// Joined fields
	ActorNickname       *string `json:"actor_nickname,omitempty"`
	ParticipantNickname *string `json:"participant_nickname,omitempty"`
}

// TeamChangeActivityListResponse represents the response for listing team change activity logs
type TeamChangeActivityListResponse struct {
	Activities []*TeamChangeActivityLog `json:"activities"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
}
