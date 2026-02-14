package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
)

const (
	discordAuthorizeURL = "https://discord.com/api/oauth2/authorize"
	discordTokenURL     = "https://discord.com/api/oauth2/token"
	discordUserURL      = "https://discord.com/api/users/@me"
)

// DiscordOAuthService handles Discord OAuth2 operations
type DiscordOAuthService struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewDiscordOAuthService creates a new DiscordOAuthService
func NewDiscordOAuthService(clientID, clientSecret, redirectURI string) *DiscordOAuthService {
	return &DiscordOAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// GetAuthorizationURL returns the Discord OAuth2 authorization URL
func (s *DiscordOAuthService) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "identify email")
	params.Set("state", state)

	return discordAuthorizeURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for access and refresh tokens
func (s *DiscordOAuthService) ExchangeCode(code string) (*model.DiscordOAuthTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest("POST", discordTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp model.DiscordOAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUser retrieves the Discord user info using an access token
func (s *DiscordOAuthService) GetUser(accessToken string) (*model.DiscordUser, error) {
	req, err := http.NewRequest("GET", discordUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get discord user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord user API failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var user model.DiscordUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}
