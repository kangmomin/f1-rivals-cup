package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrLeagueNotFound = errors.New("league not found")
)

// LeagueRepository handles league database operations
type LeagueRepository struct {
	db *database.DB
}

// NewLeagueRepository creates a new LeagueRepository
func NewLeagueRepository(db *database.DB) *LeagueRepository {
	return &LeagueRepository{db: db}
}

// Create creates a new league
func (r *LeagueRepository) Create(ctx context.Context, league *model.League) error {
	query := `
		INSERT INTO leagues (name, description, status, season, created_by, start_date, end_date, match_time, rules, settings, contact_info)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		league.Name,
		league.Description,
		league.Status,
		league.Season,
		league.CreatedBy,
		league.StartDate,
		league.EndDate,
		league.MatchTime,
		league.Rules,
		league.Settings,
		league.ContactInfo,
	).Scan(&league.ID, &league.CreatedAt, &league.UpdatedAt)

	return err
}

// GetByID retrieves a league by ID
func (r *LeagueRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.League, error) {
	query := `
		SELECT id, name, description, status, season, created_by, start_date, end_date, match_time::text, rules, settings, contact_info, created_at, updated_at
		FROM leagues
		WHERE id = $1
	`

	league := &model.League{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&league.ID,
		&league.Name,
		&league.Description,
		&league.Status,
		&league.Season,
		&league.CreatedBy,
		&league.StartDate,
		&league.EndDate,
		&league.MatchTime,
		&league.Rules,
		&league.Settings,
		&league.ContactInfo,
		&league.CreatedAt,
		&league.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	return league, nil
}

// List retrieves a paginated list of leagues
func (r *LeagueRepository) List(ctx context.Context, page, limit int, status string) ([]*model.League, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM leagues WHERE ($1 = '' OR status = $1)`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get leagues
	query := `
		SELECT id, name, description, status, season, created_by, start_date, end_date, match_time::text, rules, settings, contact_info, created_at, updated_at
		FROM leagues
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var leagues []*model.League
	for rows.Next() {
		league := &model.League{}
		if err := rows.Scan(
			&league.ID,
			&league.Name,
			&league.Description,
			&league.Status,
			&league.Season,
			&league.CreatedBy,
			&league.StartDate,
			&league.EndDate,
			&league.MatchTime,
			&league.Rules,
			&league.Settings,
			&league.ContactInfo,
			&league.CreatedAt,
			&league.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		leagues = append(leagues, league)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return leagues, total, nil
}

// Update updates a league
func (r *LeagueRepository) Update(ctx context.Context, league *model.League) error {
	query := `
		UPDATE leagues
		SET name = $1, description = $2, status = $3, season = $4, start_date = $5, end_date = $6, match_time = $7, rules = $8, settings = $9, contact_info = $10, updated_at = NOW()
		WHERE id = $11
	`

	result, err := r.db.Pool.ExecContext(ctx, query,
		league.Name,
		league.Description,
		league.Status,
		league.Season,
		league.StartDate,
		league.EndDate,
		league.MatchTime,
		league.Rules,
		league.Settings,
		league.ContactInfo,
		league.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrLeagueNotFound
	}

	return nil
}

// Delete deletes a league
func (r *LeagueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM leagues WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrLeagueNotFound
	}

	return nil
}

// CountByStatus counts leagues by status
func (r *LeagueRepository) CountByStatus(ctx context.Context, status string) (int, error) {
	query := `SELECT COUNT(*) FROM leagues WHERE ($1 = '' OR status = $1)`
	var count int
	err := r.db.Pool.QueryRowContext(ctx, query, status).Scan(&count)
	return count, err
}

// ParseTime parses a time string to time.Time
func ParseTime(s string) (*time.Time, error) {
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
