package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FinanceHandler struct {
	financeService *service.FinanceService
}

func NewFinanceHandler(financeService *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{
		financeService: financeService,
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

	accounts, err := h.financeService.ListAccounts(ctx, leagueID)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListAccounts: failed to list accounts", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 목록을 불러오는데 실패했습니다",
		})
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

	account, err := h.financeService.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
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

	account, err := h.financeService.SetAccountBalance(ctx, id, req.Balance)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.SetAccountBalance: failed to update balance", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "잔액 수정에 실패했습니다",
		})
	}

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

	ctx := c.Request().Context()

	// Get current user ID from context
	var createdBy *uuid.UUID
	if userID, ok := c.Get("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	serviceReq := &service.CreateTransactionRequest{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Category:      req.Category,
		Description:   req.Description,
		UseBalance:    req.UseBalance,
	}

	transaction, err := h.financeService.CreateTransaction(ctx, leagueID, serviceReq, createdBy)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLeagueNotFound):
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		case errors.Is(err, service.ErrAccountNotFound):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "계좌를 찾을 수 없습니다",
			})
		case errors.Is(err, service.ErrAccountNotInLeague):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "계좌가 해당 리그에 속하지 않습니다",
			})
		case errors.Is(err, service.ErrAmountMustBePositive):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "금액은 0보다 커야 합니다",
			})
		case errors.Is(err, service.ErrSameAccount):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "출금 계좌와 입금 계좌가 같을 수 없습니다",
			})
		case errors.Is(err, service.ErrInsufficientBalance):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "잔액이 부족합니다",
			})
		default:
			slog.Error("Finance.CreateTransaction: failed to create transaction", "error", err)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "거래 생성에 실패했습니다",
			})
		}
	}

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

	transactions, total, err := h.financeService.ListTransactions(ctx, leagueID, page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListTransactions: failed to list transactions", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "거래 목록을 불러오는데 실패했습니다",
		})
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

	transactions, account, err := h.financeService.ListAccountTransactions(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.ListAccountTransactions: failed to list transactions", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "거래 목록을 불러오는데 실패했습니다",
		})
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

	stats, err := h.financeService.GetFinanceStats(ctx, leagueID)
	if err != nil {
		if errors.Is(err, service.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Finance.GetFinanceStats: failed to get finance stats", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "통계를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, model.FinanceStatsResponse{
		TotalCirculation: stats.TotalCirculation,
		TeamBalances:     stats.TeamBalances,
		CategoryTotals:   stats.CategoryTotals,
		MonthlyFlow:      stats.MonthlyFlow,
	})
}

// CreateTransactionByDirector handles POST /api/v1/leagues/:id/transactions
// Allows directors to create transactions from their team accounts,
// and participants to create transactions from their own accounts
func (h *FinanceHandler) CreateTransactionByDirector(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	// Get current user ID from context
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	var req model.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	serviceReq := &service.CreateTransactionRequest{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Category:      req.Category,
		Description:   req.Description,
		UseBalance:    req.UseBalance,
	}

	transaction, err := h.financeService.CreateTransactionByDirector(ctx, leagueID, serviceReq, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLeagueNotFound):
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		case errors.Is(err, service.ErrAccountNotFound):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "계좌를 찾을 수 없습니다",
			})
		case errors.Is(err, service.ErrAccountNotInLeague):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "계좌가 해당 리그에 속하지 않습니다",
			})
		case errors.Is(err, service.ErrAmountMustBePositive):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "금액은 0보다 커야 합니다",
			})
		case errors.Is(err, service.ErrSameAccount):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "출금 계좌와 입금 계좌가 같을 수 없습니다",
			})
		case errors.Is(err, service.ErrNotDirector):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "감독 권한이 없습니다",
			})
		case errors.Is(err, service.ErrNotTeamDirector):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "본인 소속 팀의 계좌에서만 거래를 생성할 수 있습니다",
			})
		case errors.Is(err, service.ErrParticipantNotFound):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "해당 리그에 참여하고 있지 않습니다",
			})
		case errors.Is(err, service.ErrNotAccountOwner):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "본인 계좌에서만 거래를 생성할 수 있습니다",
			})
		case errors.Is(err, service.ErrNotApproved):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "승인된 참가자만 거래를 생성할 수 있습니다",
			})
		case errors.Is(err, service.ErrSystemAccountForbidden):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "시스템 계좌에서는 거래를 생성할 수 없습니다",
			})
		case errors.Is(err, service.ErrInsufficientBalance):
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "잔액이 부족합니다",
			})
		default:
			slog.Error("Finance.CreateTransactionByDirector: failed to create transaction", "error", err)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "거래 생성에 실패했습니다",
			})
		}
	}

	return c.JSON(http.StatusCreated, transaction)
}

// GetMyAccount handles GET /api/v1/leagues/:id/my-account
// Returns the current user's participant account, creating it if it doesn't exist
func (h *FinanceHandler) GetMyAccount(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	// Get current user ID from context
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	account, err := h.financeService.GetMyAccount(ctx, leagueID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrParticipantNotFound):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "해당 리그에 참여하고 있지 않습니다",
			})
		case errors.Is(err, service.ErrNotApproved):
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "승인된 참가자만 계좌를 조회할 수 있습니다",
			})
		default:
			slog.Error("Finance.GetMyAccount: failed to get my account", "error", err, "user_id", userID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "계좌 정보를 불러오는데 실패했습니다",
			})
		}
	}

	return c.JSON(http.StatusOK, account)
}
