package repository

import (
	"context"
	"encoding/json"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

type TeamChangeActivityRepository struct {
	db *database.DB
}

func NewTeamChangeActivityRepository(db *database.DB) *TeamChangeActivityRepository {
	return &TeamChangeActivityRepository{db: db}
}

// Create inserts a new activity log entry
func (r *TeamChangeActivityRepository) Create(ctx context.Context, log *model.TeamChangeActivityLog) error {
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO team_change_activity_log (actor_id, request_id, participant_id, action_type, details)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.db.Pool.QueryRowContext(ctx, query,
		log.ActorID,
		log.RequestID,
		log.ParticipantID,
		log.ActionType,
		detailsJSON,
	).Scan(&log.ID, &log.CreatedAt)
}

// ListByRequest retrieves all activity logs for a specific request
func (r *TeamChangeActivityRepository) ListByRequest(ctx context.Context, requestID uuid.UUID) ([]*model.TeamChangeActivityLog, error) {
	query := `
		SELECT
			al.id, al.actor_id, al.request_id, al.participant_id, al.action_type, al.details, al.created_at,
			actor.nickname as actor_nickname,
			participant_user.nickname as participant_nickname
		FROM team_change_activity_log al
		JOIN users actor ON al.actor_id = actor.id
		JOIN league_participants lp ON al.participant_id = lp.id
		JOIN users participant_user ON lp.user_id = participant_user.id
		WHERE al.request_id = $1
		ORDER BY al.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.TeamChangeActivityLog
	for rows.Next() {
		log := &model.TeamChangeActivityLog{}
		var detailsJSON []byte
		if err := rows.Scan(
			&log.ID,
			&log.ActorID,
			&log.RequestID,
			&log.ParticipantID,
			&log.ActionType,
			&detailsJSON,
			&log.CreatedAt,
			&log.ActorNickname,
			&log.ParticipantNickname,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = make(map[string]any)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// ListByLeague retrieves all activity logs for a league with pagination
func (r *TeamChangeActivityRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID, page, limit int) ([]*model.TeamChangeActivityLog, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM team_change_activity_log al
		JOIN league_participants lp ON al.participant_id = lp.id
		WHERE lp.league_id = $1
	`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, leagueID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	query := `
		SELECT
			al.id, al.actor_id, al.request_id, al.participant_id, al.action_type, al.details, al.created_at,
			actor.nickname as actor_nickname,
			participant_user.nickname as participant_nickname
		FROM team_change_activity_log al
		JOIN users actor ON al.actor_id = actor.id
		JOIN league_participants lp ON al.participant_id = lp.id
		JOIN users participant_user ON lp.user_id = participant_user.id
		WHERE lp.league_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.TeamChangeActivityLog
	for rows.Next() {
		log := &model.TeamChangeActivityLog{}
		var detailsJSON []byte
		if err := rows.Scan(
			&log.ID,
			&log.ActorID,
			&log.RequestID,
			&log.ParticipantID,
			&log.ActionType,
			&detailsJSON,
			&log.CreatedAt,
			&log.ActorNickname,
			&log.ParticipantNickname,
		); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = make(map[string]any)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
