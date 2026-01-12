package model

import (
	"time"

	"github.com/google/uuid"
)

// NewsComment represents a comment on a news article
type NewsComment struct {
	ID             uuid.UUID `json:"id"`
	NewsID         uuid.UUID `json:"news_id"`
	AuthorID       uuid.UUID `json:"author_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AuthorNickname string    `json:"author_nickname,omitempty"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// ListCommentsResponse represents the response for listing comments
type ListCommentsResponse struct {
	Comments []*NewsComment `json:"comments"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	Limit    int            `json:"limit"`
}
