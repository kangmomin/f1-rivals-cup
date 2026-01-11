package repository

import (
	"context"
	"encoding/json"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

// PermissionHistoryRepository handles permission history database operations
type PermissionHistoryRepository struct {
	db *database.DB
}

// NewPermissionHistoryRepository creates a new PermissionHistoryRepository
func NewPermissionHistoryRepository(db *database.DB) *PermissionHistoryRepository {
	return &PermissionHistoryRepository{db: db}
}

// Create creates a new permission history record
func (r *PermissionHistoryRepository) Create(ctx context.Context, history *model.PermissionHistory) error {
	oldValueJSON, err := json.Marshal(history.OldValue)
	if err != nil {
		return err
	}
	newValueJSON, err := json.Marshal(history.NewValue)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO permission_history (changer_id, target_id, change_type, old_value, new_value)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb)
		RETURNING id, created_at
	`

	err = r.db.Pool.QueryRowContext(ctx, query,
		history.ChangerID,
		history.TargetID,
		history.ChangeType,
		string(oldValueJSON),
		string(newValueJSON),
	).Scan(&history.ID, &history.CreatedAt)

	return err
}

// GetByTargetID retrieves permission history for a specific user
func (r *PermissionHistoryRepository) GetByTargetID(ctx context.Context, targetID uuid.UUID, page, limit int) ([]*model.PermissionHistory, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM permission_history WHERE target_id = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, targetID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get history with joined nicknames
	query := `
		SELECT ph.id, ph.changer_id, ph.target_id, ph.change_type, ph.old_value, ph.new_value, ph.created_at,
		       u1.nickname as changer_nickname, u2.nickname as target_nickname
		FROM permission_history ph
		LEFT JOIN users u1 ON ph.changer_id = u1.id
		LEFT JOIN users u2 ON ph.target_id = u2.id
		WHERE ph.target_id = $1
		ORDER BY ph.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, targetID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []*model.PermissionHistory
	for rows.Next() {
		history := &model.PermissionHistory{}
		var oldValueJSON, newValueJSON []byte
		if err := rows.Scan(
			&history.ID,
			&history.ChangerID,
			&history.TargetID,
			&history.ChangeType,
			&oldValueJSON,
			&newValueJSON,
			&history.CreatedAt,
			&history.ChangerNickname,
			&history.TargetNickname,
		); err != nil {
			return nil, 0, err
		}
		// Parse JSON values
		json.Unmarshal(oldValueJSON, &history.OldValue)
		json.Unmarshal(newValueJSON, &history.NewValue)
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetByChangerID retrieves permission history made by a specific admin
func (r *PermissionHistoryRepository) GetByChangerID(ctx context.Context, changerID uuid.UUID, page, limit int) ([]*model.PermissionHistory, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM permission_history WHERE changer_id = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, changerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get history with joined nicknames
	query := `
		SELECT ph.id, ph.changer_id, ph.target_id, ph.change_type, ph.old_value, ph.new_value, ph.created_at,
		       u1.nickname as changer_nickname, u2.nickname as target_nickname
		FROM permission_history ph
		LEFT JOIN users u1 ON ph.changer_id = u1.id
		LEFT JOIN users u2 ON ph.target_id = u2.id
		WHERE ph.changer_id = $1
		ORDER BY ph.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, changerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []*model.PermissionHistory
	for rows.Next() {
		history := &model.PermissionHistory{}
		var oldValueJSON, newValueJSON []byte
		if err := rows.Scan(
			&history.ID,
			&history.ChangerID,
			&history.TargetID,
			&history.ChangeType,
			&oldValueJSON,
			&newValueJSON,
			&history.CreatedAt,
			&history.ChangerNickname,
			&history.TargetNickname,
		); err != nil {
			return nil, 0, err
		}
		// Parse JSON values
		json.Unmarshal(oldValueJSON, &history.OldValue)
		json.Unmarshal(newValueJSON, &history.NewValue)
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetRecentHistory retrieves the most recent permission changes
func (r *PermissionHistoryRepository) GetRecentHistory(ctx context.Context, limit int) ([]*model.PermissionHistory, error) {
	query := `
		SELECT ph.id, ph.changer_id, ph.target_id, ph.change_type, ph.old_value, ph.new_value, ph.created_at,
		       u1.nickname as changer_nickname, u2.nickname as target_nickname
		FROM permission_history ph
		LEFT JOIN users u1 ON ph.changer_id = u1.id
		LEFT JOIN users u2 ON ph.target_id = u2.id
		ORDER BY ph.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*model.PermissionHistory
	for rows.Next() {
		history := &model.PermissionHistory{}
		var oldValueJSON, newValueJSON []byte
		if err := rows.Scan(
			&history.ID,
			&history.ChangerID,
			&history.TargetID,
			&history.ChangeType,
			&oldValueJSON,
			&newValueJSON,
			&history.CreatedAt,
			&history.ChangerNickname,
			&history.TargetNickname,
		); err != nil {
			return nil, err
		}
		// Parse JSON values
		json.Unmarshal(oldValueJSON, &history.OldValue)
		json.Unmarshal(newValueJSON, &history.NewValue)
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}
