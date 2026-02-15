package model

import (
	"time"

	"github.com/google/uuid"
)

// OAuthAccount represents a linked OAuth provider account
type OAuthAccount struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Provider         string     `json:"provider"`
	ProviderID       string     `json:"provider_id"`
	ProviderUsername  string     `json:"provider_username"`
	ProviderAvatar   string     `json:"provider_avatar"`
	ProviderEmail    string     `json:"provider_email"`
	AccessToken      string     `json:"-"`
	RefreshToken     string     `json:"-"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// DiscordUser represents the Discord API /users/@me response
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`
}

// DiscordOAuthTokenResponse represents Discord's OAuth2 token response
type DiscordOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// OAuthLinkStatus represents the link status of an OAuth provider for the frontend
type OAuthLinkStatus struct {
	Provider         string     `json:"provider"`
	Linked           bool       `json:"linked"`
	ProviderUsername  string     `json:"provider_username,omitempty"`
	ProviderAvatar   string     `json:"provider_avatar,omitempty"`
	LinkedAt         *time.Time `json:"linked_at,omitempty"`
}
