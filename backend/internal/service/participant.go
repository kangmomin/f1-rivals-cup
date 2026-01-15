package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// ParticipantService handles participant-related business logic
type ParticipantService struct {
	participantRepo *repository.ParticipantRepository
	leagueRepo      *repository.LeagueRepository
	accountRepo     *repository.AccountRepository
}

// NewParticipantService creates a new ParticipantService instance
func NewParticipantService(
	participantRepo *repository.ParticipantRepository,
	leagueRepo *repository.LeagueRepository,
	accountRepo *repository.AccountRepository,
) *ParticipantService {
	return &ParticipantService{
		participantRepo: participantRepo,
		leagueRepo:      leagueRepo,
		accountRepo:     accountRepo,
	}
}

// Join creates a new league participation request
func (s *ParticipantService) Join(ctx context.Context, leagueID, userID uuid.UUID, req *model.JoinLeagueRequest) (*model.LeagueParticipant, error) {
	// Validate request
	if req == nil {
		return nil, ErrNilRequest
	}

	// Validate roles
	if len(req.Roles) == 0 {
		return nil, ErrNoRolesProvided
	}

	validRoles := map[string]bool{
		string(model.RoleDirector): true,
		string(model.RolePlayer):   true,
		string(model.RoleReserve):  true,
		string(model.RoleEngineer): true,
	}
	for _, role := range req.Roles {
		if !validRoles[role] {
			return nil, ErrInvalidRole
		}
	}

	// Check if league exists and is open
	league, err := s.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	if league.Status != model.LeagueStatusOpen {
		return nil, ErrLeagueNotOpen
	}

	// Check player limit per team (2 players max)
	// Note: This check is not atomic and may allow over-limit under high concurrency.
	// For strict enforcement, consider adding a DB-level constraint or using a transaction.
	// Current behavior matches the original handler implementation.
	hasPlayerRole := false
	for _, role := range req.Roles {
		if role == string(model.RolePlayer) {
			hasPlayerRole = true
			break
		}
	}

	if hasPlayerRole && req.TeamName != nil && *req.TeamName != "" {
		playerCount, err := s.participantRepo.CountPlayersByTeam(ctx, leagueID, *req.TeamName)
		if err != nil {
			return nil, err
		}
		if playerCount >= 2 {
			return nil, ErrTeamFull
		}
	}

	participant := &model.LeagueParticipant{
		LeagueID: leagueID,
		UserID:   userID,
		Status:   model.ParticipantStatusPending,
		Roles:    req.Roles,
		TeamName: req.TeamName,
		Message:  req.Message,
	}

	if err := s.participantRepo.Create(ctx, participant); err != nil {
		if errors.Is(err, repository.ErrAlreadyParticipating) {
			return nil, ErrAlreadyParticipant
		}
		return nil, err
	}

	return participant, nil
}

// Cancel cancels a pending or rejected participation
func (s *ParticipantService) Cancel(ctx context.Context, leagueID, userID uuid.UUID) error {
	participant, err := s.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return ErrParticipantNotFound
		}
		return err
	}

	// Only pending or rejected can be cancelled by user
	if participant.Status == model.ParticipantStatusApproved {
		return ErrCannotCancelApproved
	}

	return s.participantRepo.Delete(ctx, participant.ID)
}

// GetMyStatus retrieves the current user's participation status in a league
func (s *ParticipantService) GetMyStatus(ctx context.Context, leagueID, userID uuid.UUID) (*model.LeagueParticipant, error) {
	participant, err := s.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return nil, nil // Not participating is not an error
		}
		return nil, err
	}
	return participant, nil
}

// ListByLeague retrieves all participants for a league with optional status filter
func (s *ParticipantService) ListByLeague(ctx context.Context, leagueID uuid.UUID, status string) ([]*model.LeagueParticipant, error) {
	participants, err := s.participantRepo.ListByLeague(ctx, leagueID, status)
	if err != nil {
		return nil, err
	}

	// Ensure non-nil slice for JSON serialization
	if participants == nil {
		participants = []*model.LeagueParticipant{}
	}

	return participants, nil
}

// ListMyParticipations retrieves all participations for a user
func (s *ParticipantService) ListMyParticipations(ctx context.Context, userID uuid.UUID) ([]*model.LeagueParticipant, error) {
	participants, err := s.participantRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Ensure non-nil slice for JSON serialization
	if participants == nil {
		participants = []*model.LeagueParticipant{}
	}

	return participants, nil
}

// UpdateStatus updates a participant's status and creates account when approved
func (s *ParticipantService) UpdateStatus(ctx context.Context, participantID uuid.UUID, status model.ParticipantStatus) (*model.LeagueParticipant, error) {
	// Validate status
	if status != model.ParticipantStatusApproved && status != model.ParticipantStatusRejected {
		return nil, ErrInvalidStatus
	}

	// Get participant to check current status and get league ID
	participant, err := s.participantRepo.GetByID(ctx, participantID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	previousStatus := participant.Status

	if err := s.participantRepo.UpdateStatus(ctx, participantID, status); err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	// Create participant account when approved (if not already approved)
	if status == model.ParticipantStatusApproved && previousStatus != model.ParticipantStatusApproved {
		if err := s.ensureParticipantAccount(ctx, participant.LeagueID, participantID); err != nil {
			// Log error but don't fail the request - account creation is secondary
			slog.Error("ParticipantService.UpdateStatus: failed to ensure participant account",
				"error", err,
				"participant_id", participantID,
				"league_id", participant.LeagueID,
			)
		}
	}

	// Update the participant object to reflect new status
	participant.Status = status
	return participant, nil
}

// ensureParticipantAccount creates a participant account if it doesn't exist
func (s *ParticipantService) ensureParticipantAccount(ctx context.Context, leagueID, participantID uuid.UUID) error {
	// Use the idempotent EnsureParticipantAccount method (handles race conditions with ON CONFLICT)
	_, err := s.accountRepo.EnsureParticipantAccount(ctx, leagueID, participantID)
	if err != nil {
		return err
	}

	slog.Debug("ParticipantService: ensured participant account",
		"participant_id", participantID,
		"league_id", leagueID,
	)
	return nil
}

// UpdateTeam updates a participant's team assignment
func (s *ParticipantService) UpdateTeam(ctx context.Context, participantID uuid.UUID, teamName *string) error {
	if err := s.participantRepo.UpdateTeam(ctx, participantID, teamName); err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return ErrParticipantNotFound
		}
		return err
	}
	return nil
}

// GetByID retrieves a participant by ID
func (s *ParticipantService) GetByID(ctx context.Context, participantID uuid.UUID) (*model.LeagueParticipant, error) {
	participant, err := s.participantRepo.GetByID(ctx, participantID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}
	return participant, nil
}
