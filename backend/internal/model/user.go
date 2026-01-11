package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                   uuid.UUID  `json:"id"`
	Email                string     `json:"email"`
	PasswordHash         string     `json:"-"`
	Nickname             string     `json:"nickname"`
	Role                 string     `json:"role"`
	Permissions          []string   `json:"permissions"`
	Version              int        `json:"-"` // For optimistic locking
	EmailVerified        bool       `json:"email_verified"`
	EmailVerifyToken     *string    `json:"-"`
	PasswordResetToken   *string    `json:"-"`
	PasswordResetExpires *time.Time `json:"-"`
	RefreshToken         *string    `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PermissionHistory represents a permission change record
type PermissionHistory struct {
	ID         uuid.UUID `json:"id"`
	ChangerID  uuid.UUID `json:"changer_id"`
	TargetID   uuid.UUID `json:"target_id"`
	ChangeType string    `json:"change_type"` // ROLE or PERMISSION
	OldValue   any       `json:"old_value"`
	NewValue   any       `json:"new_value"`
	CreatedAt  time.Time `json:"created_at"`
	// Joined fields
	ChangerNickname string `json:"changer_nickname,omitempty"`
	TargetNickname  string `json:"target_nickname,omitempty"`
}

// UpdateRoleRequest represents a request to update user role
type UpdateRoleRequest struct {
	Role    string `json:"role" validate:"required,oneof=USER STAFF ADMIN"`
	Version int    `json:"version" validate:"required"`
}

// UpdatePermissionsRequest represents a request to update user permissions
type UpdatePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required"`
	Version     int      `json:"version" validate:"required"`
}

// PermissionErrorResponse represents an error response with permission details
type PermissionErrorResponse struct {
	Error struct {
		Code               string   `json:"code"`
		Message            string   `json:"message"`
		RequiredPermission string   `json:"required_permission,omitempty"`
		Details            *struct {
			UserRole        string   `json:"user_role"`
			UserPermissions []string `json:"user_permissions"`
		} `json:"details,omitempty"`
	} `json:"error"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Nickname string `json:"nickname" validate:"required,min=2,max=20"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	Message string `json:"message"`
	User    *User  `json:"user"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation request
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
