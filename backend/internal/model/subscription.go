package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	LeagueID      uuid.UUID  `json:"league_id"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	Status        string     `json:"status"` // "active", "expired", "cancelled"
	StartedAt     time.Time  `json:"started_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`

	// Joined fields
	ProductName   string `json:"product_name,omitempty"`
	LeagueName    string `json:"league_name,omitempty"`
	BuyerNickname string `json:"buyer_nickname,omitempty"`
	ProductPrice  *int64 `json:"product_price,omitempty"`
}

type SubscribeRequest struct {
	ProductID  uuid.UUID  `json:"product_id" validate:"required"`
	LeagueID   uuid.UUID  `json:"league_id" validate:"required"`
	OptionID   *uuid.UUID `json:"option_id,omitempty"`
	CouponCode *string    `json:"coupon_code,omitempty"`
}
