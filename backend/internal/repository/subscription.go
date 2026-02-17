package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrAlreadySubscribed    = errors.New("already subscribed to this product")
)

type SubscriptionRepository struct {
	db *database.DB
}

func NewSubscriptionRepository(db *database.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Subscribe creates or extends a subscription within a single DB transaction.
// It handles: balance deduction, seller payment, transaction record, subscription upsert, and permission grant.
func (r *SubscriptionRepository) Subscribe(
	ctx context.Context,
	userID, productID, leagueID, buyerAccountID, sellerAccountID uuid.UUID,
	totalPrice int64,
	durationDays int,
	description string,
) (*model.Subscription, error) {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Deduct buyer balance
	var newBalance int64
	err = tx.QueryRowContext(ctx, `
		UPDATE accounts SET balance = balance - $2, updated_at = NOW()
		WHERE id = $1
		RETURNING balance
	`, buyerAccountID, totalPrice).Scan(&newBalance)
	if err != nil {
		return nil, err
	}
	if newBalance < 0 {
		return nil, ErrInsufficientBalance
	}

	// 2. Credit seller balance
	_, err = tx.ExecContext(ctx, `
		UPDATE accounts SET balance = balance + $2, updated_at = NOW()
		WHERE id = $1
	`, sellerAccountID, totalPrice)
	if err != nil {
		return nil, err
	}

	// 3. Insert transaction record
	var txID uuid.UUID
	var txCreatedAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (league_id, from_account_id, to_account_id, amount, category, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, leagueID, buyerAccountID, sellerAccountID, totalPrice, model.CategoryPurchase, description, userID).Scan(&txID, &txCreatedAt)
	if err != nil {
		return nil, err
	}

	// 4. Check existing active subscription (lock row)
	var existingSub struct {
		ID        uuid.UUID
		ExpiresAt time.Time
	}
	err = tx.QueryRowContext(ctx, `
		SELECT id, expires_at FROM subscriptions
		WHERE user_id = $1 AND product_id = $2 AND status = 'active'
		FOR UPDATE
	`, userID, productID).Scan(&existingSub.ID, &existingSub.ExpiresAt)

	duration := time.Duration(durationDays) * 24 * time.Hour
	var sub *model.Subscription

	if err == nil {
		// Existing subscription: extend expiry
		sub = &model.Subscription{}
		newExpiry := existingSub.ExpiresAt.Add(duration)
		err = tx.QueryRowContext(ctx, `
			UPDATE subscriptions SET expires_at = $2, transaction_id = $3
			WHERE id = $1
			RETURNING id, user_id, product_id, league_id, transaction_id, status, started_at, expires_at, created_at
		`, existingSub.ID, newExpiry, txID).Scan(
			&sub.ID, &sub.UserID, &sub.ProductID, &sub.LeagueID,
			&sub.TransactionID, &sub.Status, &sub.StartedAt, &sub.ExpiresAt, &sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		// New subscription
		sub = &model.Subscription{}
		newExpiry := time.Now().Add(duration)
		err = tx.QueryRowContext(ctx, `
			INSERT INTO subscriptions (user_id, product_id, league_id, transaction_id, status, expires_at)
			VALUES ($1, $2, $3, $4, 'active', $5)
			RETURNING id, user_id, product_id, league_id, transaction_id, status, started_at, expires_at, created_at
		`, userID, productID, leagueID, txID, newExpiry).Scan(
			&sub.ID, &sub.UserID, &sub.ProductID, &sub.LeagueID,
			&sub.TransactionID, &sub.Status, &sub.StartedAt, &sub.ExpiresAt, &sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 5. Add permission to user
		permKey := fmt.Sprintf("product.%s", productID.String())
		_, err = tx.ExecContext(ctx, `
			UPDATE users
			SET permissions = CASE
				WHEN NOT (permissions @> $2::jsonb) THEN permissions || $2::jsonb
				ELSE permissions
			END,
			updated_at = NOW()
			WHERE id = $1
		`, userID, fmt.Sprintf(`["%s"]`, permKey))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return sub, nil
}

// GetActiveByUser returns the user's active subscriptions with product and league names.
func (r *SubscriptionRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*model.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.product_id, s.league_id, s.transaction_id, s.status,
			s.started_at, s.expires_at, s.created_at,
			p.name AS product_name, l.name AS league_name
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		JOIN leagues l ON s.league_id = l.id
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.expires_at ASC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		s := &model.Subscription{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.ProductID, &s.LeagueID, &s.TransactionID, &s.Status,
			&s.StartedAt, &s.ExpiresAt, &s.CreatedAt,
			&s.ProductName, &s.LeagueName,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return subs, nil
}

// GetByID retrieves a subscription by ID.
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.product_id, s.league_id, s.transaction_id, s.status,
			s.started_at, s.expires_at, s.created_at,
			p.name AS product_name, l.name AS league_name
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		JOIN leagues l ON s.league_id = l.id
		WHERE s.id = $1
	`

	s := &model.Subscription{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.ProductID, &s.LeagueID, &s.TransactionID, &s.Status,
		&s.StartedAt, &s.ExpiresAt, &s.CreatedAt,
		&s.ProductName, &s.LeagueName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	return s, nil
}

// GetByProductSeller returns all subscriptions (sales) for products owned by the given seller.
func (r *SubscriptionRepository) GetByProductSeller(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*model.Subscription, int, error) {
	var total int
	err := r.db.Pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		WHERE p.seller_id = $1
	`, sellerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.id, s.user_id, s.product_id, s.league_id, s.transaction_id,
			s.status, s.started_at, s.expires_at, s.created_at,
			p.name AS product_name, l.name AS league_name,
			buyer.nickname AS buyer_nickname, p.price AS product_price
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		JOIN leagues l ON s.league_id = l.id
		JOIN users buyer ON s.user_id = buyer.id
		WHERE p.seller_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, sellerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		s := &model.Subscription{}
		var price int64
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.ProductID, &s.LeagueID, &s.TransactionID, &s.Status,
			&s.StartedAt, &s.ExpiresAt, &s.CreatedAt,
			&s.ProductName, &s.LeagueName,
			&s.BuyerNickname, &price,
		); err != nil {
			return nil, 0, err
		}
		s.ProductPrice = &price
		subs = append(subs, s)
	}

	return subs, total, nil
}

// ExpireSubscriptions finds all active subscriptions past their expiry,
// marks them expired, and removes the corresponding permissions from users.
func (r *SubscriptionRepository) ExpireSubscriptions(ctx context.Context) (int, error) {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Select expired subscriptions with row lock
	rows, err := tx.QueryContext(ctx, `
		SELECT id, user_id, product_id FROM subscriptions
		WHERE status = 'active' AND expires_at <= NOW()
		FOR UPDATE
	`)
	if err != nil {
		return 0, err
	}

	type expiredSub struct {
		ID        uuid.UUID
		UserID    uuid.UUID
		ProductID uuid.UUID
	}
	var expired []expiredSub

	for rows.Next() {
		var es expiredSub
		if err := rows.Scan(&es.ID, &es.UserID, &es.ProductID); err != nil {
			rows.Close()
			return 0, err
		}
		expired = append(expired, es)
	}
	rows.Close()

	if len(expired) == 0 {
		return 0, nil
	}

	for _, es := range expired {
		// Mark as expired
		_, err := tx.ExecContext(ctx, `
			UPDATE subscriptions SET status = 'expired' WHERE id = $1
		`, es.ID)
		if err != nil {
			return 0, err
		}

		// Remove permission from user
		permKey := fmt.Sprintf("product.%s", es.ProductID.String())
		_, err = tx.ExecContext(ctx, `
			UPDATE users SET permissions = permissions - $2, updated_at = NOW()
			WHERE id = $1
		`, es.UserID, permKey)
		if err != nil {
			slog.Error("ExpireSubscriptions: failed to remove permission",
				"user_id", es.UserID, "product_id", es.ProductID, "error", err)
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(expired), nil
}
