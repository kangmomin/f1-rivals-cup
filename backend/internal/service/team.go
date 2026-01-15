package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// TeamService handles team business logic
type TeamService struct {
	teamRepo    *repository.TeamRepository
	leagueRepo  *repository.LeagueRepository
	accountRepo *repository.AccountRepository
}

// NewTeamService creates a new TeamService instance
func NewTeamService(
	teamRepo *repository.TeamRepository,
	leagueRepo *repository.LeagueRepository,
	accountRepo *repository.AccountRepository,
) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		leagueRepo:  leagueRepo,
		accountRepo: accountRepo,
	}
}

// Create creates a new team and its associated account
func (s *TeamService) Create(ctx context.Context, leagueID uuid.UUID, req *model.CreateTeamRequest) (*model.Team, error) {
	// Validate request
	if req.Name == "" {
		return nil, ErrTeamNameRequired
	}

	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	// Create team
	team := &model.Team{
		LeagueID:   leagueID,
		Name:       req.Name,
		Color:      req.Color,
		IsOfficial: req.IsOfficial,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		if errors.Is(err, repository.ErrTeamAlreadyExists) {
			return nil, ErrTeamAlreadyExists
		}
		return nil, err
	}

	// Create team account automatically (best-effort)
	// Note: Account creation failure is logged but does not fail team creation.
	// This matches the original handler behavior and allows accounts to be
	// created later via FinanceService.CreateTeamAccount if needed.
	if s.accountRepo != nil {
		account := &model.Account{
			LeagueID:  leagueID,
			OwnerID:   team.ID,
			OwnerType: model.OwnerTypeTeam,
			Balance:   0,
		}
		if err := s.accountRepo.Create(ctx, account); err != nil {
			slog.Error("TeamService.Create: failed to create team account",
				"error", err,
				"team_id", team.ID,
				"league_id", leagueID,
			)
		}
	}

	return team, nil
}

// Get retrieves a team by ID
func (s *TeamService) Get(ctx context.Context, id uuid.UUID) (*model.Team, error) {
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	return team, nil
}

// List retrieves all teams for a league
func (s *TeamService) List(ctx context.Context, leagueID uuid.UUID) ([]*model.Team, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	teams, err := s.teamRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for consistent JSON serialization
	if teams == nil {
		teams = []*model.Team{}
	}

	return teams, nil
}

// Update updates a team's information
func (s *TeamService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateTeamRequest) (*model.Team, error) {
	// Get existing team
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	// Apply updates with validation
	if req.Name != nil {
		if *req.Name == "" {
			return nil, ErrTeamNameRequired
		}
		team.Name = *req.Name
	}
	if req.Color != nil {
		team.Color = req.Color
	}

	// Save changes
	if err := s.teamRepo.Update(ctx, team); err != nil {
		if errors.Is(err, repository.ErrTeamAlreadyExists) {
			return nil, ErrTeamAlreadyExists
		}
		return nil, err
	}

	return team, nil
}

// Delete removes a team by ID
func (s *TeamService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.teamRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return ErrTeamNotFound
		}
		return err
	}

	return nil
}
