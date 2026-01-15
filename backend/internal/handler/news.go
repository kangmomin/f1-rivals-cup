package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// NewsHandler handles news requests
type NewsHandler struct {
	newsService *service.NewsService
}

// NewNewsHandler creates a new NewsHandler
func NewNewsHandler(newsService *service.NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
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

	// Get user ID from context
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	news, err := h.newsService.Create(ctx, leagueID, &req, userID)
	if err != nil {
		return h.handleServiceError(c, err, "News.Create")
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
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := c.Request().Context()

	newsList, total, err := h.newsService.List(ctx, leagueID, page, limit)
	if err != nil {
		return h.handleServiceError(c, err, "News.List")
	}

	// Normalize pagination for response (service already normalized internally)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
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
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := c.Request().Context()

	newsList, total, err := h.newsService.ListAll(ctx, leagueID, page, limit)
	if err != nil {
		return h.handleServiceError(c, err, "News.ListAll")
	}

	// Normalize pagination for response (service already normalized internally)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
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

	news, err := h.newsService.Get(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "News.Get")
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

	news, err := h.newsService.GetAdmin(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "News.GetAdmin")
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

	news, err := h.newsService.Update(ctx, id, &req)
	if err != nil {
		return h.handleServiceError(c, err, "News.Update")
	}

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

	news, err := h.newsService.Publish(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "News.Publish")
	}

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

	news, err := h.newsService.Unpublish(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "News.Unpublish")
	}

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

	if err := h.newsService.Delete(ctx, id); err != nil {
		return h.handleServiceError(c, err, "News.Delete")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "뉴스가 삭제되었습니다",
	})
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

	// LeagueID is optional for generate (we use uuid.Nil as placeholder)
	leagueID := uuid.Nil

	ctx := c.Request().Context()

	// Convert to service request type
	serviceReq := &model.GenerateNewsRequest{
		Input: req.Input,
	}

	content, err := h.newsService.GenerateContent(ctx, leagueID, serviceReq)
	if err != nil {
		return h.handleGenerateError(c, err)
	}

	return c.JSON(http.StatusOK, model.GenerateNewsContentResponse{
		Title:        content.Title,
		Description:  content.Description,
		NewsProvider: content.NewsProvider,
	})
}

// handleServiceError handles service layer errors and returns appropriate HTTP responses
func (h *NewsHandler) handleServiceError(c echo.Context, err error, operation string) error {
	// Not found errors
	if errors.Is(err, service.ErrNewsNotFound) {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "뉴스를 찾을 수 없습니다",
		})
	}
	if errors.Is(err, service.ErrLeagueNotFound) {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "리그를 찾을 수 없습니다",
		})
	}

	// Validation errors
	if errors.Is(err, service.ErrNilRequest) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}
	if errors.Is(err, service.ErrNewsEmptyTitle) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "제목을 입력해주세요",
		})
	}
	if errors.Is(err, service.ErrNewsTitleTooShort) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "제목은 최소 2자 이상이어야 합니다",
		})
	}
	if errors.Is(err, service.ErrNewsTitleTooLong) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "제목은 최대 200자까지 가능합니다",
		})
	}
	if errors.Is(err, service.ErrNewsEmptyContent) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "내용을 입력해주세요",
		})
	}

	// Server errors - log and return generic message
	slog.Error(operation+": service error", "error", err)
	return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
		Error:   "server_error",
		Message: "서버 오류가 발생했습니다",
	})
}

// handleGenerateError handles AI content generation errors
func (h *NewsHandler) handleGenerateError(c echo.Context, err error) error {
	// Validation errors
	if errors.Is(err, service.ErrNilRequest) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}
	if errors.Is(err, service.ErrNewsEmptyInput) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "입력 내용을 입력해주세요",
		})
	}

	// AI service unavailable
	if errors.Is(err, service.ErrNewsAIUnavailable) {
		return c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_unavailable",
			Message: "AI 서비스가 설정되지 않았습니다",
		})
	}

	// Context errors
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

	// AI service specific errors
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
	if errors.Is(err, service.ErrInvalidJSON) {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "generation_failed",
			Message: "AI 응답을 처리하는데 실패했습니다. 다시 시도해주세요",
		})
	}

	// Log and return generic error
	slog.Error("News.GenerateContent: AI content generation failed", "error", err)
	return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
		Error:   "server_error",
		Message: "콘텐츠 생성 중 오류가 발생했습니다",
	})
}
