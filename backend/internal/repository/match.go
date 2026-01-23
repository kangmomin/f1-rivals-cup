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
	ErrMatchNotFound   = errors.New("match not found")
	ErrDuplicateRound  = errors.New("round already exists for this league")
)

type MatchRepository struct {
	db *database.DB
}

func NewMatchRepository(db *database.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// Create creates a new match
func (r *MatchRepository) Create(ctx context.Context, match *model.Match) error {
	query := `
		INSERT INTO matches (league_id, round, track, match_date, match_time, has_sprint, sprint_date, sprint_time, sprint_status, status, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		match.LeagueID,
		match.Round,
		match.Track,
		match.MatchDate,
		match.MatchTime,
		match.HasSprint,
		match.SprintDate,
		match.SprintTime,
		match.SprintStatus,
		match.Status,
		match.Description,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "matches_league_id_round_key"` {
			return ErrDuplicateRound
		}
		return err
	}

	return nil
}

// GetByID retrieves a match by ID
func (r *MatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Match, error) {
	query := `
		SELECT id, league_id, round, track, match_date, match_time::text, has_sprint, sprint_date::text, sprint_time::text, sprint_status, status, description, created_at, updated_at
		FROM matches
		WHERE id = $1
	`

	match := &model.Match{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.LeagueID,
		&match.Round,
		&match.Track,
		&match.MatchDate,
		&match.MatchTime,
		&match.HasSprint,
		&match.SprintDate,
		&match.SprintTime,
		&match.SprintStatus,
		&match.Status,
		&match.Description,
		&match.CreatedAt,
		&match.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}

	return match, nil
}

// ListByLeague retrieves all matches for a league
func (r *MatchRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID) ([]*model.Match, error) {
	query := `
		SELECT id, league_id, round, track, match_date, match_time::text, has_sprint, sprint_date::text, sprint_time::text, sprint_status, status, description, created_at, updated_at
		FROM matches
		WHERE league_id = $1
		ORDER BY round ASC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*model.Match
	for rows.Next() {
		m := &model.Match{}
		if err := rows.Scan(
			&m.ID,
			&m.LeagueID,
			&m.Round,
			&m.Track,
			&m.MatchDate,
			&m.MatchTime,
			&m.HasSprint,
			&m.SprintDate,
			&m.SprintTime,
			&m.SprintStatus,
			&m.Status,
			&m.Description,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// Update updates a match
func (r *MatchRepository) Update(ctx context.Context, match *model.Match) error {
	query := `
		UPDATE matches
		SET round = $1, track = $2, match_date = $3, match_time = $4, has_sprint = $5, sprint_date = $6, sprint_time = $7, sprint_status = $8, status = $9, description = $10, updated_at = NOW()
		WHERE id = $11
		RETURNING updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		match.Round,
		match.Track,
		match.MatchDate,
		match.MatchTime,
		match.HasSprint,
		match.SprintDate,
		match.SprintTime,
		match.SprintStatus,
		match.Status,
		match.Description,
		match.ID,
	).Scan(&match.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMatchNotFound
		}
		if err.Error() == `pq: duplicate key value violates unique constraint "matches_league_id_round_key"` {
			return ErrDuplicateRound
		}
		return err
	}

	return nil
}

// Delete removes a match
func (r *MatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM matches WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrMatchNotFound
	}

	return nil
}

// UpdateStatus updates only the status of a match
func (r *MatchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MatchStatus) error {
	query := `
		UPDATE matches
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
		return ErrMatchNotFound
	}

	return nil
}

// UpdateSprintStatus updates only the sprint status of a match
func (r *MatchRepository) UpdateSprintStatus(ctx context.Context, id uuid.UUID, sprintStatus model.MatchStatus) error {
	query := `
		UPDATE matches
		SET sprint_status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Pool.ExecContext(ctx, query, sprintStatus, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrMatchNotFound
	}

	return nil
}
