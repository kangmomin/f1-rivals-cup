package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/f1-rivals-cup/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// NewsHandler handles news requests
type NewsHandler struct {
	newsRepo   *repository.NewsRepository
	leagueRepo *repository.LeagueRepository
	aiService  *service.AIService
}

// NewNewsHandler creates a new NewsHandler
func NewNewsHandler(newsRepo *repository.NewsRepository, leagueRepo *repository.LeagueRepository, aiService *service.AIService) *NewsHandler {
	return &NewsHandler{
		newsRepo:   newsRepo,
		leagueRepo: leagueRepo,
		aiService:  aiService,
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
		slog.Error("News.Create: failed to get league", "error", err, "league_id", leagueID)
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
		slog.Error("News.Create: failed to create news", "error", err, "league_id", leagueID, "user_id", userID)
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
		slog.Error("News.List: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	// Public endpoint shows only published news
	newsList, total, err := h.newsRepo.ListByLeague(ctx, leagueID, page, limit, true)
	if err != nil {
		slog.Error("News.List: failed to list news", "error", err, "league_id", leagueID, "page", page, "limit", limit)
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
		slog.Error("News.ListAll: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그를 불러오는데 실패했습니다",
		})
	}

	// Admin endpoint shows all news (including unpublished)
	newsList, total, err := h.newsRepo.ListByLeague(ctx, leagueID, page, limit, false)
	if err != nil {
		slog.Error("News.ListAll: failed to list news", "error", err, "league_id", leagueID, "page", page, "limit", limit)
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
		slog.Error("News.Get: failed to get news", "error", err, "news_id", id)
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
		slog.Error("News.GetAdmin: failed to get news", "error", err, "news_id", id)
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
		slog.Error("News.Update: failed to get news", "error", err, "news_id", id)
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
		slog.Error("News.Update: failed to update news", "error", err, "news_id", id)
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
		slog.Error("News.Publish: failed to get news", "error", err, "news_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	if err := h.newsRepo.Publish(ctx, id, true); err != nil {
		slog.Error("News.Publish: failed to publish news", "error", err, "news_id", id)
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
		slog.Error("News.Unpublish: failed to get news", "error", err, "news_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "뉴스를 불러오는데 실패했습니다",
		})
	}

	if err := h.newsRepo.Publish(ctx, id, false); err != nil {
		slog.Error("News.Unpublish: failed to unpublish news", "error", err, "news_id", id)
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
		slog.Error("News.Delete: failed to delete news", "error", err, "news_id", id)
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

// GenerateContent handles POST /api/v1/admin/news/generate
func (h *NewsHandler) GenerateContent(c echo.Context) error {
	var req model.GenerateNewsContentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate input
	req.Input = strings.TrimSpace(req.Input)
	if req.Input == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "입력 내용을 입력해주세요",
		})
	}

	// Check if AI service is configured
	if h.aiService == nil || !h.aiService.IsConfigured() {
		return c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_unavailable",
			Message: "AI 서비스가 설정되지 않았습니다",
		})
	}

	ctx := c.Request().Context()

	// Generate content using AI
	content, err := h.aiService.GenerateNewsContent(ctx, req.Input)
	if err != nil {
		// Log the error for debugging
		slog.Error("News.GenerateContent: AI content generation failed",
			"error", err,
			"input_length", len(req.Input),
		)

		// Handle context cancellation/timeout
		if errors.Is(err, context.Canceled) {
			return c.JSON(499, model.ErrorResponse{ // 499 Client Closed Request
				Error:   "request_cancelled",
				Message: "요청이 취소되었습니다",
			})
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return c.JSON(http.StatusGatewayTimeout, model.ErrorResponse{
				Error:   "timeout",
				Message: "AI 서비스 응답 시간이 초과되었습니다",
			})
		}
		// Handle service-specific errors
		if errors.Is(err, service.ErrNoAPIKey) {
			return c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
				Error:   "service_unavailable",
				Message: "AI 서비스가 설정되지 않았습니다",
			})
		}
		if errors.Is(err, service.ErrNoContent) {
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "generation_failed",
				Message: "콘텐츠 생성에 실패했습니다. 다시 시도해주세요",
			})
		}
		if errors.Is(err, service.ErrAPIRateLimit) {
			return c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
				Error:   "rate_limit",
				Message: "요청이 너무 많습니다. 잠시 후 다시 시도해주세요",
			})
		}
		if errors.Is(err, service.ErrAPIUnavailable) {
			return c.JSON(http.StatusBadGateway, model.ErrorResponse{
				Error:   "service_unavailable",
				Message: "AI 서비스가 일시적으로 이용 불가합니다",
			})
		}
		if errors.Is(err, service.ErrAPIBadRequest) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "bad_request",
				Message: "AI 서비스 요청이 잘못되었습니다",
			})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "콘텐츠 생성 중 오류가 발생했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.GenerateNewsContentResponse{
		Content: content,
	})
}
