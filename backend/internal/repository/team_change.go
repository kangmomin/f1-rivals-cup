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
	ErrTeamChangeRequestNotFound = errors.New("team change request not found")
	ErrPendingRequestExists      = errors.New("pending team change request already exists")
)

type TeamChangeRepository struct {
	db *database.DB
}

func NewTeamChangeRepository(db *database.DB) *TeamChangeRepository {
	return &TeamChangeRepository{db: db}
}

// CreateRequest creates a new team change request
func (r *TeamChangeRepository) CreateRequest(ctx context.Context, req *model.TeamChangeRequest) error {
	query := `
		INSERT INTO team_change_requests (participant_id, current_team_name, requested_team_name, status, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		req.ParticipantID,
		req.CurrentTeamName,
		req.RequestedTeamName,
		req.Status,
		req.Reason,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_team_change_requests_pending_unique"` {
			return ErrPendingRequestExists
		}
		return err
	}

	return nil
}

// GetRequestByID retrieves a team change request by ID
func (r *TeamChangeRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*model.TeamChangeRequest, error) {
	query := `
		SELECT tcr.id, tcr.participant_id, tcr.current_team_name, tcr.requested_team_name,
		       tcr.status, tcr.reason, tcr.reviewed_by, tcr.reviewed_at, tcr.created_at, tcr.updated_at,
		       u.nickname as participant_name, lp.league_id, rev.nickname as reviewer_name
		FROM team_change_requests tcr
		JOIN league_participants lp ON tcr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		LEFT JOIN users rev ON tcr.reviewed_by = rev.id
		WHERE tcr.id = $1
	`

	req := &model.TeamChangeRequest{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.ParticipantID,
		&req.CurrentTeamName,
		&req.RequestedTeamName,
		&req.Status,
		&req.Reason,
		&req.ReviewedBy,
		&req.ReviewedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.ParticipantName,
		&req.LeagueID,
		&req.ReviewerName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamChangeRequestNotFound
		}
		return nil, err
	}

	return req, nil
}

// ListRequestsByLeague retrieves all team change requests for a league
func (r *TeamChangeRepository) ListRequestsByLeague(ctx context.Context, leagueID uuid.UUID, status string) ([]*model.TeamChangeRequest, error) {
	query := `
		SELECT tcr.id, tcr.participant_id, tcr.current_team_name, tcr.requested_team_name,
		       tcr.status, tcr.reason, tcr.reviewed_by, tcr.reviewed_at, tcr.created_at, tcr.updated_at,
		       u.nickname as participant_name, lp.league_id, rev.nickname as reviewer_name
		FROM team_change_requests tcr
		JOIN league_participants lp ON tcr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		LEFT JOIN users rev ON tcr.reviewed_by = rev.id
		WHERE lp.league_id = $1 AND ($2 = '' OR tcr.status = $2)
		ORDER BY tcr.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.TeamChangeRequest
	for rows.Next() {
		req := &model.TeamChangeRequest{}
		if err := rows.Scan(
			&req.ID,
			&req.ParticipantID,
			&req.CurrentTeamName,
			&req.RequestedTeamName,
			&req.Status,
			&req.Reason,
			&req.ReviewedBy,
			&req.ReviewedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.ParticipantName,
			&req.LeagueID,
			&req.ReviewerName,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// ListRequestsByParticipant retrieves all team change requests for a participant
func (r *TeamChangeRepository) ListRequestsByParticipant(ctx context.Context, participantID uuid.UUID) ([]*model.TeamChangeRequest, error) {
	query := `
		SELECT tcr.id, tcr.participant_id, tcr.current_team_name, tcr.requested_team_name,
		       tcr.status, tcr.reason, tcr.reviewed_by, tcr.reviewed_at, tcr.created_at, tcr.updated_at,
		       u.nickname as participant_name, lp.league_id, rev.nickname as reviewer_name
		FROM team_change_requests tcr
		JOIN league_participants lp ON tcr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		LEFT JOIN users rev ON tcr.reviewed_by = rev.id
		WHERE tcr.participant_id = $1
		ORDER BY tcr.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, participantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.TeamChangeRequest
	for rows.Next() {
		req := &model.TeamChangeRequest{}
		if err := rows.Scan(
			&req.ID,
			&req.ParticipantID,
			&req.CurrentTeamName,
			&req.RequestedTeamName,
			&req.Status,
			&req.Reason,
			&req.ReviewedBy,
			&req.ReviewedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.ParticipantName,
			&req.LeagueID,
			&req.ReviewerName,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetPendingRequestByParticipant gets any pending request for a participant
func (r *TeamChangeRepository) GetPendingRequestByParticipant(ctx context.Context, participantID uuid.UUID) (*model.TeamChangeRequest, error) {
	query := `
		SELECT tcr.id, tcr.participant_id, tcr.current_team_name, tcr.requested_team_name,
		       tcr.status, tcr.reason, tcr.reviewed_by, tcr.reviewed_at, tcr.created_at, tcr.updated_at,
		       u.nickname as participant_name, lp.league_id
		FROM team_change_requests tcr
		JOIN league_participants lp ON tcr.participant_id = lp.id
		JOIN users u ON lp.user_id = u.id
		WHERE tcr.participant_id = $1 AND tcr.status = 'pending'
	`

	req := &model.TeamChangeRequest{}
	err := r.db.Pool.QueryRowContext(ctx, query, participantID).Scan(
		&req.ID,
		&req.ParticipantID,
		&req.CurrentTeamName,
		&req.RequestedTeamName,
		&req.Status,
		&req.Reason,
		&req.ReviewedBy,
		&req.ReviewedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.ParticipantName,
		&req.LeagueID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamChangeRequestNotFound
		}
		return nil, err
	}

	return req, nil
}

// UpdateRequestStatus updates the status of a team change request
func (r *TeamChangeRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status model.TeamChangeRequestStatus, reviewedBy uuid.UUID, reason *string) error {
	query := `
		UPDATE team_change_requests
		SET status = $1, reviewed_by = $2, reviewed_at = NOW(), reason = COALESCE($3, reason), updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Pool.ExecContext(ctx, query, status, reviewedBy, reason, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamChangeRequestNotFound
	}

	return nil
}

// DeleteRequest deletes a team change request (only pending requests can be deleted)
func (r *TeamChangeRepository) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM team_change_requests WHERE id = $1 AND status = 'pending'`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamChangeRequestNotFound
	}

	return nil
}


// ApproveTeamChange approves a team change request and updates participant's team_name
func (r *TeamChangeRepository) ApproveTeamChange(ctx context.Context, requestID uuid.UUID, reviewedBy uuid.UUID) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the request details
	var participantID uuid.UUID
	var requestedTeamName string
	err = tx.QueryRowContext(ctx, `
		SELECT participant_id, requested_team_name
		FROM team_change_requests
		WHERE id = $1 AND status = 'pending'
	`, requestID).Scan(&participantID, &requestedTeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamChangeRequestNotFound
		}
		return err
	}

	// Update participant's team_name
	_, err = tx.ExecContext(ctx, `
		UPDATE league_participants
		SET team_name = $1, updated_at = NOW()
		WHERE id = $2
	`, requestedTeamName, participantID)
	if err != nil {
		return err
	}

	// Update request status to approved
	_, err = tx.ExecContext(ctx, `
		UPDATE team_change_requests
		SET status = 'approved', reviewed_by = $1, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, reviewedBy, requestID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RejectTeamChange rejects a team change request
func (r *TeamChangeRepository) RejectTeamChange(ctx context.Context, requestID uuid.UUID, reviewedBy uuid.UUID, reason *string) error {
	query := `
		UPDATE team_change_requests
		SET status = 'rejected', reviewed_by = $1, reviewed_at = NOW(), reason = COALESCE($2, reason), updated_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`

	result, err := r.db.Pool.ExecContext(ctx, query, reviewedBy, reason, requestID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamChangeRequestNotFound
	}

	return nil
}
