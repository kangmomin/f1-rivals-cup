package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	commentRepo *repository.CommentRepository
}

func NewCommentHandler(commentRepo *repository.CommentRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
	}
}

// Create handles POST /api/v1/news/:id/comments
func (h *CommentHandler) Create(c echo.Context) error {
	newsIDStr := c.Param("id")
	newsID, err := uuid.Parse(newsIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate content
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "댓글 내용을 입력해주세요",
		})
	}
	if len(req.Content) > 1000 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "댓글은 1000자를 초과할 수 없습니다",
		})
	}

	ctx := c.Request().Context()

	comment := &model.NewsComment{
		NewsID:   newsID,
		AuthorID: userID.(uuid.UUID),
		Content:  req.Content,
	}

	if err := h.commentRepo.Create(ctx, comment); err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "뉴스를 찾을 수 없습니다",
			})
		}
		slog.Error("Comment.Create: failed to create comment", "error", err, "news_id", newsID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글 작성에 실패했습니다",
		})
	}

	// Get comment with author nickname
	fullComment, err := h.commentRepo.GetByID(ctx, comment.ID)
	if err != nil {
		// Return basic comment if fetching full info fails
		return c.JSON(http.StatusCreated, comment)
	}

	return c.JSON(http.StatusCreated, fullComment)
}

// List handles GET /api/v1/news/:id/comments
func (h *CommentHandler) List(c echo.Context) error {
	newsIDStr := c.Param("id")
	newsID, err := uuid.Parse(newsIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 뉴스 ID입니다",
		})
	}

	// Parse pagination params
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request().Context()

	comments, total, err := h.commentRepo.ListByNews(ctx, newsID, page, limit)
	if err != nil {
		slog.Error("Comment.List: failed to list comments", "error", err, "news_id", newsID, "page", page, "limit", limit)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글 목록을 불러오는데 실패했습니다",
		})
	}

	if comments == nil {
		comments = []*model.NewsComment{}
	}

	return c.JSON(http.StatusOK, model.ListCommentsResponse{
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// Update handles PUT /api/v1/comments/:id
func (h *CommentHandler) Update(c echo.Context) error {
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 댓글 ID입니다",
		})
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	var req model.UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate content
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "댓글 내용을 입력해주세요",
		})
	}
	if len(req.Content) > 1000 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "댓글은 1000자를 초과할 수 없습니다",
		})
	}

	ctx := c.Request().Context()

	// Get existing comment to check ownership
	comment, err := h.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "댓글을 찾을 수 없습니다",
			})
		}
		slog.Error("Comment.Update: failed to get comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글을 불러오는데 실패했습니다",
		})
	}

	// Check if user is the author
	if comment.AuthorID != userID.(uuid.UUID) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 댓글만 수정할 수 있습니다",
		})
	}

	// Update comment
	comment.Content = req.Content
	if err := h.commentRepo.Update(ctx, comment); err != nil {
		slog.Error("Comment.Update: failed to update comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글 수정에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, comment)
}

// Delete handles DELETE /api/v1/comments/:id
func (h *CommentHandler) Delete(c echo.Context) error {
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 댓글 ID입니다",
		})
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	// Get existing comment to check ownership
	comment, err := h.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "댓글을 찾을 수 없습니다",
			})
		}
		slog.Error("Comment.Delete: failed to get comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글을 불러오는데 실패했습니다",
		})
	}

	// Check if user is the author
	if comment.AuthorID != userID.(uuid.UUID) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 댓글만 삭제할 수 있습니다",
		})
	}

	// Delete comment
	if err := h.commentRepo.Delete(ctx, commentID); err != nil {
		slog.Error("Comment.Delete: failed to delete comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "댓글 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "댓글이 삭제되었습니다",
	})
}
