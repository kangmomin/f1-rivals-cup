package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DiscordCallbackRequest represents the Discord OAuth callback request
type DiscordCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// DiscordLogin handles GET /api/v1/auth/discord
// Generates a state token and returns the Discord authorization URL
func (h *AuthHandler) DiscordLogin(c echo.Context) error {
	state := h.oauthState.Generate("login", nil)

	return c.JSON(http.StatusOK, map[string]string{
		"url": h.discordService.GetAuthorizationURL(state),
	})
}

// DiscordCallback handles POST /api/v1/auth/discord/callback
// Exchanges the authorization code and handles login/link/register flows
func (h *AuthHandler) DiscordCallback(c echo.Context) error {
	var req DiscordCallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.Code == "" || req.State == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "코드와 상태값이 필요합니다",
		})
	}

	// Validate state (one-time consumption)
	stateEntry, valid := h.oauthState.Validate(req.State)
	if !valid {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_state",
			Message: "유효하지 않거나 만료된 상태값입니다",
		})
	}

	// Exchange code for tokens
	tokenResp, err := h.discordService.ExchangeCode(req.Code)
	if err != nil {
		slog.Error("DiscordCallback: failed to exchange code", "error", err)
		return c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Error:   "oauth_error",
			Message: "Discord 인증에 실패했습니다",
		})
	}

	// Get Discord user info
	discordUser, err := h.discordService.GetUser(tokenResp.AccessToken)
	if err != nil {
		slog.Error("DiscordCallback: failed to get Discord user", "error", err)
		return c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Error:   "oauth_error",
			Message: "Discord 사용자 정보를 가져오는데 실패했습니다",
		})
	}

	ctx := c.Request().Context()
	tokenExpiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Check if this Discord account is already linked
	existingOAuth, err := h.oauthRepo.GetByProviderID(ctx, "discord", discordUser.ID)
	if err != nil && !errors.Is(err, repository.ErrOAuthAccountNotFound) {
		slog.Error("DiscordCallback: failed to check existing OAuth link", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Scenario A: Existing OAuth link found → login as that user
	if existingOAuth != nil {
		// Update OAuth account info
		existingOAuth.ProviderUsername = discordUser.Username
		existingOAuth.ProviderAvatar = discordUser.Avatar
		existingOAuth.ProviderEmail = discordUser.Email
		existingOAuth.AccessToken = tokenResp.AccessToken
		existingOAuth.RefreshToken = tokenResp.RefreshToken
		existingOAuth.TokenExpiresAt = &tokenExpiresAt
		if err := h.oauthRepo.Update(ctx, existingOAuth); err != nil {
			slog.Error("DiscordCallback: failed to update OAuth account", "error", err)
		}

		return h.loginOAuthUser(c, existingOAuth.UserID)
	}

	// Scenario B: Purpose is "link" and we have a user ID → link to existing user
	if stateEntry.Purpose == "link" && stateEntry.UserID != nil {
		oauthAccount := &model.OAuthAccount{
			UserID:           *stateEntry.UserID,
			Provider:         "discord",
			ProviderID:       discordUser.ID,
			ProviderUsername: discordUser.Username,
			ProviderAvatar:   discordUser.Avatar,
			ProviderEmail:    discordUser.Email,
			AccessToken:      tokenResp.AccessToken,
			RefreshToken:     tokenResp.RefreshToken,
			TokenExpiresAt:   &tokenExpiresAt,
		}

		if err := h.oauthRepo.Create(ctx, oauthAccount); err != nil {
			if errors.Is(err, repository.ErrOAuthAccountExists) {
				return c.JSON(http.StatusConflict, model.ErrorResponse{
					Error:   "already_linked",
					Message: "이미 연동된 Discord 계정입니다",
				})
			}
			slog.Error("DiscordCallback: failed to create OAuth link", "error", err, "userID", stateEntry.UserID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "서버 오류가 발생했습니다",
			})
		}

		return h.loginOAuthUser(c, *stateEntry.UserID)
	}

	// Scenario C: Purpose is "login" and no link → create new user + OAuth link
	// Generate nickname from Discord username
	nickname := discordUser.GlobalName
	if nickname == "" {
		nickname = discordUser.Username
	}
	// Ensure nickname length constraints (2-20 chars)
	if len(nickname) < 2 {
		nickname = nickname + "_user"
	}
	if len(nickname) > 20 {
		nickname = nickname[:20]
	}

	// Check if nickname is taken, append random suffix if so
	nicknameExists, err := h.userRepo.ExistsByNickname(ctx, nickname)
	if err != nil {
		slog.Error("DiscordCallback: failed to check nickname", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}
	if nicknameExists {
		suffix := fmt.Sprintf("%04d", rand.Intn(10000))
		if len(nickname)+len(suffix) > 20 {
			nickname = nickname[:20-len(suffix)]
		}
		nickname = nickname + suffix
	}

	// Use Discord email
	email := discordUser.Email
	if email == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "email_required",
			Message: "Discord 계정에 이메일이 설정되어 있지 않습니다",
		})
	}

	// Check if email already exists (DON'T auto-link for security)
	emailExists, err := h.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		slog.Error("DiscordCallback: failed to check email", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}
	if emailExists {
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "email_exists",
			Message: "이미 사용 중인 이메일입니다. 기존 계정에 로그인 후 Discord를 연동해주세요.",
		})
	}

	// Create new user without password
	newUser := &model.User{
		Email:    email,
		Nickname: nickname,
	}
	if err := h.userRepo.CreateOAuthUser(ctx, newUser); err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "email_exists",
				Message: "이미 사용 중인 이메일입니다. 기존 계정에 로그인 후 Discord를 연동해주세요.",
			})
		}
		if errors.Is(err, repository.ErrNicknameExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "nickname_exists",
				Message: "이미 사용 중인 닉네임입니다",
			})
		}
		slog.Error("DiscordCallback: failed to create OAuth user", "error", err, "email", email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Create OAuth account link
	oauthAccount := &model.OAuthAccount{
		UserID:           newUser.ID,
		Provider:         "discord",
		ProviderID:       discordUser.ID,
		ProviderUsername: discordUser.Username,
		ProviderAvatar:   discordUser.Avatar,
		ProviderEmail:    discordUser.Email,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		TokenExpiresAt:   &tokenExpiresAt,
	}

	if err := h.oauthRepo.Create(ctx, oauthAccount); err != nil {
		slog.Error("DiscordCallback: failed to create OAuth link for new user", "error", err, "userID", newUser.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	return h.loginOAuthUser(c, newUser.ID)
}

// DiscordLink handles GET /api/v1/auth/discord/link (requires auth)
// Generates a state token for linking Discord to an existing account
func (h *AuthHandler) DiscordLink(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	state := h.oauthState.Generate("link", &userID)

	return c.JSON(http.StatusOK, map[string]string{
		"url": h.discordService.GetAuthorizationURL(state),
	})
}

// DiscordUnlink handles DELETE /api/v1/auth/discord/link (requires auth)
// Removes the Discord link from the user's account
func (h *AuthHandler) DiscordUnlink(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	// Check if user has a password set (can't unlink if no password - would lock out)
	hasPassword, err := h.userRepo.HasPassword(ctx, userID)
	if err != nil {
		slog.Error("DiscordUnlink: failed to check password", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}
	if !hasPassword {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "cannot_unlink",
			Message: "비밀번호가 설정되지 않은 계정은 소셜 로그인 연동을 해제할 수 없습니다",
		})
	}

	if err := h.oauthRepo.DeleteByUserAndProvider(ctx, userID, "discord"); err != nil {
		if errors.Is(err, repository.ErrOAuthAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_linked",
				Message: "Discord 계정이 연동되어 있지 않습니다",
			})
		}
		slog.Error("DiscordUnlink: failed to delete OAuth link", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Discord 연동이 해제되었습니다",
	})
}

// GetLinkedAccounts handles GET /api/v1/auth/linked-accounts (requires auth)
// Returns the list of linked OAuth accounts for the current user
func (h *AuthHandler) GetLinkedAccounts(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	accounts, err := h.oauthRepo.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("GetLinkedAccounts: failed to list OAuth accounts", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Convert to OAuthLinkStatus array
	statuses := make([]model.OAuthLinkStatus, 0)

	// Always include Discord status
	discordStatus := model.OAuthLinkStatus{
		Provider: "discord",
		Linked:   false,
	}

	for _, account := range accounts {
		if account.Provider == "discord" {
			discordStatus.Linked = true
			discordStatus.ProviderUsername = account.ProviderUsername
			discordStatus.ProviderAvatar = account.ProviderAvatar
			discordStatus.LinkedAt = &account.CreatedAt
		}
	}

	statuses = append(statuses, discordStatus)

	return c.JSON(http.StatusOK, statuses)
}

// loginOAuthUser generates tokens and returns a login response for the given user
func (h *AuthHandler) loginOAuthUser(c echo.Context, userID uuid.UUID) error {
	ctx := c.Request().Context()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("loginOAuthUser: failed to get user", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.Nickname, user.Role, user.Permissions)
	if err != nil {
		slog.Error("loginOAuthUser: failed to generate access token", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.Error("loginOAuthUser: failed to generate refresh token", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	// Save refresh token
	if h.refreshTokenRepo != nil {
		deviceInfo := c.Request().UserAgent()
		if len(deviceInfo) > 255 {
			deviceInfo = deviceInfo[:255]
		}
		ipAddress := c.RealIP()

		rt := &model.RefreshToken{
			UserID:     user.ID,
			Token:      refreshToken,
			DeviceInfo: &deviceInfo,
			IPAddress:  &ipAddress,
			ExpiresAt:  time.Now().Add(h.jwtService.RefreshExpiry()),
		}
		if err := h.refreshTokenRepo.Create(ctx, rt); err != nil {
			slog.Error("loginOAuthUser: failed to create refresh token", "error", err, "userID", userID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "서버 오류가 발생했습니다",
			})
		}
	} else {
		if err := h.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
			slog.Error("loginOAuthUser: failed to update refresh token", "error", err, "userID", userID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "서버 오류가 발생했습니다",
			})
		}
	}

	// Set httpOnly cookie
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.jwtService.RefreshExpiry().Seconds()),
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken: accessToken,
		User:        user,
	})
}
