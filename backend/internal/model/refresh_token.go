package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token for multi-device support
type RefreshToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Token      string    `json:"-"` // Never expose the actual token
	DeviceInfo *string   `json:"device_info,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// RefreshTokenSession represents a user's active session (for listing)
type RefreshTokenSession struct {
	ID         uuid.UUID `json:"id"`
	DeviceInfo *string   `json:"device_info,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	IsCurrent  bool      `json:"is_current"` // Whether this is the current session
}
