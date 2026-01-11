package handler

import (
	"errors"
	"net/http"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MatchHandler struct {
	matchRepo  *repository.MatchRepository
	leagueRepo *repository.LeagueRepository
}

func NewMatchHandler(matchRepo *repository.MatchRepository, leagueRepo *repository.LeagueRepository) *MatchHandler {
	return &MatchHandler{
		matchRepo:  matchRepo,
		leagueRepo: leagueRepo,
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

	if req.Track == "" || req.MatchDate == "" || req.Round < 1 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "라운드, 트랙, 날짜는 필수입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	match := &model.Match{
		LeagueID:    leagueID,
		Round:       req.Round,
		Track:       req.Track,
		MatchDate:   req.MatchDate,
		MatchTime:   req.MatchTime,
		HasSprint:   req.HasSprint,
		SprintDate:  req.SprintDate,
		SprintTime:  req.SprintTime,
		Status:      model.MatchStatusUpcoming,
		Description: req.Description,
	}

	if err := h.matchRepo.Create(ctx, match); err != nil {
		if errors.Is(err, repository.ErrDuplicateRound) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "duplicate_round",
				Message: "이미 해당 라운드가 존재합니다",
			})
		}
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

	matches, err := h.matchRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 일정을 불러오는데 실패했습니다",
		})
	}

	if matches == nil {
		matches = []*model.Match{}
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

	match, err := h.matchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
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

	match, err := h.matchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 정보를 불러오는데 실패했습니다",
		})
	}

	// Update fields if provided
	if req.Round != nil {
		match.Round = *req.Round
	}
	if req.Track != nil {
		match.Track = *req.Track
	}
	if req.MatchDate != nil {
		match.MatchDate = *req.MatchDate
	}
	if req.MatchTime != nil {
		match.MatchTime = req.MatchTime
	}
	if req.HasSprint != nil {
		match.HasSprint = *req.HasSprint
	}
	if req.SprintDate != nil {
		match.SprintDate = req.SprintDate
	}
	if req.SprintTime != nil {
		match.SprintTime = req.SprintTime
	}
	if req.Status != nil {
		match.Status = *req.Status
	}
	if req.Description != nil {
		match.Description = req.Description
	}

	if err := h.matchRepo.Update(ctx, match); err != nil {
		if errors.Is(err, repository.ErrDuplicateRound) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "duplicate_round",
				Message: "이미 해당 라운드가 존재합니다",
			})
		}
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

	if err := h.matchRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "경기를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "경기 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "경기가 삭제되었습니다",
	})
}
