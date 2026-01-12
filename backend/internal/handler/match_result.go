package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MatchResultHandler struct {
	resultRepo *repository.MatchResultRepository
	matchRepo  *repository.MatchRepository
	leagueRepo *repository.LeagueRepository
}

func NewMatchResultHandler(resultRepo *repository.MatchResultRepository, matchRepo *repository.MatchRepository, leagueRepo *repository.LeagueRepository) *MatchResultHandler {
	return &MatchResultHandler{
		resultRepo: resultRepo,
		matchRepo:  matchRepo,
		leagueRepo: leagueRepo,
	}
}

// List handles GET /api/v1/matches/:id/results
func (h *MatchResultHandler) List(c echo.Context) error {
	matchIDStr := c.Param("id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if match exists
	_, err = h.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		slog.Error("MatchResult.List: failed to get match", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보를 불러오는데 실패했습니다",
		})
	}

	results, err := h.resultRepo.ListByMatch(ctx, matchID)
	if err != nil {
		slog.Error("MatchResult.List: failed to list results", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 결과를 불러오는데 실패했습니다",
		})
	}

	if results == nil {
		results = []*model.MatchResult{}
	}

	return c.JSON(http.StatusOK, model.ListMatchResultsResponse{
		Results: results,
		Total:   len(results),
	})
}

// BulkUpdate handles PUT /api/v1/admin/matches/:id/results
func (h *MatchResultHandler) BulkUpdate(c echo.Context) error {
	matchIDStr := c.Param("id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	var req model.BulkUpdateResultsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if match exists
	match, err := h.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		slog.Error("MatchResult.BulkUpdate: failed to get match", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보를 불러오는데 실패했습니다",
		})
	}

	// Bulk upsert results
	if err := h.resultRepo.BulkUpsert(ctx, matchID, req.Results); err != nil {
		slog.Error("MatchResult.BulkUpdate: failed to bulk upsert results", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 결과 저장에 실패했습니다",
		})
	}

	// Update match status to completed if not already
	if match.Status != model.MatchStatusCompleted {
		match.Status = model.MatchStatusCompleted
		if err := h.matchRepo.Update(ctx, match); err != nil {
			slog.Error("MatchResult.BulkUpdate: failed to update match status", "error", err, "match_id", matchID)
		}
	}

	// Return updated results
	results, err := h.resultRepo.ListByMatch(ctx, matchID)
	if err != nil {
		slog.Error("MatchResult.BulkUpdate: failed to list results", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 결과를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.ListMatchResultsResponse{
		Results: results,
		Total:   len(results),
	})
}

// Delete handles DELETE /api/v1/admin/matches/:id/results
func (h *MatchResultHandler) Delete(c echo.Context) error {
	matchIDStr := c.Param("id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 경기 ID입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.resultRepo.DeleteByMatch(ctx, matchID); err != nil {
		slog.Error("MatchResult.Delete: failed to delete results", "error", err, "match_id", matchID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 결과 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "경기 결과가 삭제되었습니다",
	})
}

// Standings handles GET /api/v1/leagues/:id/standings
func (h *MatchResultHandler) Standings(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Get league info
	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("MatchResult.Standings: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	// Get total races count
	matches, err := h.matchRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		slog.Error("MatchResult.Standings: failed to list matches", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보를 불러오는데 실패했습니다",
		})
	}

	// Get driver standings
	standings, err := h.resultRepo.GetLeagueStandings(ctx, leagueID)
	if err != nil {
		slog.Error("MatchResult.Standings: failed to get driver standings", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "순위 정보를 불러오는데 실패했습니다",
		})
	}

	// Get team standings
	teamStandings, err := h.resultRepo.GetTeamStandings(ctx, leagueID)
	if err != nil {
		slog.Error("MatchResult.Standings: failed to get team standings", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 순위 정보를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.LeagueStandingsResponse{
		LeagueID:      leagueID,
		LeagueName:    league.Name,
		Season:        league.Season,
		TotalRaces:    len(matches),
		Standings:     standings,
		TeamStandings: teamStandings,
	})
}
