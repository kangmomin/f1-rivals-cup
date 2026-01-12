package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/f1-rivals-cup/backend/internal/auth"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate request
	if err := validateRegisterRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check if email exists
	exists, err := h.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		slog.Error("Register: ExistsByEmail failed", "email", req.Email, "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}
	if exists {
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "email_exists",
			Message: "이미 사용 중인 이메일입니다",
		})
	}

	// Check if nickname exists
	exists, err = h.userRepo.ExistsByNickname(ctx, req.Nickname)
	if err != nil {
		slog.Error("Register: ExistsByNickname failed", "nickname", req.Nickname, "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}
	if exists {
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "nickname_exists",
			Message: "이미 사용 중인 닉네임입니다",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Register: password hashing failed", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Generate email verification token
	verifyToken, err := generateToken(32)
	if err != nil {
		slog.Error("Register: token generation failed", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Create user
	user := &model.User{
		Email:            req.Email,
		PasswordHash:     string(hashedPassword),
		Nickname:         req.Nickname,
		EmailVerifyToken: &verifyToken,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "email_exists",
				Message: "이미 사용 중인 이메일입니다",
			})
		}
		if errors.Is(err, repository.ErrNicknameExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "nickname_exists",
				Message: "이미 사용 중인 닉네임입니다",
			})
		}
		slog.Error("Register: user creation failed", "email", req.Email, "nickname", req.Nickname, "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "회원가입에 실패했습니다",
		})
	}

	// TODO: Send verification email

	return c.JSON(http.StatusCreated, model.RegisterResponse{
		Message: "회원가입이 완료되었습니다. 이메일을 확인해주세요.",
		User:    user,
	})
}

// validateRegisterRequest validates the registration request
func validateRegisterRequest(req *model.RegisterRequest) error {
	req.Email = strings.TrimSpace(req.Email)
	req.Nickname = strings.TrimSpace(req.Nickname)

	if req.Email == "" {
		return errors.New("이메일을 입력해주세요")
	}
	if !strings.Contains(req.Email, "@") {
		return errors.New("유효한 이메일을 입력해주세요")
	}
	if req.Nickname == "" {
		return errors.New("닉네임을 입력해주세요")
	}
	if len(req.Nickname) < 2 {
		return errors.New("닉네임은 최소 2자 이상이어야 합니다")
	}
	if len(req.Nickname) > 20 {
		return errors.New("닉네임은 최대 20자까지 가능합니다")
	}
	if req.Password == "" {
		return errors.New("비밀번호를 입력해주세요")
	}
	if len(req.Password) < 8 {
		return errors.New("비밀번호는 최소 8자 이상이어야 합니다")
	}

	return nil
}

// generateToken generates a random hex token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate access token
	claims, err := h.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "invalid_token",
			Message: "유효하지 않은 토큰입니다",
		})
	}

	ctx := c.Request().Context()

	// Clear refresh token from database
	if err := h.userRepo.ClearRefreshToken(ctx, claims.UserID); err != nil {
		slog.Error("Auth.Logout: failed to clear refresh token", "error", err, "userID", claims.UserID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "로그아웃되었습니다",
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req model.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "리프레시 토큰을 입력해주세요",
		})
	}

	// Validate refresh token
	userID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "invalid_token",
			Message: "유효하지 않은 토큰입니다",
		})
	}

	ctx := c.Request().Context()

	// Get user from database
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error:   "invalid_token",
				Message: "유효하지 않은 토큰입니다",
			})
		}
		slog.Error("Auth.RefreshToken: failed to get user by ID", "error", err, "userID", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Verify refresh token matches stored token
	if user.RefreshToken == nil || *user.RefreshToken != req.RefreshToken {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "invalid_token",
			Message: "유효하지 않은 토큰입니다",
		})
	}

	// Generate new tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.Nickname, user.Role, user.Permissions)
	if err != nil {
		slog.Error("Auth.RefreshToken: failed to generate access token", "error", err, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	newRefreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.Error("Auth.RefreshToken: failed to generate refresh token", "error", err, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	// Save new refresh token
	if err := h.userRepo.UpdateRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		slog.Error("Auth.RefreshToken: failed to update refresh token", "error", err, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate request
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "이메일과 비밀번호를 입력해주세요",
		})
	}

	ctx := c.Request().Context()

	// Get user by email
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "이메일 또는 비밀번호가 올바르지 않습니다",
			})
		}
		slog.Error("Auth.Login: failed to get user by email", "error", err, "email", req.Email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "이메일 또는 비밀번호가 올바르지 않습니다",
		})
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.Nickname, user.Role, user.Permissions)
	if err != nil {
		slog.Error("Auth.Login: failed to generate access token", "error", err, "email", req.Email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.Error("Auth.Login: failed to generate refresh token", "error", err, "email", req.Email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "토큰 생성에 실패했습니다",
		})
	}

	// Save refresh token
	if err := h.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		slog.Error("Auth.Login: failed to update refresh token", "error", err, "email", req.Email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// RequestPasswordReset handles POST /api/v1/auth/password-reset
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req model.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "이메일을 입력해주세요",
		})
	}

	ctx := c.Request().Context()

	// Get user by email
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return c.JSON(http.StatusOK, map[string]string{
			"message": "비밀번호 재설정 링크가 이메일로 전송되었습니다",
		})
	}

	// Generate reset token
	resetToken, err := generateToken(32)
	if err != nil {
		slog.Error("Auth.RequestPasswordReset: failed to generate reset token", "error", err, "email", req.Email)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Set token expiry to 1 hour
	expires := time.Now().Add(1 * time.Hour)

	if err := h.userRepo.SetPasswordResetToken(ctx, user.ID, resetToken, expires); err != nil {
		slog.Error("Auth.RequestPasswordReset: failed to set password reset token", "error", err, "email", req.Email, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// TODO: Send password reset email with resetToken

	return c.JSON(http.StatusOK, map[string]string{
		"message": "비밀번호 재설정 링크가 이메일로 전송되었습니다",
	})
}

// ConfirmPasswordReset handles POST /api/v1/auth/password-reset/confirm
func (h *AuthHandler) ConfirmPasswordReset(c echo.Context) error {
	var req model.PasswordResetConfirmRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "토큰을 입력해주세요",
		})
	}

	if len(req.NewPassword) < 8 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "비밀번호는 최소 8자 이상이어야 합니다",
		})
	}

	ctx := c.Request().Context()

	// Get user by reset token
	user, err := h.userRepo.GetByPasswordResetToken(ctx, req.Token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_token",
			Message: "유효하지 않거나 만료된 토큰입니다",
		})
	}

	// Check if token is expired
	if user.PasswordResetExpires == nil || time.Now().After(*user.PasswordResetExpires) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "token_expired",
			Message: "토큰이 만료되었습니다. 다시 요청해주세요",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Auth.ConfirmPasswordReset: failed to hash new password", "error", err, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "서버 오류가 발생했습니다",
		})
	}

	// Update password
	if err := h.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		slog.Error("Auth.ConfirmPasswordReset: failed to update password", "error", err, "userID", user.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "비밀번호 변경에 실패했습니다",
		})
	}

	// Clear all refresh tokens to force re-login
	if err := h.userRepo.ClearRefreshToken(ctx, user.ID); err != nil {
		// Log error but don't fail the request
		slog.Error("Auth.ConfirmPasswordReset: failed to clear refresh token", "error", err, "userID", user.ID)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "비밀번호가 성공적으로 변경되었습니다",
	})
}
