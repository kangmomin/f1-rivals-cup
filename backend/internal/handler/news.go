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

// NewsHandler handles news requests
type NewsHandler struct {
	newsRepo   *repository.NewsRepository
	leagueRepo *repository.LeagueRepository
}

// NewNewsHandler creates a new NewsHandler
func NewNewsHandler(newsRepo *repository.NewsRepository, leagueRepo *repository.LeagueRepository) *NewsHandler {
	return &NewsHandler{
		newsRepo:   newsRepo,
		leagueRepo: leagueRepo,
	}
}

// Create handles POST /api/v1/leagues/:id/news
func (h *NewsHandler) Create(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	var req model.CreateNewsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate request
	if err := validateCreateNewsRequest(&req); err != nil {
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
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	news := &model.News{
		LeagueID:    leagueID,
		AuthorID:    userID,
		Title:       req.Title,
		Content:     req.Content,
		IsPublished: false,
	}

	if err := h.newsRepo.Create(ctx, news); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 작성에 실패했습니다",
		})
	}

	// Get author nickname from context
	nickname, _ := c.Get("nickname").(string)
	news.AuthorNickname = nickname

	return c.JSON(http.StatusCreated, news)
}

// List handles GET /api/v1/leagues/:id/news
func (h *NewsHandler) List(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
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
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	// Public endpoint shows only published news
	newsList, total, err := h.newsRepo.ListByLeague(ctx, leagueID, page, limit, true)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 목록을 불러오는데 실패했습니다",
		})
	}

	if newsList == nil {
		newsList = []*model.News{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, model.ListNewsResponse{
		News:       newsList,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// ListAll handles GET /api/v1/admin/leagues/:id/news (includes unpublished)
func (h *NewsHandler) ListAll(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
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
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	// Admin endpoint shows all news (including unpublished)
	newsList, total, err := h.newsRepo.ListByLeague(ctx, leagueID, page, limit, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 목록을 불러오는데 실패했습니다",
		})
	}

	if newsList == nil {
		newsList = []*model.News{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, model.ListNewsResponse{
		News:       newsList,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// Get handles GET /api/v1/news/:id
func (h *NewsHandler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	ctx := c.Request().Context()

	news, err := h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	// Public endpoint only returns published news
	if !news.IsPublished {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "뉴스를 찾을 수 없습니다",
		})
	}

	return c.JSON(http.StatusOK, news)
}

// GetAdmin handles GET /api/v1/admin/news/:id (includes unpublished)
func (h *NewsHandler) GetAdmin(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	ctx := c.Request().Context()

	news, err := h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, news)
}

// Update handles PUT /api/v1/news/:id
func (h *NewsHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	var req model.UpdateNewsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	// Get existing news
	news, err := h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	// Update fields
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "제목을 입력해주세요",
			})
		}
		if len(title) < 2 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "제목은 최소 2자 이상이어야 합니다",
			})
		}
		if len(title) > 200 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "제목은 최대 200자까지 가능합니다",
			})
		}
		news.Title = title
	}
	if req.Content != nil {
		news.Content = *req.Content
	}

	if err := h.newsRepo.Update(ctx, news); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 수정에 실패했습니다",
		})
	}

	// Reload to get updated_at
	news, _ = h.newsRepo.GetByID(ctx, id)

	return c.JSON(http.StatusOK, news)
}

// Publish handles PUT /api/v1/news/:id/publish
func (h *NewsHandler) Publish(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if news exists
	_, err = h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	if err := h.newsRepo.Publish(ctx, id, true); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 발행에 실패했습니다",
		})
	}

	// Reload to get updated state
	news, _ := h.newsRepo.GetByID(ctx, id)

	return c.JSON(http.StatusOK, news)
}

// Unpublish handles PUT /api/v1/news/:id/unpublish
func (h *NewsHandler) Unpublish(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if news exists
	_, err = h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	if err := h.newsRepo.Publish(ctx, id, false); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 발행 취소에 실패했습니다",
		})
	}

	// Reload to get updated state
	news, _ := h.newsRepo.GetByID(ctx, id)

	return c.JSON(http.StatusOK, news)
}

// Delete handles DELETE /api/v1/news/:id
func (h *NewsHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.newsRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "뉴스가 삭제되었습니다",
	})
}

func validateCreateNewsRequest(req *model.CreateNewsRequest) error {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)

	if req.Title == "" {
		return errors.New("제목을 입력해주세요")
	}
	if len(req.Title) < 2 {
		return errors.New("제목은 최소 2자 이상이어야 합니다")
	}
	if len(req.Title) > 200 {
		return errors.New("제목은 최대 200자까지 가능합니다")
	}
	if req.Content == "" {
		return errors.New("내용을 입력해주세요")
	}

	return nil
}
