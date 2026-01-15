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

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// List handles GET /api/v1/leagues/:id/teams
func (h *TeamHandler) List(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	teams, err := h.teamService.List(ctx, leagueID)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Team.List: failed to list teams", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 목록을 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.ListTeamsResponse{
		Teams: teams,
		Total: len(teams),
	})
}

// Create handles POST /api/v1/admin/leagues/:id/teams
func (h *TeamHandler) Create(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	var req model.CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	team, err := h.teamService.Create(ctx, leagueID, &req)
	if err != nil {
		if errors.Is(err, service.ErrTeamNameRequired) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "팀 이름을 입력해주세요",
			})
		}
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		if errors.Is(err, service.ErrTeamAlreadyExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "conflict",
				Message: "이미 같은 이름의 팀이 있습니다",
			})
		}
		slog.Error("Team.Create: failed to create team", "error", err, "league_id", leagueID, "name", req.Name)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 생성에 실패했습니다",
		})
	}

	return c.JSON(http.StatusCreated, team)
}

// Update handles PUT /api/v1/admin/teams/:id
func (h *TeamHandler) Update(c echo.Context) error {
	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 팀 ID입니다",
		})
	}

	var req model.UpdateTeamRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	team, err := h.teamService.Update(ctx, teamID, &req)
	if err != nil {
		if errors.Is(err, service.ErrTeamNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "팀을 찾을 수 없습니다",
			})
		}
		if errors.Is(err, service.ErrTeamNameRequired) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "팀 이름을 입력해주세요",
			})
		}
		if errors.Is(err, service.ErrTeamAlreadyExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "conflict",
				Message: "이미 같은 이름의 팀이 있습니다",
			})
		}
		slog.Error("Team.Update: failed to update team", "error", err, "team_id", teamID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 수정에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, team)
}

// Delete handles DELETE /api/v1/admin/teams/:id
func (h *TeamHandler) Delete(c echo.Context) error {
	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 팀 ID입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.teamService.Delete(ctx, teamID); err != nil {
		if errors.Is(err, service.ErrTeamNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "팀을 찾을 수 없습니다",
			})
		}
		slog.Error("Team.Delete: failed to delete team", "error", err, "team_id", teamID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "팀이 삭제되었습니다",
	})
}
