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
				}
			}
			return next(c)
		}
	}
}
