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
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

// RefreshTokenRepository handles refresh token database operations
type RefreshTokenRepository struct {
	db *database.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository
func NewRefreshTokenRepository(db *database.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, rt *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, device_info, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, last_used_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		rt.UserID,
		rt.Token,
		rt.DeviceInfo,
		rt.IPAddress,
		rt.ExpiresAt,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.LastUsedAt)

	return err
}

// GetByToken retrieves a refresh token by its token value
func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, device_info, ip_address, expires_at, created_at, last_used_at
		FROM refresh_tokens
		WHERE token = $1
	`

	rt := &model.RefreshToken{}
	err := r.db.Pool.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.DeviceInfo,
		&rt.IPAddress,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	// Check if token is expired
	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	return rt, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *RefreshTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.ExecContext(ctx, query, id)
	return err
}

// Delete deletes a refresh token by ID
func (r *RefreshTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`
	_, err := r.db.Pool.ExecContext(ctx, query, id)
	return err
}

// DeleteByToken deletes a refresh token by its token value
func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.Pool.ExecContext(ctx, query, token)
	return err
}

// DeleteAllByUserID deletes all refresh tokens for a user (logout from all devices)
func (r *RefreshTokenRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Pool.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	result, err := r.db.Pool.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ListByUserID lists all refresh tokens for a user
func (r *RefreshTokenRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, device_info, ip_address, expires_at, created_at, last_used_at
		FROM refresh_tokens
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*model.RefreshToken
	for rows.Next() {
		rt := &model.RefreshToken{}
		if err := rows.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.Token,
			&rt.DeviceInfo,
			&rt.IPAddress,
			&rt.ExpiresAt,
			&rt.CreatedAt,
			&rt.LastUsedAt,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, rt)
	}

	return tokens, rows.Err()
}

// CountByUserID returns the number of active refresh tokens for a user
func (r *RefreshTokenRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND expires_at > NOW()`
	var count int
	err := r.db.Pool.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// ErrTokenAlreadyUsed is returned when a refresh token has already been consumed
var ErrTokenAlreadyUsed = errors.New("refresh token already used or expired")

// RotateToken deletes the old token and creates a new one atomically
// Uses DELETE ... RETURNING to ensure single-use: if zero rows are returned,
// the token was already consumed by a concurrent request (race condition protection)
func (r *RefreshTokenRepository) RotateToken(ctx context.Context, oldToken string, newRT *model.RefreshToken) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old token and verify it existed (single-use enforcement)
	// Using DELETE ... RETURNING to atomically check and delete
	var deletedID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`DELETE FROM refresh_tokens WHERE token = $1 RETURNING id`,
		oldToken,
	).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Token was already consumed by another request (race condition)
			// or doesn't exist
			return ErrTokenAlreadyUsed
		}
		return err
	}

	// Create new token
	query := `
		INSERT INTO refresh_tokens (user_id, token, device_info, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, last_used_at
	`
	err = tx.QueryRowContext(ctx, query,
		newRT.UserID,
		newRT.Token,
		newRT.DeviceInfo,
		newRT.IPAddress,
		newRT.ExpiresAt,
	).Scan(&newRT.ID, &newRT.CreatedAt, &newRT.LastUsedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}
