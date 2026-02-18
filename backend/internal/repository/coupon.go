package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrCouponNotFound    = errors.New("coupon not found")
	ErrCouponExpired     = errors.New("coupon expired")
	ErrCouponMaxUsed     = errors.New("coupon max uses reached")
	ErrCouponCodeExists  = errors.New("coupon code already exists for this product")
	ErrCouponInvalid     = errors.New("invalid coupon")
)

type CouponRepository struct {
	db *database.DB
}

func NewCouponRepository(db *database.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func generateCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// Create creates a new coupon. If Code is empty, generates a random 8-char code.
func (r *CouponRepository) Create(ctx context.Context, coupon *model.Coupon) error {
	if coupon.Code == "" {
		coupon.Code = generateCode(8)
	}
	coupon.Code = strings.ToUpper(coupon.Code)

	err := r.db.Pool.QueryRowContext(ctx, `
		INSERT INTO coupons (product_id, code, discount_type, discount_value, max_uses, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, coupon.ProductID, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUses, coupon.ExpiresAt).
		Scan(&coupon.ID, &coupon.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return ErrCouponCodeExists
		}
		return err
	}
	return nil
}

// GetByID retrieves a coupon by ID with product name.
func (r *CouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Coupon, error) {
	c := &model.Coupon{}
	err := r.db.Pool.QueryRowContext(ctx, `
		SELECT c.id, c.product_id, c.code, c.discount_type, c.discount_value,
			c.max_uses, c.used_count, c.expires_at, c.created_at,
			p.name AS product_name
		FROM coupons c
		JOIN products p ON c.product_id = p.id
		WHERE c.id = $1
	`, id).Scan(
		&c.ID, &c.ProductID, &c.Code, &c.DiscountType, &c.DiscountValue,
		&c.MaxUses, &c.UsedCount, &c.ExpiresAt, &c.CreatedAt,
		&c.ProductName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return c, nil
}

// ListByProduct lists coupons for a specific product with pagination.
func (r *CouponRepository) ListByProduct(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*model.Coupon, int, error) {
	var total int
	err := r.db.Pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM coupons WHERE product_id = $1", productID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.QueryContext(ctx, `
		SELECT c.id, c.product_id, c.code, c.discount_type, c.discount_value,
			c.max_uses, c.used_count, c.expires_at, c.created_at,
			p.name AS product_name
		FROM coupons c
		JOIN products p ON c.product_id = p.id
		WHERE c.product_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`, productID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var coupons []*model.Coupon
	for rows.Next() {
		c := &model.Coupon{}
		if err := rows.Scan(
			&c.ID, &c.ProductID, &c.Code, &c.DiscountType, &c.DiscountValue,
			&c.MaxUses, &c.UsedCount, &c.ExpiresAt, &c.CreatedAt,
			&c.ProductName,
		); err != nil {
			return nil, 0, err
		}
		coupons = append(coupons, c)
	}
	return coupons, total, nil
}

// ListBySeller lists all coupons for products owned by the given seller.
func (r *CouponRepository) ListBySeller(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*model.Coupon, int, error) {
	var total int
	err := r.db.Pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM coupons c
		JOIN products p ON c.product_id = p.id
		WHERE p.seller_id = $1
	`, sellerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.QueryContext(ctx, `
		SELECT c.id, c.product_id, c.code, c.discount_type, c.discount_value,
			c.max_uses, c.used_count, c.expires_at, c.created_at,
			p.name AS product_name
		FROM coupons c
		JOIN products p ON c.product_id = p.id
		WHERE p.seller_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`, sellerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var coupons []*model.Coupon
	for rows.Next() {
		c := &model.Coupon{}
		if err := rows.Scan(
			&c.ID, &c.ProductID, &c.Code, &c.DiscountType, &c.DiscountValue,
			&c.MaxUses, &c.UsedCount, &c.ExpiresAt, &c.CreatedAt,
			&c.ProductName,
		); err != nil {
			return nil, 0, err
		}
		coupons = append(coupons, c)
	}
	return coupons, total, nil
}

// Delete deletes a coupon by ID.
func (r *CouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Pool.ExecContext(ctx, "DELETE FROM coupons WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCouponNotFound
	}
	return nil
}

// GetByCodeAndProduct retrieves a coupon by code and product ID.
func (r *CouponRepository) GetByCodeAndProduct(ctx context.Context, code string, productID uuid.UUID) (*model.Coupon, error) {
	c := &model.Coupon{}
	err := r.db.Pool.QueryRowContext(ctx, `
		SELECT id, product_id, code, discount_type, discount_value,
			max_uses, used_count, expires_at, created_at
		FROM coupons
		WHERE code = $1 AND product_id = $2
	`, strings.ToUpper(code), productID).Scan(
		&c.ID, &c.ProductID, &c.Code, &c.DiscountType, &c.DiscountValue,
		&c.MaxUses, &c.UsedCount, &c.ExpiresAt, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return c, nil
}

// CalculateDiscount computes the discount amount for a given coupon and base price.
func CalculateDiscount(coupon *model.Coupon, basePrice int64) int64 {
	switch coupon.DiscountType {
	case "fixed":
		if coupon.DiscountValue > basePrice {
			return basePrice
		}
		return coupon.DiscountValue
	case "percentage":
		return basePrice * coupon.DiscountValue / 100
	default:
		return 0
	}
}

// ValidateCoupon checks if a coupon is valid (not expired, not maxed out).
func ValidateCoupon(coupon *model.Coupon) error {
	if time.Now().After(coupon.ExpiresAt) {
		return ErrCouponExpired
	}
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return ErrCouponMaxUsed
	}
	return nil
}

// RecordUsage inserts a coupon_usage row and increments used_count within a transaction.
func (r *CouponRepository) RecordUsage(ctx context.Context, tx *sql.Tx, couponID, userID, subscriptionID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO coupon_usages (coupon_id, user_id, subscription_id)
		VALUES ($1, $2, $3)
	`, couponID, userID, subscriptionID)
	if err != nil {
		return fmt.Errorf("insert coupon usage: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE coupons SET used_count = used_count + 1 WHERE id = $1
	`, couponID)
	if err != nil {
		return fmt.Errorf("update coupon used_count: %w", err)
	}
	return nil
}
