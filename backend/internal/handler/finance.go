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
	participantRepo *repository.ParticipantRepository
	teamRepo        *repository.TeamRepository
}

func NewFinanceHandler(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	leagueRepo *repository.LeagueRepository,
	participantRepo *repository.ParticipantRepository,
	teamRepo *repository.TeamRepository,
) *FinanceHandler {
	return &FinanceHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		leagueRepo:      leagueRepo,
		participantRepo: participantRepo,
		teamRepo:        teamRepo,
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

	// FIA(system) 계좌에서 출금 시 UseBalance 옵션 적용
	// UseBalance: nil/true=잔액 지출(기본), false=비잔액 지출(화폐 발행)
	useBalance := true
	if fromAccount.OwnerType == model.OwnerTypeSystem && req.UseBalance != nil && !*req.UseBalance {
		useBalance = false
	}

	if err := h.transactionRepo.Create(ctx, transaction, useBalance); err != nil {
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

	// Get weekly flow for this account
	weeklyFlow, err := h.transactionRepo.GetAccountWeeklyFlow(ctx, id)
	if err != nil {
		slog.Error("Finance.ListAccountTransactions: failed to get weekly flow", "error", err, "account_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "주별 통계를 불러오는데 실패했습니다",
		})
	}

	if weeklyFlow == nil {
		weeklyFlow = []model.WeeklyFlow{}
	}

	return c.JSON(http.StatusOK, model.AccountTransactionsResponse{
		Transactions: transactions,
		Total:        len(transactions),
		Balance:      account.Balance,
		WeeklyFlow:   weeklyFlow,
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

	// Get team weekly flows
	teamWeeklyFlows, err := h.transactionRepo.GetTeamWeeklyFlows(ctx, leagueID)
	if err != nil {
		slog.Error("Finance.GetFinanceStats: failed to get team weekly flows", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀별 주별 통계를 불러오는데 실패했습니다",
		})
	}

	// Ensure non-nil slices
	if stats.TeamBalances == nil {
		stats.TeamBalances = []model.TeamBalance{}
	}
	if stats.WeeklyFlow == nil {
		stats.WeeklyFlow = []model.WeeklyFlow{}
	}
	if stats.CategoryTotals == nil {
		stats.CategoryTotals = make(map[string]int64)
	}
	if teamWeeklyFlows == nil {
		teamWeeklyFlows = []model.TeamWeeklyFlow{}
	}
	stats.TeamWeeklyFlows = teamWeeklyFlows

	return c.JSON(http.StatusOK, stats)
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
		slog.Error("Finance.CreateTransactionByDirector: failed to get league", "error", err, "league_id", leagueID)
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
		slog.Error("Finance.CreateTransactionByDirector: failed to get from account", "error", err)
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

	// Check authorization based on account type
	switch fromAccount.OwnerType {
	case model.OwnerTypeTeam:
		// For team accounts, user must be director of that team
		directorTeamIDs, err := h.participantRepo.GetDirectorTeamIDs(ctx, leagueID, userID)
		if err != nil {
			slog.Error("Finance.CreateTransactionByDirector: failed to get director team IDs", "error", err, "user_id", userID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "권한 확인에 실패했습니다",
			})
		}

		if len(directorTeamIDs) == 0 {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "감독 권한이 없습니다",
			})
		}

		// Check if user is director of this team by comparing team IDs
		isDirectorOfTeam := false
		for _, teamID := range directorTeamIDs {
			if teamID == fromAccount.OwnerID {
				isDirectorOfTeam = true
				break
			}
		}

		if !isDirectorOfTeam {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "본인 소속 팀의 계좌에서만 거래를 생성할 수 있습니다",
			})
		}

	case model.OwnerTypeParticipant:
		// For participant accounts, user must own that participant record
		participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrParticipantNotFound) {
				return c.JSON(http.StatusForbidden, model.ErrorResponse{
					Error:   "forbidden",
					Message: "해당 리그에 참여하고 있지 않습니다",
				})
			}
			slog.Error("Finance.CreateTransactionByDirector: failed to get participant", "error", err, "user_id", userID)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "권한 확인에 실패했습니다",
			})
		}

		// Verify the account belongs to this user's participant record
		if fromAccount.OwnerID != participant.ID {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "본인 계좌에서만 거래를 생성할 수 있습니다",
			})
		}

		// Verify participant is approved
		if participant.Status != model.ParticipantStatusApproved {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "승인된 참가자만 거래를 생성할 수 있습니다",
			})
		}

	case model.OwnerTypeSystem:
		// System accounts cannot be used by regular users
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "시스템 계좌에서는 거래를 생성할 수 없습니다",
		})

	default:
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "알 수 없는 계좌 유형입니다",
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
		slog.Error("Finance.CreateTransactionByDirector: failed to get to account", "error", err)
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

	transaction := &model.Transaction{
		LeagueID:      leagueID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Category:      req.Category,
		Description:   req.Description,
		CreatedBy:     &userID,
	}

	// 감독/참가자는 system 계좌 사용 불가, 항상 잔액 지출
	if err := h.transactionRepo.Create(ctx, transaction, true); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "잔액이 부족합니다",
			})
		}
		slog.Error("Finance.CreateTransactionByDirector: failed to create transaction", "error", err)
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

	// Check if user is an approved participant
	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "해당 리그에 참여하고 있지 않습니다",
			})
		}
		slog.Error("Finance.GetMyAccount: failed to get participant", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 정보를 불러오는데 실패했습니다",
		})
	}

	if participant.Status != model.ParticipantStatusApproved {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "승인된 참가자만 계좌를 조회할 수 있습니다",
		})
	}

	// Ensure participant account exists (creates if missing)
	account, err := h.accountRepo.EnsureParticipantAccount(ctx, leagueID, participant.ID)
	if err != nil {
		slog.Error("Finance.GetMyAccount: failed to ensure participant account", "error", err, "participant_id", participant.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, account)
}
