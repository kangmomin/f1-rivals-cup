package handler

import (
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminHandler handles admin requests
type AdminHandler struct {
	userRepo *repository.UserRepository
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
	}
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []*model.User `json:"users"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c echo.Context) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	search := c.QueryParam("search")

	ctx := c.Request().Context()

	users, total, err := h.userRepo.ListUsers(ctx, page, limit, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "회원 목록을 불러오는데 실패했습니다",
		})
	}

	// Ensure users is not nil
	if users == nil {
		users = []*model.User{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetUserStats handles GET /api/v1/admin/stats
func (h *AdminHandler) GetUserStats(c echo.Context) error {
	ctx := c.Request().Context()

	totalUsers, err := h.userRepo.CountUsers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "통계를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_users": totalUsers,
	})
}

// UpdateUserRoleRequest represents the request to update user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}

// UpdateUserRole handles PUT /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "유저 ID가 필요합니다",
		})
	}

	var req UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate role
	if req.Role != "user" && req.Role != "admin" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_role",
			Message: "유효하지 않은 권한입니다. user 또는 admin만 가능합니다",
		})
	}

	ctx := c.Request().Context()

	// Parse UUID
	uid, err := parseUUID(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "유효하지 않은 유저 ID입니다",
		})
	}

	// Update role
	if err := h.userRepo.UpdateUserRole(ctx, uid, req.Role); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "권한 변경에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "권한이 변경되었습니다",
	})
}

// parseUUID parses a string to UUID
func parseUUID(s string) ([16]byte, error) {
	var uid [16]byte
	parsed, err := uuid.Parse(s)
	if err != nil {
		return uid, err
	}
	return parsed, nil
}
