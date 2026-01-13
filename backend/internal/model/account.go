package model

import (
	"time"

	"github.com/google/uuid"
)

type OwnerType string

const (
	OwnerTypeTeam        OwnerType = "team"
	OwnerTypeParticipant OwnerType = "participant"
	OwnerTypeSystem      OwnerType = "system"
)

type Account struct {
	ID        uuid.UUID `json:"id"`
	LeagueID  uuid.UUID `json:"league_id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	OwnerType OwnerType `json:"owner_type"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Joined fields
	OwnerName string `json:"owner_name,omitempty"`
}

type SetBalanceRequest struct {
	Balance int64 `json:"balance"`
}

type ListAccountsResponse struct {
	Accounts []*Account `json:"accounts"`
	Total    int        `json:"total"`
}
