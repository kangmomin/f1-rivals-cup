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
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidTransactionID = errors.New("invalid transaction id")
)

type TransactionRepository struct {
	db *database.DB
}

func NewTransactionRepository(db *database.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction and updates account balances atomically
// useBalance: true=잔액 지출(기본), false=비잔액 지출(FIA만, 잔액 차감 없이 화폐 발행)
func (r *TransactionRepository) Create(ctx context.Context, tx *model.Transaction, useBalance bool) error {
	dbTx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	if useBalance {
		// 잔액 지출: from 계좌에서 잔액 차감 (음수 잔액 허용)
		updateFromQuery := `
			UPDATE accounts
			SET balance = balance - $2, updated_at = NOW()
			WHERE id = $1
			RETURNING balance
		`
		var newFromBalance int64
		if err := dbTx.QueryRowContext(ctx, updateFromQuery, tx.FromAccountID, tx.Amount).Scan(&newFromBalance); err != nil {
			return err
		}
	}
	// 비잔액 지출(useBalance=false): from 계좌 잔액 변동 없음 (화폐 발행 개념)

	// Lock and update to account (add)
	updateToQuery := `
		UPDATE accounts
		SET balance = balance + $2, updated_at = NOW()
		WHERE id = $1
	`
	if _, err := dbTx.ExecContext(ctx, updateToQuery, tx.ToAccountID, tx.Amount); err != nil {
		return err
	}

	// Insert transaction record
	insertQuery := `
		INSERT INTO transactions (league_id, from_account_id, to_account_id, amount, category, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	if err := dbTx.QueryRowContext(ctx, insertQuery,
		tx.LeagueID,
		tx.FromAccountID,
		tx.ToAccountID,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.CreatedBy,
	).Scan(&tx.ID, &tx.CreatedAt); err != nil {
		return err
	}

	return dbTx.Commit()
}

// getOwnerName returns a case expression for owner name
func getOwnerNameCase(alias string, accountAlias string) string {
	return `
		CASE
			WHEN ` + accountAlias + `.owner_type = 'team' THEN (SELECT name FROM teams WHERE id = ` + accountAlias + `.owner_id)
			WHEN ` + accountAlias + `.owner_type = 'participant' THEN (SELECT u.nickname FROM league_participants lp JOIN users u ON lp.user_id = u.id WHERE lp.id = ` + accountAlias + `.owner_id)
			WHEN ` + accountAlias + `.owner_type = 'system' THEN 'FIA'
			ELSE ''
		END as ` + alias
}

// ListByLeague retrieves all transactions for a league with pagination
func (r *TransactionRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID, page, pageSize int) ([]*model.Transaction, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM transactions WHERE league_id = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, leagueID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT
			t.id, t.league_id, t.from_account_id, t.to_account_id, t.amount, t.category,
			t.description, t.created_by, t.created_at,
			` + getOwnerNameCase("from_name", "fa") + `,
			` + getOwnerNameCase("to_name", "ta") + `
		FROM transactions t
		JOIN accounts fa ON t.from_account_id = fa.id
		JOIN accounts ta ON t.to_account_id = ta.id
		WHERE t.league_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		tx := &model.Transaction{}
		var fromName, toName sql.NullString
		if err := rows.Scan(
			&tx.ID,
			&tx.LeagueID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Category,
			&tx.Description,
			&tx.CreatedBy,
			&tx.CreatedAt,
			&fromName,
			&toName,
		); err != nil {
			return nil, 0, err
		}
		if fromName.Valid {
			tx.FromName = fromName.String
		}
		if toName.Valid {
			tx.ToName = toName.String
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// ListByAccount retrieves all transactions for an account
func (r *TransactionRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*model.Transaction, error) {
	query := `
		SELECT
			t.id, t.league_id, t.from_account_id, t.to_account_id, t.amount, t.category,
			t.description, t.created_by, t.created_at,
			` + getOwnerNameCase("from_name", "fa") + `,
			` + getOwnerNameCase("to_name", "ta") + `
		FROM transactions t
		JOIN accounts fa ON t.from_account_id = fa.id
		JOIN accounts ta ON t.to_account_id = ta.id
		WHERE t.from_account_id = $1 OR t.to_account_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		tx := &model.Transaction{}
		var fromName, toName sql.NullString
		if err := rows.Scan(
			&tx.ID,
			&tx.LeagueID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Category,
			&tx.Description,
			&tx.CreatedBy,
			&tx.CreatedAt,
			&fromName,
			&toName,
		); err != nil {
			return nil, err
		}
		if fromName.Valid {
			tx.FromName = fromName.String
		}
		if toName.Valid {
			tx.ToName = toName.String
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetAccountWeeklyFlow retrieves weekly income/expense flow for a specific account (last 12 weeks)
func (r *TransactionRepository) GetAccountWeeklyFlow(ctx context.Context, accountID uuid.UUID) ([]model.WeeklyFlow, error) {
	query := `
		SELECT
			TO_CHAR(t.created_at, 'IYYY-IW') as week,
			COALESCE(SUM(CASE WHEN t.to_account_id = $1 THEN t.amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.from_account_id = $1 THEN t.amount ELSE 0 END), 0) as expense
		FROM transactions t
		WHERE (t.from_account_id = $1 OR t.to_account_id = $1)
		  AND t.created_at >= NOW() - INTERVAL '12 weeks'
		GROUP BY TO_CHAR(t.created_at, 'IYYY-IW')
		ORDER BY week
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weeklyFlow []model.WeeklyFlow
	for rows.Next() {
		var wf model.WeeklyFlow
		if err := rows.Scan(&wf.Week, &wf.Income, &wf.Expense); err != nil {
			return nil, err
		}
		weeklyFlow = append(weeklyFlow, wf)
	}

	return weeklyFlow, nil
}

// GetFinanceStats retrieves finance statistics for a league
func (r *TransactionRepository) GetFinanceStats(ctx context.Context, leagueID uuid.UUID) (*model.FinanceStatsResponse, error) {
	stats := &model.FinanceStatsResponse{
		CategoryTotals: make(map[string]int64),
	}

	// Get total circulation (sum of all transactions)
	circulationQuery := `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE league_id = $1`
	if err := r.db.Pool.QueryRowContext(ctx, circulationQuery, leagueID).Scan(&stats.TotalCirculation); err != nil {
		return nil, err
	}

	// Get team balances
	teamBalanceQuery := `
		SELECT a.owner_id, t.name, a.balance
		FROM accounts a
		JOIN teams t ON a.owner_id = t.id
		WHERE a.league_id = $1 AND a.owner_type = 'team'
		ORDER BY a.balance DESC
	`
	rows, err := r.db.Pool.QueryContext(ctx, teamBalanceQuery, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tb model.TeamBalance
		if err := rows.Scan(&tb.TeamID, &tb.TeamName, &tb.Balance); err != nil {
			return nil, err
		}
		stats.TeamBalances = append(stats.TeamBalances, tb)
	}

	// Get category totals
	categoryQuery := `
		SELECT category, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE league_id = $1
		GROUP BY category
	`
	catRows, err := r.db.Pool.QueryContext(ctx, categoryQuery, leagueID)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	for catRows.Next() {
		var category string
		var total int64
		if err := catRows.Scan(&category, &total); err != nil {
			return nil, err
		}
		stats.CategoryTotals[category] = total
	}

	// Get weekly flow (last 12 weeks)
	weeklyQuery := `
		SELECT
			TO_CHAR(t.created_at, 'IYYY-IW') as week,
			COALESCE(SUM(CASE WHEN fa.owner_type = 'system' THEN t.amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN ta.owner_type = 'system' THEN t.amount ELSE 0 END), 0) as expense
		FROM transactions t
		JOIN accounts fa ON t.from_account_id = fa.id
		JOIN accounts ta ON t.to_account_id = ta.id
		WHERE t.league_id = $1
		  AND t.created_at >= NOW() - INTERVAL '12 weeks'
		GROUP BY TO_CHAR(t.created_at, 'IYYY-IW')
		ORDER BY week
	`
	weekRows, err := r.db.Pool.QueryContext(ctx, weeklyQuery, leagueID)
	if err != nil {
		return nil, err
	}
	defer weekRows.Close()

	for weekRows.Next() {
		var wf model.WeeklyFlow
		if err := weekRows.Scan(&wf.Week, &wf.Income, &wf.Expense); err != nil {
			return nil, err
		}
		stats.WeeklyFlow = append(stats.WeeklyFlow, wf)
	}

	return stats, nil
}
