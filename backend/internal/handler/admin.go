package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/auth"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminHandler handles admin requests
type AdminHandler struct {
	userRepo    *repository.UserRepository
	historyRepo *repository.PermissionHistoryRepository
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(userRepo *repository.UserRepository, historyRepo *repository.PermissionHistoryRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		historyRepo: historyRepo,
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
	roleFilter := c.QueryParam("role")

	ctx := c.Request().Context()

	var users []*model.User
	var total int
	var err error

	// Use roleFilter if provided
	if roleFilter != "" {
		users, total, err = h.userRepo.ListUsers(ctx, page, limit, search, roleFilter)
	} else {
		users, total, err = h.userRepo.ListUsersSimple(ctx, page, limit, search)
	}

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

// GetUser handles GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "유저 ID가 필요합니다",
		})
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "유효하지 않은 유저 ID입니다",
		})
	}

	ctx := c.Request().Context()

	user, err := h.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "유저를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "유저 정보를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, user)
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

	roleStats, err := h.userRepo.CountUsersByRole(ctx)
	if err != nil {
		// Non-critical error, just log and continue
		roleStats = make(map[string]int)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_users":    totalUsers,
		"users_by_role":  roleStats,
	})
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

	var req model.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate role
	role := auth.Role(req.Role)
	if !role.IsValid() {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_role",
			Message: "유효하지 않은 역할입니다. USER, STAFF, ADMIN 중 하나여야 합니다",
		})
	}

	ctx := c.Request().Context()

	// Parse UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "유효하지 않은 유저 ID입니다",
		})
	}

	// Get current user for history
	targetUser, err := h.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "유저를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "유저 정보를 불러오는데 실패했습니다",
		})
	}

	// Prevent downgrading ADMIN if they are the only one
	if targetUser.Role == "ADMIN" && req.Role != "ADMIN" {
		roleStats, _ := h.userRepo.CountUsersByRole(ctx)
		if roleStats["ADMIN"] <= 1 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "last_admin",
				Message: "마지막 관리자는 역할을 변경할 수 없습니다",
			})
		}
	}

	// Check version for optimistic locking
	if req.Version != targetUser.Version {
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "version_conflict",
			Message: "다른 관리자가 이 유저를 수정 중입니다. 페이지를 새로고침해주세요",
		})
	}

	oldRole := targetUser.Role

	// Update role with optimistic locking
	newVersion, err := h.userRepo.UpdateUserRole(ctx, uid, req.Role, req.Version)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "version_conflict",
				Message: "다른 관리자가 이 유저를 수정 중입니다. 페이지를 새로고침해주세요",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "역할 변경에 실패했습니다",
		})
	}

	// Record history
	changerID := c.Get("user_id").(uuid.UUID)
	history := &model.PermissionHistory{
		ChangerID:  changerID,
		TargetID:   uid,
		ChangeType: "ROLE",
		OldValue:   oldRole,
		NewValue:   req.Role,
	}
	if err := h.historyRepo.Create(ctx, history); err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "역할이 변경되었습니다",
		"new_version": newVersion,
	})
}

// UpdateUserPermissions handles PUT /api/v1/admin/users/:id/permissions
func (h *AdminHandler) UpdateUserPermissions(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "유저 ID가 필요합니다",
		})
	}

	var req model.UpdatePermissionsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate all permissions
	for _, p := range req.Permissions {
		perm := auth.Permission(p)
		if !perm.IsValid() {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_permission",
				Message: "유효하지 않은 권한입니다: " + p,
			})
		}
	}

	ctx := c.Request().Context()

	// Parse UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "유효하지 않은 유저 ID입니다",
		})
	}

	// Get current user for history
	targetUser, err := h.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "유저를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "유저 정보를 불러오는데 실패했습니다",
		})
	}

	// Check version for optimistic locking
	if req.Version != targetUser.Version {
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "version_conflict",
			Message: "다른 관리자가 이 유저를 수정 중입니다. 페이지를 새로고침해주세요",
		})
	}

	oldPermissions := targetUser.Permissions

	// Update permissions with optimistic locking
	newVersion, err := h.userRepo.UpdateUserPermissions(ctx, uid, req.Permissions, req.Version)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "version_conflict",
				Message: "다른 관리자가 이 유저를 수정 중입니다. 페이지를 새로고침해주세요",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "권한 변경에 실패했습니다",
		})
	}

	// Record history
	changerID := c.Get("user_id").(uuid.UUID)
	history := &model.PermissionHistory{
		ChangerID:  changerID,
		TargetID:   uid,
		ChangeType: "PERMISSION",
		OldValue:   oldPermissions,
		NewValue:   req.Permissions,
	}
	if err := h.historyRepo.Create(ctx, history); err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "권한이 변경되었습니다",
		"new_version": newVersion,
	})
}

// PermissionHistoryResponse represents permission history response
type PermissionHistoryResponse struct {
	History    []*model.PermissionHistory `json:"history"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// GetUserPermissionHistory handles GET /api/v1/admin/users/:id/history
func (h *AdminHandler) GetUserPermissionHistory(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "유저 ID가 필요합니다",
		})
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "유효하지 않은 유저 ID입니다",
		})
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request().Context()

	history, total, err := h.historyRepo.GetByTargetID(ctx, uid, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "권한 변경 기록을 불러오는데 실패했습니다",
		})
	}

	if history == nil {
		history = []*model.PermissionHistory{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, PermissionHistoryResponse{
		History:    history,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetRecentPermissionHistory handles GET /api/v1/admin/permission-history
func (h *AdminHandler) GetRecentPermissionHistory(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request().Context()

	history, err := h.historyRepo.GetRecentHistory(ctx, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "권한 변경 기록을 불러오는데 실패했습니다",
		})
	}

	if history == nil {
		history = []*model.PermissionHistory{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

// GetPermissionsList handles GET /api/v1/admin/permissions
func (h *AdminHandler) GetPermissionsList(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"permissions": auth.GetPermissionInfo(),
		"roles": []map[string]string{
			{"code": "USER", "name": "일반 유저", "description": "기본 사용자 권한"},
			{"code": "STAFF", "name": "스태프", "description": "권한에 따른 관리 기능 접근"},
			{"code": "ADMIN", "name": "관리자", "description": "모든 권한 보유"},
		},
	})
}
