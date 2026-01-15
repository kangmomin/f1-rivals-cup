package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// LeagueService handles league business logic
type LeagueService struct {
	leagueRepo *repository.LeagueRepository
}

// NewLeagueService creates a new LeagueService
func NewLeagueService(leagueRepo *repository.LeagueRepository) *LeagueService {
	return &LeagueService{leagueRepo: leagueRepo}
}

// Create creates a new league
func (s *LeagueService) Create(ctx context.Context, req *model.CreateLeagueRequest, userID uuid.UUID) (*model.League, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Parse dates
	startDate, err := parseTime(safeString(req.StartDate))
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	endDate, err := parseTime(safeString(req.EndDate))
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

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

	if err := s.leagueRepo.Create(ctx, league); err != nil {
		return nil, err
	}

	return league, nil
}

// Get retrieves a league by ID
func (s *LeagueService) Get(ctx context.Context, id uuid.UUID) (*model.League, error) {
	league, err := s.leagueRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}
	return league, nil
}

// ListResult contains paginated league list result
type ListResult struct {
	Leagues    []*model.League
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves a paginated list of leagues
func (s *LeagueService) List(ctx context.Context, page, pageSize int, status string) (*ListResult, error) {
	// Normalize pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Validate status if provided
	if status != "" && !isValidStatus(status) {
		return nil, ErrInvalidLeagueStatus
	}

	leagues, total, err := s.leagueRepo.List(ctx, page, pageSize, status)
	if err != nil {
		return nil, err
	}

	if leagues == nil {
		leagues = []*model.League{}
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &ListResult{
		Leagues:    leagues,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates an existing league
func (s *LeagueService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateLeagueRequest) (*model.League, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Get existing league
	league, err := s.leagueRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	// Validate and apply updates
	if err := s.applyUpdates(league, req); err != nil {
		return nil, err
	}

	if err := s.leagueRepo.Update(ctx, league); err != nil {
		return nil, err
	}

	return league, nil
}

// Delete removes a league
func (s *LeagueService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.leagueRepo.Delete(ctx, id)
}

// validateCreateRequest validates the create league request
func (s *LeagueService) validateCreateRequest(req *model.CreateLeagueRequest) error {
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

// applyUpdates applies update request fields to an existing league
func (s *LeagueService) applyUpdates(league *model.League, req *model.UpdateLeagueRequest) error {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return errors.New("리그 이름을 입력해주세요")
		}
		if len(name) < 2 {
			return errors.New("리그 이름은 최소 2자 이상이어야 합니다")
		}
		if len(name) > 100 {
			return errors.New("리그 이름은 최대 100자까지 가능합니다")
		}
		league.Name = name
	}
	if req.Description != nil {
		league.Description = req.Description
	}
	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return ErrInvalidLeagueStatus
		}
		league.Status = model.LeagueStatus(*req.Status)
	}
	if req.Season != nil {
		if *req.Season < 1 {
			return errors.New("시즌은 1 이상이어야 합니다")
		}
		league.Season = *req.Season
	}
	if req.StartDate != nil {
		startDate, err := parseTime(*req.StartDate)
		if err != nil {
			return ErrInvalidDateFormat
		}
		league.StartDate = startDate
	}
	if req.EndDate != nil {
		endDate, err := parseTime(*req.EndDate)
		if err != nil {
			return ErrInvalidDateFormat
		}
		league.EndDate = endDate
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

	return nil
}

// isValidStatus checks if the given status is a valid league status
func isValidStatus(status string) bool {
	switch model.LeagueStatus(status) {
	case model.LeagueStatusDraft,
		model.LeagueStatusOpen,
		model.LeagueStatusInProgress,
		model.LeagueStatusCompleted,
		model.LeagueStatusCancelled:
		return true
	}
	return false
}

// parseTime parses a time string to time.Time pointer
func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try date only format
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}

// safeString safely dereferences a string pointer
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
