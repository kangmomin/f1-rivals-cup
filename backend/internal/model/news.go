package model

import (
	"time"

	"github.com/google/uuid"
)

// News represents a news article in the system
type News struct {
	ID          uuid.UUID  `json:"id"`
	LeagueID    uuid.UUID  `json:"league_id"`
	AuthorID    uuid.UUID  `json:"author_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	IsPublished bool       `json:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Joined fields
	AuthorNickname string `json:"author_nickname,omitempty"`
}

// CreateNewsRequest represents a request to create a news article
type CreateNewsRequest struct {
	Title   string `json:"title" validate:"required,min=2,max=200"`
	Content string `json:"content" validate:"required"`
}

// UpdateNewsRequest represents a request to update a news article
type UpdateNewsRequest struct {
	Title   *string `json:"title,omitempty" validate:"omitempty,min=2,max=200"`
	Content *string `json:"content,omitempty"`
}

// ListNewsResponse represents the response for listing news
type ListNewsResponse struct {
	News       []*News `json:"news"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}

// GenerateNewsContentRequest represents a request to generate news content using AI
type GenerateNewsContentRequest struct {
	Input string `json:"input" validate:"required"`
}

// GenerateNewsContentResponse represents the response for AI-generated news content
type GenerateNewsContentResponse struct {
	Content string `json:"content"`
}
