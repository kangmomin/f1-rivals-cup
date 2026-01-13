package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FinanceHandler struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	leagueRepo      *repository.LeagueRepository
}

func NewFinanceHandler(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	leagueRepo *repository.LeagueRepository,
) *FinanceHandler {
	return &FinanceHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		leagueRepo:      leagueRepo,
	}
}

// ListAccounts handles GET /api/v1/leagues/:id/accounts
func (h *FinanceHandler) ListAccounts(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListAccounts: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	// Ensure system account exists
	_, err = h.accountRepo.GetOrCreateSystemAccount(ctx, leagueID)
	if err != nil {
		slog.Error("Finance.ListAccounts: failed to get/create system account", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "시스템 계좌를 생성하는데 실패했습니다",
		})
	}

	accounts, err := h.accountRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		slog.Error("Finance.ListAccounts: failed to list accounts", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 목록을 불러오는데 실패했습니다",
		})
	}

	if accounts == nil {
		accounts = []*model.Account{}
	}

	return c.JSON(http.StatusOK, model.ListAccountsResponse{
		Accounts: accounts,
		Total:    len(accounts),
	})
}

// GetAccount handles GET /api/v1/accounts/:id
func (h *FinanceHandler) GetAccount(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 계좌 ID입니다",
		})
	}

	ctx := c.Request().Context()

	account, err := h.accountRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.GetAccount: failed to get account", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, account)
}

// SetAccountBalance handles PUT /api/v1/admin/accounts/:id/balance
func (h *FinanceHandler) SetAccountBalance(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 계좌 ID입니다",
		})
	}

	var req model.SetBalanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if account exists
	account, err := h.accountRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.SetAccountBalance: failed to get account", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}

	if err := h.accountRepo.UpdateBalance(ctx, id, req.Balance); err != nil {
		slog.Error("Finance.SetAccountBalance: failed to update balance", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "잔액 수정에 실패했습니다",
		})
	}

	account.Balance = req.Balance
	return c.JSON(http.StatusOK, account)
}

// CreateTransaction handles POST /api/v1/admin/leagues/:id/transactions
func (h *FinanceHandler) CreateTransaction(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	var req model.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate required fields
	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "금액은 0보다 커야 합니다",
		})
	}

	if req.FromAccountID == req.ToAccountID {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "출금 계좌와 입금 계좌가 같을 수 없습니다",
		})
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.CreateTransaction: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	// Verify from account exists and belongs to this league
	fromAccount, err := h.accountRepo.GetByID(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "출금 계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.CreateTransaction: failed to get from account", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}
	if fromAccount.LeagueID != leagueID {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "출금 계좌가 해당 리그에 속하지 않습니다",
		})
	}

	// Verify to account exists and belongs to this league
	toAccount, err := h.accountRepo.GetByID(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "입금 계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.CreateTransaction: failed to get to account", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}
	if toAccount.LeagueID != leagueID {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "입금 계좌가 해당 리그에 속하지 않습니다",
		})
	}

	// Get current user ID from context
	var createdBy *uuid.UUID
	if userID, ok := c.Get("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	transaction := &model.Transaction{
		LeagueID:      leagueID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Category:      req.Category,
		Description:   req.Description,
		CreatedBy:     createdBy,
	}

	if err := h.transactionRepo.Create(ctx, transaction); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "잔액이 부족합니다",
			})
		}
		slog.Error("Finance.CreateTransaction: failed to create transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "거래 생성에 실패했습니다",
		})
	}

	// Set names for response
	transaction.FromName = fromAccount.OwnerName
	transaction.ToName = toAccount.OwnerName

	return c.JSON(http.StatusCreated, transaction)
}

// ListTransactions handles GET /api/v1/leagues/:id/transactions
func (h *FinanceHandler) ListTransactions(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	// Parse pagination
	page := 1
	pageSize := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListTransactions: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	transactions, total, err := h.transactionRepo.ListByLeague(ctx, leagueID, page, pageSize)
	if err != nil {
		slog.Error("Finance.ListTransactions: failed to list transactions", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "거래 목록을 불러오는데 실패했습니다",
		})
	}

	if transactions == nil {
		transactions = []*model.Transaction{}
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, model.ListTransactionsResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		TotalPages:   totalPages,
	})
}

// ListAccountTransactions handles GET /api/v1/accounts/:id/transactions
func (h *FinanceHandler) ListAccountTransactions(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 계좌 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if account exists and get balance
	account, err := h.accountRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListAccountTransactions: failed to get account", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}

	transactions, err := h.transactionRepo.ListByAccount(ctx, id)
	if err != nil {
		slog.Error("Finance.ListAccountTransactions: failed to list transactions", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "거래 목록을 불러오는데 실패했습니다",
		})
	}

	if transactions == nil {
		transactions = []*model.Transaction{}
	}

	return c.JSON(http.StatusOK, model.AccountTransactionsResponse{
		Transactions: transactions,
		Total:        len(transactions),
		Balance:      account.Balance,
	})
}

// GetFinanceStats handles GET /api/v1/leagues/:id/finance/stats
func (h *FinanceHandler) GetFinanceStats(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.GetFinanceStats: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	stats, err := h.transactionRepo.GetFinanceStats(ctx, leagueID)
	if err != nil {
		slog.Error("Finance.GetFinanceStats: failed to get finance stats", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "통계를 불러오는데 실패했습니다",
		})
	}

	// Ensure non-nil slices
	if stats.TeamBalances == nil {
		stats.TeamBalances = []model.TeamBalance{}
	}
	if stats.MonthlyFlow == nil {
		stats.MonthlyFlow = []model.MonthlyFlow{}
	}
	if stats.CategoryTotals == nil {
		stats.CategoryTotals = make(map[string]int64)
	}

	return c.JSON(http.StatusOK, stats)
}
