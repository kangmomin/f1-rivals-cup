package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MatchHandler struct {
	matchSvc *service.MatchService
}

func NewMatchHandler(matchSvc *service.MatchService) *MatchHandler {
	return &MatchHandler{
		matchSvc: matchSvc,
	}
}

// Create handles POST /api/v1/admin/leagues/:id/matches
func (h *MatchHandler) Create(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	var req model.CreateMatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	match, err := h.matchSvc.Create(ctx, leagueID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidMatchRequest) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "라운드, 트랙, 날짜는 필수입니다",
			})
		}
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		if errors.Is(err, service.ErrDuplicateRound) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "duplicate_round",
				Message: "이미 해당 라운드가 존재합니다",
			})
		}
		slog.Error("Match.Create: failed to create match", "error", err, "league_id", leagueID, "round", req.Round)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 일정 생성에 실패했습니다",
		})
	}

	return c.JSON(http.StatusCreated, match)
}

// List handles GET /api/v1/leagues/:id/matches
func (h *MatchHandler) List(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	matches, err := h.matchSvc.List(ctx, leagueID)
	if err != nil {
		slog.Error("Match.List: failed to list matches", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 일정을 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.ListMatchesResponse{
		Matches: matches,
		Total:   len(matches),
	})
}

// Get handles GET /api/v1/matches/:id
func (h *MatchHandler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	ctx := c.Request().Context()

	match, err := h.matchSvc.Get(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		slog.Error("Match.Get: failed to get match", "error", err, "match_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, match)
}

// Update handles PUT /api/v1/admin/matches/:id
func (h *MatchHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	var req model.UpdateMatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	match, err := h.matchSvc.Update(ctx, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		if errors.Is(err, service.ErrInvalidMatchRequest) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "잘못된 요청입니다",
			})
		}
		if errors.Is(err, service.ErrDuplicateRound) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "duplicate_round",
				Message: "이미 해당 라운드가 존재합니다",
			})
		}
		slog.Error("Match.Update: failed to update match", "error", err, "match_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보 수정에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, match)
}

// Delete handles DELETE /api/v1/admin/matches/:id
func (h *MatchHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.matchSvc.Delete(ctx, id); err != nil {
		if errors.Is(err, service.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		slog.Error("Match.Delete: failed to delete match", "error", err, "match_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "경기가 삭제되었습니다",
	})
}
