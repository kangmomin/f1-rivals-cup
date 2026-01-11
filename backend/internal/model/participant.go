package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ParticipantStatus string

const (
	ParticipantStatusPending  ParticipantStatus = "pending"
	ParticipantStatusApproved ParticipantStatus = "approved"
	ParticipantStatusRejected ParticipantStatus = "rejected"
)

type ParticipantRole string

const (
	RoleDirector ParticipantRole = "director" // 감독
	RolePlayer   ParticipantRole = "player"   // 선수
	RoleReserve  ParticipantRole = "reserve"  // 리저브선수
	RoleEngineer ParticipantRole = "engineer" // 엔지니어
)

// LeagueParticipant represents a user's participation in a league
type LeagueParticipant struct {
	ID        uuid.UUID         `json:"id"`
	LeagueID  uuid.UUID         `json:"league_id"`
	UserID    uuid.UUID         `json:"user_id"`
	Status    ParticipantStatus `json:"status"`
	Roles     pq.StringArray    `json:"roles"`
	TeamName  *string           `json:"team_name,omitempty"`
	Message   *string           `json:"message,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`

	// Joined fields
	UserNickname *string `json:"user_nickname,omitempty"`
	UserEmail    *string `json:"user_email,omitempty"`
	LeagueName   *string `json:"league_name,omitempty"`
}

// JoinLeagueRequest represents a request to join a league
type JoinLeagueRequest struct {
	TeamName *string  `json:"team_name,omitempty"`
	Message  *string  `json:"message,omitempty"`
	Roles    []string `json:"roles" validate:"required,min=1"`
}

// UpdateParticipantRequest represents a request to update participant status
type UpdateParticipantRequest struct {
	Status ParticipantStatus `json:"status" validate:"required"`
}

// ListParticipantsResponse represents the response for listing participants
type ListParticipantsResponse struct {
	Participants []*LeagueParticipant `json:"participants"`
	Total        int                  `json:"total"`
}
