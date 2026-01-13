package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrParticipantNotFound    = errors.New("participant not found")
	ErrAlreadyParticipating   = errors.New("already participating in this league")
)

type ParticipantRepository struct {
	db *database.DB
}

func NewParticipantRepository(db *database.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

// Create creates a new league participant
func (r *ParticipantRepository) Create(ctx context.Context, participant *model.LeagueParticipant) error {
	query := `
		INSERT INTO league_participants (league_id, user_id, status, roles, team_name, message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		participant.LeagueID,
		participant.UserID,
		participant.Status,
		pq.Array(participant.Roles),
		participant.TeamName,
		participant.Message,
	).Scan(&participant.ID, &participant.CreatedAt, &participant.UpdatedAt)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "league_participants_league_id_user_id_key"` {
			return ErrAlreadyParticipating
		}
		return err
	}

	return nil
}

// GetByLeagueAndUser retrieves a participant by league and user ID
func (r *ParticipantRepository) GetByLeagueAndUser(ctx context.Context, leagueID, userID uuid.UUID) (*model.LeagueParticipant, error) {
	query := `
		SELECT id, league_id, user_id, status, roles, team_name, message, created_at, updated_at
		FROM league_participants
		WHERE league_id = $1 AND user_id = $2
	`

	participant := &model.LeagueParticipant{}
	err := r.db.Pool.QueryRowContext(ctx, query, leagueID, userID).Scan(
		&participant.ID,
		&participant.LeagueID,
		&participant.UserID,
		&participant.Status,
		&participant.Roles,
		&participant.TeamName,
		&participant.Message,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	return participant, nil
}

// GetByID retrieves a participant by ID
func (r *ParticipantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.LeagueParticipant, error) {
	query := `
		SELECT id, league_id, user_id, status, roles, team_name, message, created_at, updated_at
		FROM league_participants
		WHERE id = $1
	`

	participant := &model.LeagueParticipant{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&participant.ID,
		&participant.LeagueID,
		&participant.UserID,
		&participant.Status,
		&participant.Roles,
		&participant.TeamName,
		&participant.Message,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	return participant, nil
}

// ListByLeague retrieves all participants for a league
func (r *ParticipantRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID, status string) ([]*model.LeagueParticipant, error) {
	query := `
		SELECT lp.id, lp.league_id, lp.user_id, lp.status, lp.roles, lp.team_name, lp.message, lp.created_at, lp.updated_at,
		       u.nickname, u.email
		FROM league_participants lp
		JOIN users u ON lp.user_id = u.id
		WHERE lp.league_id = $1 AND ($2 = '' OR lp.status = $2)
		ORDER BY lp.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*model.LeagueParticipant
	for rows.Next() {
		p := &model.LeagueParticipant{}
		if err := rows.Scan(
			&p.ID,
			&p.LeagueID,
			&p.UserID,
			&p.Status,
			&p.Roles,
			&p.TeamName,
			&p.Message,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.UserNickname,
			&p.UserEmail,
		); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	return participants, nil
}

// ListByUser retrieves all participations for a user
func (r *ParticipantRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.LeagueParticipant, error) {
	query := `
		SELECT lp.id, lp.league_id, lp.user_id, lp.status, lp.roles, lp.team_name, lp.message, lp.created_at, lp.updated_at,
		       l.name
		FROM league_participants lp
		JOIN leagues l ON lp.league_id = l.id
		WHERE lp.user_id = $1
		ORDER BY lp.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*model.LeagueParticipant
	for rows.Next() {
		p := &model.LeagueParticipant{}
		if err := rows.Scan(
			&p.ID,
			&p.LeagueID,
			&p.UserID,
			&p.Status,
			&p.Roles,
			&p.TeamName,
			&p.Message,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.LeagueName,
		); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	return participants, nil
}

// UpdateStatus updates the status of a participant
func (r *ParticipantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ParticipantStatus) error {
	query := `
		UPDATE league_participants
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Pool.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrParticipantNotFound
	}

	return nil
}

// Delete removes a participant
func (r *ParticipantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM league_participants WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrParticipantNotFound
	}

	return nil
}

// CountByLeague counts participants in a league
func (r *ParticipantRepository) CountByLeague(ctx context.Context, leagueID uuid.UUID, status string) (int, error) {
	query := `SELECT COUNT(*) FROM league_participants WHERE league_id = $1 AND ($2 = '' OR status = $2)`
	var count int
	err := r.db.Pool.QueryRowContext(ctx, query, leagueID, status).Scan(&count)
	return count, err
}

// CountPlayersByTeam counts approved players (role='player') in a specific team for a league
func (r *ParticipantRepository) CountPlayersByTeam(ctx context.Context, leagueID uuid.UUID, teamName string) (int, error) {
	query := `
		SELECT COUNT(*) FROM league_participants
		WHERE league_id = $1
		AND team_name = $2
		AND status = 'approved'
		AND 'player' = ANY(roles)
	`
	var count int
	err := r.db.Pool.QueryRowContext(ctx, query, leagueID, teamName).Scan(&count)
	return count, err
}

// UpdateTeam updates the team assignment of a participant
func (r *ParticipantRepository) UpdateTeam(ctx context.Context, id uuid.UUID, teamName *string) error {
	query := `
		UPDATE league_participants
		SET team_name = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Pool.ExecContext(ctx, query, teamName, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrParticipantNotFound
	}

	return nil
}

// GetDirectorTeams returns the team names where the user is an approved director in a league
// Deprecated: Use GetDirectorTeamIDs instead for more robust authorization
func (r *ParticipantRepository) GetDirectorTeams(ctx context.Context, leagueID, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT team_name
		FROM league_participants
		WHERE league_id = $1
		AND user_id = $2
		AND status = 'approved'
		AND 'director' = ANY(roles)
		AND team_name IS NOT NULL
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []string
	for rows.Next() {
		var teamName string
		if err := rows.Scan(&teamName); err != nil {
			return nil, err
		}
		teams = append(teams, teamName)
	}

	return teams, nil
}

// GetDirectorTeamIDs returns the team IDs where the user is an approved director in a league
func (r *ParticipantRepository) GetDirectorTeamIDs(ctx context.Context, leagueID, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT t.id
		FROM league_participants lp
		JOIN teams t ON t.league_id = lp.league_id AND t.name = lp.team_name
		WHERE lp.league_id = $1
		AND lp.user_id = $2
		AND lp.status = 'approved'
		AND 'director' = ANY(lp.roles)
		AND lp.team_name IS NOT NULL
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamIDs []uuid.UUID
	for rows.Next() {
		var teamID uuid.UUID
		if err := rows.Scan(&teamID); err != nil {
			return nil, err
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}
