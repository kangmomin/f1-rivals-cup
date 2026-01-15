package service

import (
	"context"
	"errors"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// NewsService handles news business logic
type NewsService struct {
	newsRepo   *repository.NewsRepository
	leagueRepo *repository.LeagueRepository
	aiService  *AIService
}

// NewNewsService creates a new NewsService
func NewNewsService(
	newsRepo *repository.NewsRepository,
	leagueRepo *repository.LeagueRepository,
	aiService *AIService,
) *NewsService {
	return &NewsService{
		newsRepo:   newsRepo,
		leagueRepo: leagueRepo,
		aiService:  aiService,
	}
}

// Create creates a new news article
func (s *NewsService) Create(ctx context.Context, leagueID uuid.UUID, req *model.CreateNewsRequest, authorID uuid.UUID) (*model.News, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	news := &model.News{
		LeagueID:    leagueID,
		AuthorID:    authorID,
		Title:       req.Title,
		Content:     req.Content,
		IsPublished: false,
	}

	if err := s.newsRepo.Create(ctx, news); err != nil {
		return nil, err
	}

	return news, nil
}

// Get retrieves a published news article by ID
func (s *NewsService) Get(ctx context.Context, id uuid.UUID) (*model.News, error) {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	// Public access returns only published news
	if !news.IsPublished {
		return nil, ErrNewsNotFound
	}

	return news, nil
}

// GetAdmin retrieves any news article by ID (including unpublished)
func (s *NewsService) GetAdmin(ctx context.Context, id uuid.UUID) (*model.News, error) {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	return news, nil
}

// List retrieves a paginated list of published news for a league
func (s *NewsService) List(ctx context.Context, leagueID uuid.UUID, page, pageSize int) ([]*model.News, int, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, 0, ErrLeagueNotFound
		}
		return nil, 0, err
	}

	// Normalize pagination parameters
	page, pageSize = s.normalizePagination(page, pageSize)

	// List only published news
	newsList, total, err := s.newsRepo.ListByLeague(ctx, leagueID, page, pageSize, true)
	if err != nil {
		return nil, 0, err
	}

	if newsList == nil {
		newsList = []*model.News{}
	}

	return newsList, total, nil
}

// ListAll retrieves a paginated list of all news for a league (including unpublished)
func (s *NewsService) ListAll(ctx context.Context, leagueID uuid.UUID, page, pageSize int) ([]*model.News, int, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, 0, ErrLeagueNotFound
		}
		return nil, 0, err
	}

	// Normalize pagination parameters
	page, pageSize = s.normalizePagination(page, pageSize)

	// List all news (including unpublished)
	newsList, total, err := s.newsRepo.ListByLeague(ctx, leagueID, page, pageSize, false)
	if err != nil {
		return nil, 0, err
	}

	if newsList == nil {
		newsList = []*model.News{}
	}

	return newsList, total, nil
}

// Update updates a news article
func (s *NewsService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateNewsRequest) (*model.News, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Get existing news
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	// Update title if provided
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if err := s.validateTitle(title); err != nil {
			return nil, err
		}
		news.Title = title
	}

	// Update content if provided
	if req.Content != nil {
		news.Content = *req.Content
	}

	if err := s.newsRepo.Update(ctx, news); err != nil {
		return nil, err
	}

	// Reload to get updated_at
	updatedNews, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		// Update succeeded but reload failed; return original with updated fields
		return news, nil
	}

	return updatedNews, nil
}

// Publish publishes a news article
func (s *NewsService) Publish(ctx context.Context, id uuid.UUID) (*model.News, error) {
	// Check if news exists
	existingNews, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	if err := s.newsRepo.Publish(ctx, id, true); err != nil {
		return nil, err
	}

	// Reload to get updated state
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		// Publish succeeded but reload failed; return existing with updated state
		existingNews.IsPublished = true
		return existingNews, nil
	}

	return news, nil
}

// Unpublish unpublishes a news article
func (s *NewsService) Unpublish(ctx context.Context, id uuid.UUID) (*model.News, error) {
	// Check if news exists
	existingNews, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	if err := s.newsRepo.Publish(ctx, id, false); err != nil {
		return nil, err
	}

	// Reload to get updated state
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		// Unpublish succeeded but reload failed; return existing with updated state
		existingNews.IsPublished = false
		existingNews.PublishedAt = nil
		return existingNews, nil
	}

	return news, nil
}

// Delete deletes a news article
func (s *NewsService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.newsRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNewsNotFound) {
			return ErrNewsNotFound
		}
		return err
	}

	return nil
}

// GenerateContent generates news content using AI
func (s *NewsService) GenerateContent(ctx context.Context, leagueID uuid.UUID, req *model.GenerateNewsRequest) (*model.GeneratedNews, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Validate input
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return nil, ErrNewsEmptyInput
	}

	// Check if AI service is configured
	if s.aiService == nil || !s.aiService.IsConfigured() {
		return nil, ErrNewsAIUnavailable
	}

	// Generate content using AI
	content, err := s.aiService.GenerateNewsContent(ctx, input)
	if err != nil {
		return nil, err
	}

	return &model.GeneratedNews{
		Title:        content.Title,
		Description:  content.Description,
		NewsProvider: content.NewsProvider,
	}, nil
}

// validateCreateRequest validates a CreateNewsRequest
func (s *NewsService) validateCreateRequest(req *model.CreateNewsRequest) error {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)

	if err := s.validateTitle(req.Title); err != nil {
		return err
	}

	if req.Content == "" {
		return ErrNewsEmptyContent
	}

	return nil
}

// validateTitle validates the news title
func (s *NewsService) validateTitle(title string) error {
	if title == "" {
		return ErrNewsEmptyTitle
	}
	if len(title) < 2 {
		return ErrNewsTitleTooShort
	}
	if len(title) > 200 {
		return ErrNewsTitleTooLong
	}
	return nil
}

// normalizePagination normalizes page and pageSize values
func (s *NewsService) normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
