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
	ErrAccountNotFound = errors.New("account not found")
)

type AccountRepository struct {
	db *database.DB
}

func NewAccountRepository(db *database.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// ListByLeague retrieves all accounts for a league with owner names
func (r *AccountRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID) ([]*model.Account, error) {
	query := `
		SELECT
			a.id, a.league_id, a.owner_id, a.owner_type, a.balance, a.created_at, a.updated_at,
			CASE
				WHEN a.owner_type = 'team' THEN (SELECT name FROM teams WHERE id = a.owner_id)
				WHEN a.owner_type = 'participant' THEN (SELECT u.nickname FROM league_participants lp JOIN users u ON lp.user_id = u.id WHERE lp.id = a.owner_id)
				WHEN a.owner_type = 'system' THEN 'FIA'
				ELSE ''
			END as owner_name
		FROM accounts a
		WHERE a.league_id = $1
		ORDER BY a.owner_type, a.created_at
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		a := &model.Account{}
		var ownerName sql.NullString
		if err := rows.Scan(
			&a.ID,
			&a.LeagueID,
			&a.OwnerID,
			&a.OwnerType,
			&a.Balance,
			&a.CreatedAt,
			&a.UpdatedAt,
			&ownerName,
		); err != nil {
			return nil, err
		}
		if ownerName.Valid {
			a.OwnerName = ownerName.String
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

// GetByID retrieves an account by ID with owner name
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	query := `
		SELECT
			a.id, a.league_id, a.owner_id, a.owner_type, a.balance, a.created_at, a.updated_at,
			CASE
				WHEN a.owner_type = 'team' THEN (SELECT name FROM teams WHERE id = a.owner_id)
				WHEN a.owner_type = 'participant' THEN (SELECT u.nickname FROM league_participants lp JOIN users u ON lp.user_id = u.id WHERE lp.id = a.owner_id)
				WHEN a.owner_type = 'system' THEN 'FIA'
				ELSE ''
			END as owner_name
		FROM accounts a
		WHERE a.id = $1
	`

	a := &model.Account{}
	var ownerName sql.NullString
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&a.ID,
		&a.LeagueID,
		&a.OwnerID,
		&a.OwnerType,
		&a.Balance,
		&a.CreatedAt,
		&a.UpdatedAt,
		&ownerName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	if ownerName.Valid {
		a.OwnerName = ownerName.String
	}

	return a, nil
}

// Create creates a new account
func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	query := `
		INSERT INTO accounts (league_id, owner_id, owner_type, balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.db.Pool.QueryRowContext(ctx, query,
		account.LeagueID,
		account.OwnerID,
		account.OwnerType,
		account.Balance,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

// UpdateBalance updates an account's balance
func (r *AccountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance int64) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool.ExecContext(ctx, query, id, balance)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrAccountNotFound
	}

	return nil
}

// GetOrCreateSystemAccount gets or creates the system (FIA) account for a league
func (r *AccountRepository) GetOrCreateSystemAccount(ctx context.Context, leagueID uuid.UUID) (*model.Account, error) {
	// System account uses a deterministic UUID based on league ID
	// We use a fixed namespace UUID for system accounts
	systemOwnerID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("system-"+leagueID.String()))

	query := `
		SELECT id, league_id, owner_id, owner_type, balance, created_at, updated_at
		FROM accounts
		WHERE league_id = $1 AND owner_type = 'system'
		LIMIT 1
	`

	account := &model.Account{}
	err := r.db.Pool.QueryRowContext(ctx, query, leagueID).Scan(
		&account.ID,
		&account.LeagueID,
		&account.OwnerID,
		&account.OwnerType,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == nil {
		account.OwnerName = "FIA"
		return account, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Create system account
	account = &model.Account{
		LeagueID:  leagueID,
		OwnerID:   systemOwnerID,
		OwnerType: model.OwnerTypeSystem,
		Balance:   0,
	}

	if err := r.Create(ctx, account); err != nil {
		return nil, err
	}

	account.OwnerName = "FIA"
	return account, nil
}

// GetByOwner retrieves an account by owner
func (r *AccountRepository) GetByOwner(ctx context.Context, leagueID, ownerID uuid.UUID, ownerType model.OwnerType) (*model.Account, error) {
	query := `
		SELECT id, league_id, owner_id, owner_type, balance, created_at, updated_at
		FROM accounts
		WHERE league_id = $1 AND owner_id = $2 AND owner_type = $3
	`

	account := &model.Account{}
	err := r.db.Pool.QueryRowContext(ctx, query, leagueID, ownerID, ownerType).Scan(
		&account.ID,
		&account.LeagueID,
		&account.OwnerID,
		&account.OwnerType,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

// EnsureParticipantAccount gets or creates a participant account
// This is idempotent and handles the case where account creation failed during approval
func (r *AccountRepository) EnsureParticipantAccount(ctx context.Context, leagueID, participantID uuid.UUID) (*model.Account, error) {
	// Try to get existing account
	account, err := r.GetByOwner(ctx, leagueID, participantID, model.OwnerTypeParticipant)
	if err == nil {
		// GetByOwner doesn't include owner_name, so fetch complete account info
		return r.GetByID(ctx, account.ID)
	}
	if !errors.Is(err, ErrAccountNotFound) {
		return nil, err
	}

	// Create new account for participant
	newAccount := &model.Account{
		LeagueID:  leagueID,
		OwnerID:   participantID,
		OwnerType: model.OwnerTypeParticipant,
		Balance:   0,
	}

	if err := r.Create(ctx, newAccount); err != nil {
		return nil, err
	}

	// Fetch complete account info including owner_name
	return r.GetByID(ctx, newAccount.ID)
}
