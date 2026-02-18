package model

import (
	"time"

	"github.com/google/uuid"
)

type Coupon struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discount_type"`  // "fixed" or "percentage"
	DiscountValue int64     `json:"discount_value"`
	MaxUses       int       `json:"max_uses"`
	UsedCount     int       `json:"used_count"`
	OncePerUser   bool      `json:"once_per_user"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`

	// Joined fields
	ProductName string `json:"product_name,omitempty"`
}

type CreateCouponRequest struct {
	ProductID     uuid.UUID `json:"product_id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discount_type" validate:"required,oneof=fixed percentage"`
	DiscountValue int64     `json:"discount_value" validate:"required,min=1"`
	MaxUses       int       `json:"max_uses" validate:"min=0"`
	OncePerUser   bool      `json:"once_per_user"`
	ExpiresAt     time.Time `json:"expires_at" validate:"required"`
}

type CouponUsage struct {
	ID             uuid.UUID `json:"id"`
	CouponID       uuid.UUID `json:"coupon_id"`
	UserID         uuid.UUID `json:"user_id"`
	SubscriptionID uuid.UUID `json:"subscription_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type ValidateCouponRequest struct {
	Code      string    `json:"code" validate:"required"`
	ProductID uuid.UUID `json:"product_id" validate:"required"`
}

type ValidateCouponResponse struct {
	Valid          bool   `json:"valid"`
	DiscountAmount int64  `json:"discount_amount"`
	Message        string `json:"message"`
}
