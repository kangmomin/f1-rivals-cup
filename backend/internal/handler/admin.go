package handler

import (
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
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
