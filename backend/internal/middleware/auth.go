package middleware

import (
	"net/http"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/auth"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtService *auth.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error:   "unauthorized",
					Message: "인증이 필요합니다",
				})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error:   "invalid_token",
					Message: "유효하지 않은 토큰입니다",
				})
			}

			// Store claims in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("nickname", claims.Nickname)
			c.Set("role", claims.Role)
			c.Set("permissions", claims.Permissions)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware extracts user info if token is provided, but doesn't require it
func OptionalAuthMiddleware(jwtService *auth.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := jwtService.ValidateAccessToken(tokenString)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
					c.Set("nickname", claims.Nickname)
					c.Set("role", claims.Role)
					c.Set("permissions", claims.Permissions)
				}
			}
			return next(c)
		}
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...auth.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role")
			if userRole == nil {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error:   "unauthorized",
					Message: "인증이 필요합니다",
				})
			}

			role := auth.Role(userRole.(string))

			// ADMIN has access to everything
			if role == auth.RoleAdmin {
				return next(c)
			}

			// Check if user has one of the required roles
			for _, requiredRole := range roles {
				if role == requiredRole {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, model.PermissionErrorResponse{
				Error: struct {
					Code               string   `json:"code"`
					Message            string   `json:"message"`
					RequiredPermission string   `json:"required_permission,omitempty"`
					Details            *struct {
						UserRole        string   `json:"user_role"`
						UserPermissions []string `json:"user_permissions"`
					} `json:"details,omitempty"`
				}{
					Code:    "insufficient_role",
					Message: "이 작업을 수행할 권한이 없습니다",
				},
			})
		}
	}
}

// RequirePermission creates a middleware that requires specific permissions
func RequirePermission(required ...auth.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role")
			userPerms := c.Get("permissions")

			if userRole == nil {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error:   "unauthorized",
					Message: "인증이 필요합니다",
				})
			}

			role := auth.Role(userRole.(string))

			// ADMIN has all permissions (wildcard)
			if role == auth.RoleAdmin {
				return next(c)
			}

			// Get user permissions
			var permissions []string
			if userPerms != nil {
				permissions = userPerms.([]string)
			}

			// Check if user has any of the required permissions
			if auth.HasAnyPermission(permissions, required) {
				return next(c)
			}

			// Build permission names for error message
			permNames := make([]string, len(required))
			for i, p := range required {
				permNames[i] = string(p)
			}

			return c.JSON(http.StatusForbidden, model.PermissionErrorResponse{
				Error: struct {
					Code               string   `json:"code"`
					Message            string   `json:"message"`
					RequiredPermission string   `json:"required_permission,omitempty"`
					Details            *struct {
						UserRole        string   `json:"user_role"`
						UserPermissions []string `json:"user_permissions"`
					} `json:"details,omitempty"`
				}{
					Code:               "insufficient_permission",
					Message:            "이 작업을 수행할 권한이 없습니다",
					RequiredPermission: strings.Join(permNames, ", "),
					Details: &struct {
						UserRole        string   `json:"user_role"`
						UserPermissions []string `json:"user_permissions"`
					}{
						UserRole:        string(role),
						UserPermissions: permissions,
					},
				},
			})
		}
	}
}

// RequireAllPermissions creates a middleware that requires ALL specified permissions
func RequireAllPermissions(required ...auth.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role")
			userPerms := c.Get("permissions")

			if userRole == nil {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error:   "unauthorized",
					Message: "인증이 필요합니다",
				})
			}

			role := auth.Role(userRole.(string))

			// ADMIN has all permissions (wildcard)
			if role == auth.RoleAdmin {
				return next(c)
			}

			// Get user permissions
			var permissions []string
			if userPerms != nil {
				permissions = userPerms.([]string)
			}

			// Check if user has all required permissions
			if auth.HasAllPermissions(permissions, required) {
				return next(c)
			}

			// Build permission names for error message
			permNames := make([]string, len(required))
			for i, p := range required {
				permNames[i] = string(p)
			}

			return c.JSON(http.StatusForbidden, model.PermissionErrorResponse{
				Error: struct {
					Code               string   `json:"code"`
					Message            string   `json:"message"`
					RequiredPermission string   `json:"required_permission,omitempty"`
					Details            *struct {
						UserRole        string   `json:"user_role"`
						UserPermissions []string `json:"user_permissions"`
					} `json:"details,omitempty"`
				}{
					Code:               "insufficient_permission",
					Message:            "이 작업을 수행할 권한이 없습니다",
					RequiredPermission: strings.Join(permNames, ", "),
					Details: &struct {
						UserRole        string   `json:"user_role"`
						UserPermissions []string `json:"user_permissions"`
					}{
						UserRole:        string(role),
						UserPermissions: permissions,
					},
				},
			})
		}
	}
}
