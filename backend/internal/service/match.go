package service

import (
	"context"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// MatchService handles match-related business logic
type MatchService struct {
	matchRepo  *repository.MatchRepository
	leagueRepo *repository.LeagueRepository
}

// NewMatchService creates a new MatchService instance
func NewMatchService(
	matchRepo *repository.MatchRepository,
	leagueRepo *repository.LeagueRepository,
) *MatchService {
	return &MatchService{
		matchRepo:  matchRepo,
		leagueRepo: leagueRepo,
	}
}

// Create creates a new match for a league
func (s *MatchService) Create(ctx context.Context, leagueID uuid.UUID, req *model.CreateMatchRequest) (*model.Match, error) {
	// Validate request
	if req.Track == "" || req.MatchDate == "" || req.Round < 1 {
		return nil, ErrInvalidMatchRequest
	}

	// Check if league exists
	_, err := s.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
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

	if err := s.matchRepo.Create(ctx, match); err != nil {
		if errors.Is(err, repository.ErrDuplicateRound) {
			return nil, ErrDuplicateRound
		}
		return nil, err
	}

	return match, nil
}

// Get retrieves a match by ID
func (s *MatchService) Get(ctx context.Context, id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}

	return match, nil
}

// List retrieves all matches for a league
func (s *MatchService) List(ctx context.Context, leagueID uuid.UUID) ([]*model.Match, error) {
	matches, err := s.matchRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	// Return empty slice instead of nil
	if matches == nil {
		matches = []*model.Match{}
	}

	return matches, nil
}

// Update updates an existing match
func (s *MatchService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateMatchRequest) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}

	// Update fields if provided with validation
	if req.Round != nil {
		if *req.Round < 1 {
			return nil, ErrInvalidMatchRequest
		}
		match.Round = *req.Round
	}
	if req.Track != nil {
		if *req.Track == "" {
			return nil, ErrInvalidMatchRequest
		}
		match.Track = *req.Track
	}
	if req.MatchDate != nil {
		if *req.MatchDate == "" {
			return nil, ErrInvalidMatchRequest
		}
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

	if err := s.matchRepo.Update(ctx, match); err != nil {
		if errors.Is(err, repository.ErrDuplicateRound) {
			return nil, ErrDuplicateRound
		}
		return nil, err
	}

	return match, nil
}

// Delete removes a match by ID
func (s *MatchService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.matchRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrMatchNotFound) {
			return ErrMatchNotFound
		}
		return err
	}

	return nil
}
