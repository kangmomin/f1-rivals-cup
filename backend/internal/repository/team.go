package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists in this league")
)

type TeamRepository struct {
	db *database.DB
}

func NewTeamRepository(db *database.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team
func (r *TeamRepository) Create(ctx context.Context, team *model.Team) error {
	query := `
		INSERT INTO teams (league_id, name, color, is_official)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		team.LeagueID,
		team.Name,
		team.Color,
		team.IsOfficial,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"teams_league_id_name_key\"" {
			return ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retrieves a team by ID
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Team, error) {
	query := `
		SELECT id, league_id, name, color, is_official, created_at, updated_at
		FROM teams
		WHERE id = $1
	`

	team := &model.Team{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.LeagueID,
		&team.Name,
		&team.Color,
		&team.IsOfficial,
		&team.CreatedAt,
		&team.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	return team, nil
}

// ListByLeague retrieves all teams for a league
func (r *TeamRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID) ([]*model.Team, error) {
	query := `
		SELECT id, league_id, name, color, is_official, created_at, updated_at
		FROM teams
		WHERE league_id = $1
		ORDER BY is_official DESC, name ASC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*model.Team
	for rows.Next() {
		team := &model.Team{}
		if err := rows.Scan(
			&team.ID,
			&team.LeagueID,
			&team.Name,
			&team.Color,
			&team.IsOfficial,
			&team.CreatedAt,
			&team.UpdatedAt,
		); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// Update updates a team
func (r *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	query := `
		UPDATE teams
		SET name = $1, color = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		team.Name,
		team.Color,
		team.ID,
	).Scan(&team.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamNotFound
		}
		if err.Error() == "pq: duplicate key value violates unique constraint \"teams_league_id_name_key\"" {
			return ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

// Delete removes a team
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamNotFound
	}

	return nil
}
