package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrOAuthAccountNotFound = errors.New("oauth account not found")
	ErrOAuthAccountExists   = errors.New("oauth account already linked")
)

// OAuthAccountRepository handles OAuth account database operations
type OAuthAccountRepository struct {
	db *database.DB
}

// NewOAuthAccountRepository creates a new OAuthAccountRepository
func NewOAuthAccountRepository(db *database.DB) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

// Create creates a new OAuth account link
func (r *OAuthAccountRepository) Create(ctx context.Context, account *model.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_id, provider_username, provider_avatar, provider_email, access_token, refresh_token, token_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		account.UserID,
		account.Provider,
		account.ProviderID,
		account.ProviderUsername,
		account.ProviderAvatar,
		account.ProviderEmail,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrOAuthAccountExists
		}
		return err
	}

	return nil
}

// GetByProviderID retrieves an OAuth account by provider and provider ID
func (r *OAuthAccountRepository) GetByProviderID(ctx context.Context, provider, providerID string) (*model.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, provider_username, provider_avatar, provider_email,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_id = $2
	`

	account := &model.OAuthAccount{}
	err := r.db.Pool.QueryRowContext(ctx, query, provider, providerID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderID,
		&account.ProviderUsername,
		&account.ProviderAvatar,
		&account.ProviderEmail,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOAuthAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

// GetByUserAndProvider retrieves an OAuth account by user ID and provider
func (r *OAuthAccountRepository) GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*model.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, provider_username, provider_avatar, provider_email,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1 AND provider = $2
	`

	account := &model.OAuthAccount{}
	err := r.db.Pool.QueryRowContext(ctx, query, userID, provider).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderID,
		&account.ProviderUsername,
		&account.ProviderAvatar,
		&account.ProviderEmail,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOAuthAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

// DeleteByUserAndProvider deletes an OAuth account by user ID and provider
func (r *OAuthAccountRepository) DeleteByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	query := `DELETE FROM oauth_accounts WHERE user_id = $1 AND provider = $2`

	result, err := r.db.Pool.ExecContext(ctx, query, userID, provider)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOAuthAccountNotFound
	}

	return nil
}

// ListByUser retrieves all OAuth accounts for a user
func (r *OAuthAccountRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, provider_username, provider_avatar, provider_email,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.OAuthAccount
	for rows.Next() {
		account := &model.OAuthAccount{}
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderID,
			&account.ProviderUsername,
			&account.ProviderAvatar,
			&account.ProviderEmail,
			&account.AccessToken,
			&account.RefreshToken,
			&account.TokenExpiresAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// Update updates an OAuth account's provider info and tokens
func (r *OAuthAccountRepository) Update(ctx context.Context, account *model.OAuthAccount) error {
	query := `
		UPDATE oauth_accounts
		SET provider_username = $1, provider_avatar = $2, provider_email = $3,
		    access_token = $4, refresh_token = $5, token_expires_at = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.Pool.ExecContext(ctx, query,
		account.ProviderUsername,
		account.ProviderAvatar,
		account.ProviderEmail,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOAuthAccountNotFound
	}

	return nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key")
}
