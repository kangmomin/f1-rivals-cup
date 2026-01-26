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
	ErrMatchResultNotFound = errors.New("match result not found")
)

type MatchResultRepository struct {
	db *database.DB
}

func NewMatchResultRepository(db *database.DB) *MatchResultRepository {
	return &MatchResultRepository{db: db}
}

// Upsert creates or updates a match result
func (r *MatchResultRepository) Upsert(ctx context.Context, result *model.MatchResult) error {
	query := `
		INSERT INTO match_results (match_id, participant_id, team_name, position, points, fastest_lap, dnf, dnf_reason, sprint_position, sprint_points)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (match_id, participant_id)
		DO UPDATE SET
			team_name = COALESCE(EXCLUDED.team_name, match_results.team_name),
			position = EXCLUDED.position,
			points = EXCLUDED.points,
			fastest_lap = EXCLUDED.fastest_lap,
			dnf = EXCLUDED.dnf,
			dnf_reason = EXCLUDED.dnf_reason,
			sprint_position = EXCLUDED.sprint_position,
			sprint_points = EXCLUDED.sprint_points,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		result.MatchID,
		result.ParticipantID,
		result.StoredTeamName,
		result.Position,
		result.Points,
		result.FastestLap,
		result.DNF,
		result.DNFReason,
		result.SprintPosition,
		result.SprintPoints,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

	return err
}

// GetByID retrieves a match result by ID
func (r *MatchResultRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MatchResult, error) {
	query := `
		SELECT mr.id, mr.match_id, mr.participant_id, mr.team_name, mr.position, mr.points, mr.fastest_lap,
		       mr.dnf, mr.dnf_reason, mr.sprint_position, mr.sprint_points, mr.created_at, mr.updated_at,
		       u.nickname, lp.team_name
		FROM match_results mr
		JOIN league_participants lp ON mr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		WHERE mr.id = $1
	`

	result := &model.MatchResult{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.MatchID,
		&result.ParticipantID,
		&result.StoredTeamName,
		&result.Position,
		&result.Points,
		&result.FastestLap,
		&result.DNF,
		&result.DNFReason,
		&result.SprintPosition,
		&result.SprintPoints,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.ParticipantName,
		&result.TeamName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMatchResultNotFound
		}
		return nil, err
	}

	return result, nil
}

// ListByMatch retrieves all results for a match
func (r *MatchResultRepository) ListByMatch(ctx context.Context, matchID uuid.UUID) ([]*model.MatchResult, error) {
	query := `
		SELECT mr.id, mr.match_id, mr.participant_id, mr.team_name, mr.position, mr.points, mr.fastest_lap,
		       mr.dnf, mr.dnf_reason, mr.sprint_position, mr.sprint_points, mr.created_at, mr.updated_at,
		       u.nickname, lp.team_name
		FROM match_results mr
		JOIN league_participants lp ON mr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		WHERE mr.match_id = $1
		ORDER BY
			CASE WHEN mr.position IS NULL THEN 1 ELSE 0 END,
			mr.position ASC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.MatchResult
	for rows.Next() {
		r := &model.MatchResult{}
		if err := rows.Scan(
			&r.ID,
			&r.MatchID,
			&r.ParticipantID,
			&r.StoredTeamName,
			&r.Position,
			&r.Points,
			&r.FastestLap,
			&r.DNF,
			&r.DNFReason,
			&r.SprintPosition,
			&r.SprintPoints,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.ParticipantName,
			&r.TeamName,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// Delete removes a match result
func (r *MatchResultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM match_results WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrMatchResultNotFound
	}

	return nil
}

// DeleteByMatch removes all results for a match
func (r *MatchResultRepository) DeleteByMatch(ctx context.Context, matchID uuid.UUID) error {
	query := `DELETE FROM match_results WHERE match_id = $1`
	_, err := r.db.Pool.ExecContext(ctx, query, matchID)
	return err
}

// GetLeagueStandings returns aggregated standings for a league
func (r *MatchResultRepository) GetLeagueStandings(ctx context.Context, leagueID uuid.UUID) ([]model.StandingsEntry, error) {
	query := `
		SELECT
			lp.id as participant_id,
			lp.user_id,
			u.nickname as driver_name,
			lp.team_name,
			COALESCE(SUM(mr.points), 0) + COALESCE(SUM(mr.sprint_points), 0) as total_points,
			COALESCE(SUM(mr.points), 0) as race_points,
			COALESCE(SUM(mr.sprint_points), 0) as sprint_points,
			COUNT(CASE WHEN mr.position = 1 THEN 1 END) as wins,
			COUNT(CASE WHEN mr.position <= 3 AND mr.position IS NOT NULL THEN 1 END) as podiums,
			COUNT(CASE WHEN mr.fastest_lap = true THEN 1 END) as fastest_laps,
			COUNT(CASE WHEN mr.dnf = true THEN 1 END) as dnfs,
			COUNT(CASE WHEN mr.position IS NOT NULL OR mr.dnf = true THEN 1 END) as races_completed
		FROM league_participants lp
		JOIN users u ON lp.user_id = u.id
		LEFT JOIN match_results mr ON lp.id = mr.participant_id
		LEFT JOIN matches m ON mr.match_id = m.id AND m.league_id = $1
		WHERE lp.league_id = $1
		  AND lp.status = 'approved'
		  AND 'player' = ANY(lp.roles)
		GROUP BY lp.id, lp.user_id, u.nickname, lp.team_name
		ORDER BY total_points DESC, wins DESC, podiums DESC, fastest_laps DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standings []model.StandingsEntry
	rank := 0
	for rows.Next() {
		rank++
		entry := model.StandingsEntry{Rank: rank}
		if err := rows.Scan(
			&entry.ParticipantID,
			&entry.UserID,
			&entry.DriverName,
			&entry.TeamName,
			&entry.TotalPoints,
			&entry.RacePoints,
			&entry.SprintPoints,
			&entry.Wins,
			&entry.Podiums,
			&entry.FastestLaps,
			&entry.DNFs,
			&entry.RacesCompleted,
		); err != nil {
			return nil, err
		}
		standings = append(standings, entry)
	}

	return standings, nil
}

// GetTeamStandings returns aggregated team standings for a league
// Uses team_name stored in match_results at the time of result recording
func (r *MatchResultRepository) GetTeamStandings(ctx context.Context, leagueID uuid.UUID) ([]model.TeamStandingsEntry, error) {
	query := `
		SELECT
			mr.team_name,
			COALESCE(SUM(mr.points), 0) + COALESCE(SUM(mr.sprint_points), 0) as total_points,
			COALESCE(SUM(mr.points), 0) as race_points,
			COALESCE(SUM(mr.sprint_points), 0) as sprint_points,
			COUNT(CASE WHEN mr.position = 1 THEN 1 END) as wins,
			COUNT(CASE WHEN mr.position <= 3 AND mr.position IS NOT NULL THEN 1 END) as podiums,
			COUNT(CASE WHEN mr.fastest_lap = true THEN 1 END) as fastest_laps,
			COUNT(CASE WHEN mr.dnf = true THEN 1 END) as dnfs,
			COUNT(DISTINCT lp.id) as driver_count
		FROM match_results mr
		JOIN matches m ON mr.match_id = m.id
		JOIN league_participants lp ON mr.participant_id = lp.id
		WHERE m.league_id = $1
		  AND lp.status = 'approved'
		  AND 'player' = ANY(lp.roles)
		  AND mr.team_name IS NOT NULL
		  AND mr.team_name != ''
		GROUP BY mr.team_name
		ORDER BY total_points DESC, wins DESC, podiums DESC, fastest_laps DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standings []model.TeamStandingsEntry
	rank := 0
	for rows.Next() {
		rank++
		entry := model.TeamStandingsEntry{Rank: rank}
		if err := rows.Scan(
			&entry.TeamName,
			&entry.TotalPoints,
			&entry.RacePoints,
			&entry.SprintPoints,
			&entry.Wins,
			&entry.Podiums,
			&entry.FastestLaps,
			&entry.DNFs,
			&entry.DriverCount,
		); err != nil {
			return nil, err
		}
		standings = append(standings, entry)
	}

	return standings, nil
}

// BulkUpsert creates or updates multiple results at once (all fields)
func (r *MatchResultRepository) BulkUpsert(ctx context.Context, matchID uuid.UUID, results []model.CreateMatchResultRequest) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO match_results (match_id, participant_id, team_name, position, points, fastest_lap, dnf, dnf_reason, sprint_position, sprint_points)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (match_id, participant_id)
		DO UPDATE SET
			team_name = COALESCE(EXCLUDED.team_name, match_results.team_name),
			position = EXCLUDED.position,
			points = EXCLUDED.points,
			fastest_lap = EXCLUDED.fastest_lap,
			dnf = EXCLUDED.dnf,
			dnf_reason = EXCLUDED.dnf_reason,
			sprint_position = EXCLUDED.sprint_position,
			sprint_points = EXCLUDED.sprint_points,
			updated_at = NOW()
	`

	for _, result := range results {
		_, err := tx.ExecContext(ctx, query,
			matchID,
			result.ParticipantID,
			result.TeamName,
			result.Position,
			result.Points,
			result.FastestLap,
			result.DNF,
			result.DNFReason,
			result.SprintPosition,
			result.SprintPoints,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BulkUpsertSprintResults creates or updates sprint results only (sprint_position, sprint_points)
func (r *MatchResultRepository) BulkUpsertSprintResults(ctx context.Context, matchID uuid.UUID, results []model.CreateMatchResultRequest) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO match_results (match_id, participant_id, team_name, sprint_position, sprint_points)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (match_id, participant_id)
		DO UPDATE SET
			team_name = COALESCE(EXCLUDED.team_name, match_results.team_name),
			sprint_position = EXCLUDED.sprint_position,
			sprint_points = EXCLUDED.sprint_points,
			updated_at = NOW()
	`

	for _, result := range results {
		_, err := tx.ExecContext(ctx, query,
			matchID,
			result.ParticipantID,
			result.TeamName,
			result.SprintPosition,
			result.SprintPoints,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BulkUpsertRaceResults creates or updates race results only (position, points, fastest_lap, dnf, dnf_reason)
func (r *MatchResultRepository) BulkUpsertRaceResults(ctx context.Context, matchID uuid.UUID, results []model.CreateMatchResultRequest) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO match_results (match_id, participant_id, team_name, position, points, fastest_lap, dnf, dnf_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (match_id, participant_id)
		DO UPDATE SET
			team_name = COALESCE(EXCLUDED.team_name, match_results.team_name),
			position = EXCLUDED.position,
			points = EXCLUDED.points,
			fastest_lap = EXCLUDED.fastest_lap,
			dnf = EXCLUDED.dnf,
			dnf_reason = EXCLUDED.dnf_reason,
			updated_at = NOW()
	`

	for _, result := range results {
		_, err := tx.ExecContext(ctx, query,
			matchID,
			result.ParticipantID,
			result.TeamName,
			result.Position,
			result.Points,
			result.FastestLap,
			result.DNF,
			result.DNFReason,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
