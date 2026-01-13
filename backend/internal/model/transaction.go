package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionCategory string

const (
	CategoryPrize       TransactionCategory = "prize"
	CategoryTransfer    TransactionCategory = "transfer"
	CategoryPenalty     TransactionCategory = "penalty"
	CategorySponsorship TransactionCategory = "sponsorship"
	CategoryOther       TransactionCategory = "other"
)

type Transaction struct {
	ID            uuid.UUID           `json:"id"`
	LeagueID      uuid.UUID           `json:"league_id"`
	FromAccountID uuid.UUID           `json:"from_account_id"`
	ToAccountID   uuid.UUID           `json:"to_account_id"`
	Amount        int64               `json:"amount"`
	Category      TransactionCategory `json:"category"`
	Description   *string             `json:"description,omitempty"`
	CreatedBy     *uuid.UUID          `json:"created_by,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`

	// Joined fields
	FromName string `json:"from_name,omitempty"`
	ToName   string `json:"to_name,omitempty"`
}

type CreateTransactionRequest struct {
	FromAccountID uuid.UUID           `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID           `json:"to_account_id" validate:"required"`
	Amount        int64               `json:"amount" validate:"required,gt=0"`
	Category      TransactionCategory `json:"category" validate:"required"`
	Description   *string             `json:"description,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int            `json:"total"`
	Page         int            `json:"page"`
	TotalPages   int            `json:"total_pages"`
}

type AccountTransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int            `json:"total"`
	Balance      int64          `json:"balance"`
}

// Finance stats for graphs
type TeamBalance struct {
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
	Balance  int64     `json:"balance"`
}

type MonthlyFlow struct {
	Month   string `json:"month"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

type FinanceStatsResponse struct {
	TotalCirculation int64            `json:"total_circulation"`
	TeamBalances     []TeamBalance    `json:"team_balances"`
	CategoryTotals   map[string]int64 `json:"category_totals"`
	MonthlyFlow      []MonthlyFlow    `json:"monthly_flow"`
}
