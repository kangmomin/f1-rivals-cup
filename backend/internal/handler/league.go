package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LeagueHandler handles league requests
type LeagueHandler struct {
	leagueRepo *repository.LeagueRepository
}

// NewLeagueHandler creates a new LeagueHandler
func NewLeagueHandler(leagueRepo *repository.LeagueRepository) *LeagueHandler {
	return &LeagueHandler{
		leagueRepo: leagueRepo,
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

	// Validate request
	if err := validateCreateLeagueRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
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

	// Parse dates
	startDate, _ := repository.ParseTime(safeString(req.StartDate))
	endDate, _ := repository.ParseTime(safeString(req.EndDate))

	// Set default season
	season := req.Season
	if season < 1 {
		season = 1
	}

	league := &model.League{
		Name:        req.Name,
		Description: req.Description,
		Status:      model.LeagueStatusDraft,
		Season:      season,
		CreatedBy:   userID,
		StartDate:   startDate,
		EndDate:     endDate,
		MatchTime:   req.MatchTime,
		Rules:       req.Rules,
		Settings:    req.Settings,
		ContactInfo: req.ContactInfo,
	}

	ctx := c.Request().Context()

	if err := h.leagueRepo.Create(ctx, league); err != nil {
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
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := c.QueryParam("status")

	ctx := c.Request().Context()

	leagues, total, err := h.leagueRepo.List(ctx, page, limit, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 목록을 불러오는데 실패했습니다",
		})
	}

	if leagues == nil {
		leagues = []*model.League{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, model.ListLeaguesResponse{
		Leagues:    leagues,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
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

	league, err := h.leagueRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
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

	// Get existing league
	league, err := h.leagueRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	// Update fields
	if req.Name != nil {
		league.Name = *req.Name
	}
	if req.Description != nil {
		league.Description = req.Description
	}
	if req.Status != nil {
		league.Status = model.LeagueStatus(*req.Status)
	}
	if req.Season != nil {
		league.Season = *req.Season
	}
	if req.StartDate != nil {
		league.StartDate, _ = repository.ParseTime(*req.StartDate)
	}
	if req.EndDate != nil {
		league.EndDate, _ = repository.ParseTime(*req.EndDate)
	}
	if req.MatchTime != nil {
		league.MatchTime = req.MatchTime
	}
	if req.Rules != nil {
		league.Rules = req.Rules
	}
	if req.Settings != nil {
		league.Settings = req.Settings
	}
	if req.ContactInfo != nil {
		league.ContactInfo = req.ContactInfo
	}

	if err := h.leagueRepo.Update(ctx, league); err != nil {
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

	if err := h.leagueRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "리그가 삭제되었습니다",
	})
}

func validateCreateLeagueRequest(req *model.CreateLeagueRequest) error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return errors.New("리그 이름을 입력해주세요")
	}
	if len(req.Name) < 2 {
		return errors.New("리그 이름은 최소 2자 이상이어야 합니다")
	}
	if len(req.Name) > 100 {
		return errors.New("리그 이름은 최대 100자까지 가능합니다")
	}

	return nil
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
