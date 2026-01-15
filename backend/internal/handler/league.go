package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LeagueHandler handles league requests
type LeagueHandler struct {
	leagueSvc *service.LeagueService
}

// NewLeagueHandler creates a new LeagueHandler
func NewLeagueHandler(leagueSvc *service.LeagueService) *LeagueHandler {
	return &LeagueHandler{
		leagueSvc: leagueSvc,
	}
}

// Create handles POST /api/v1/admin/leagues
func (h *LeagueHandler) Create(c echo.Context) error {
	var req model.CreateLeagueRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Get user ID from context
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	league, err := h.leagueSvc.Create(ctx, &req, userID)
	if err != nil {
		// Handle validation errors from service
		if errors.Is(err, service.ErrNilRequest) ||
			errors.Is(err, service.ErrInvalidDateFormat) ||
			errors.Is(err, service.ErrInvalidLeagueName) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
		}
		// Check for validation error messages (Korean error messages from service)
		errMsg := err.Error()
		if errMsg == "리그 이름을 입력해주세요" ||
			errMsg == "리그 이름은 최소 2자 이상이어야 합니다" ||
			errMsg == "리그 이름은 최대 100자까지 가능합니다" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: errMsg,
			})
		}

		slog.Error("League.Create: failed to create league", "error", err, "name", req.Name)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 생성에 실패했습니다",
		})
	}

	return c.JSON(http.StatusCreated, league)
}

// List handles GET /api/v1/admin/leagues
func (h *LeagueHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	status := c.QueryParam("status")

	ctx := c.Request().Context()

	result, err := h.leagueSvc.List(ctx, page, limit, status)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLeagueStatus) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "잘못된 리그 상태입니다",
			})
		}
		slog.Error("League.List: failed to list leagues", "error", err, "page", page, "limit", limit, "status", status)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 목록을 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.ListLeaguesResponse{
		Leagues:    result.Leagues,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.PageSize,
		TotalPages: result.TotalPages,
	})
}

// Get handles GET /api/v1/admin/leagues/:id
func (h *LeagueHandler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	league, err := h.leagueSvc.Get(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("League.Get: failed to get league", "error", err, "id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, league)
}

// Update handles PUT /api/v1/admin/leagues/:id
func (h *LeagueHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	var req model.UpdateLeagueRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	league, err := h.leagueSvc.Update(ctx, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		// Handle validation errors from service
		if errors.Is(err, service.ErrNilRequest) ||
			errors.Is(err, service.ErrInvalidDateFormat) ||
			errors.Is(err, service.ErrInvalidLeagueStatus) ||
			errors.Is(err, service.ErrInvalidLeagueName) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
		}
		// Check for validation error messages (Korean error messages from service)
		errMsg := err.Error()
		if errMsg == "리그 이름을 입력해주세요" ||
			errMsg == "리그 이름은 최소 2자 이상이어야 합니다" ||
			errMsg == "리그 이름은 최대 100자까지 가능합니다" ||
			errMsg == "시즌은 1 이상이어야 합니다" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: errMsg,
			})
		}

		slog.Error("League.Update: failed to update league", "error", err, "id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 수정에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, league)
}

// Delete handles DELETE /api/v1/admin/leagues/:id
func (h *LeagueHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.leagueSvc.Delete(ctx, id); err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("League.Delete: failed to delete league", "error", err, "id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "리그가 삭제되었습니다",
	})
}
