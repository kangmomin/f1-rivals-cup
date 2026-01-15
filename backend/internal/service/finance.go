package service

import (
	"context"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        int64
	Category      model.TransactionCategory
	Description   *string
	UseBalance    *bool // FIA only: nil/true=use balance, false=no balance (currency issuance)
}

// FinanceStats represents finance statistics for a league
type FinanceStats struct {
	TotalCirculation int64
	TeamBalances     []model.TeamBalance
	CategoryTotals   map[string]int64
	MonthlyFlow      []model.MonthlyFlow
}

// FinanceService handles finance-related business logic
type FinanceService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	leagueRepo      *repository.LeagueRepository
	participantRepo *repository.ParticipantRepository
	teamRepo        *repository.TeamRepository
}

// NewFinanceService creates a new FinanceService instance
func NewFinanceService(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	leagueRepo *repository.LeagueRepository,
	participantRepo *repository.ParticipantRepository,
	teamRepo *repository.TeamRepository,
) *FinanceService {
	return &FinanceService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		leagueRepo:      leagueRepo,
		participantRepo: participantRepo,
		teamRepo:        teamRepo,
	}
}

// ListAccounts returns all accounts for a league
func (s *FinanceService) ListAccounts(ctx context.Context, leagueID uuid.UUID) ([]*model.Account, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	// Ensure system account exists
	if _, err := s.accountRepo.GetOrCreateSystemAccount(ctx, leagueID); err != nil {
		return nil, err
	}

	accounts, err := s.accountRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	if accounts == nil {
		accounts = []*model.Account{}
	}

	return accounts, nil
}

// GetAccount returns a single account by ID
func (s *FinanceService) GetAccount(ctx context.Context, accountID uuid.UUID) (*model.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

// SetAccountBalance updates an account's balance (admin only)
func (s *FinanceService) SetAccountBalance(ctx context.Context, accountID uuid.UUID, balance int64) (*model.Account, error) {
	// Check if account exists
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	if err := s.accountRepo.UpdateBalance(ctx, accountID, balance); err != nil {
		return nil, err
	}

	account.Balance = balance
	return account, nil
}

// validateBasicTransactionRequest validates basic transaction request fields
func (s *FinanceService) validateBasicTransactionRequest(req *CreateTransactionRequest) error {
	if req.Amount <= 0 {
		return ErrAmountMustBePositive
	}

	if req.FromAccountID == req.ToAccountID {
		return ErrSameAccount
	}

	return nil
}

// validateAccountsInLeague validates that both accounts exist and belong to the league
func (s *FinanceService) validateAccountsInLeague(ctx context.Context, leagueID uuid.UUID, fromAccountID, toAccountID uuid.UUID) (*model.Account, *model.Account, error) {
	// Verify from account exists and belongs to this league
	fromAccount, err := s.accountRepo.GetByID(ctx, fromAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, nil, ErrAccountNotFound
		}
		return nil, nil, err
	}
	if fromAccount.LeagueID != leagueID {
		return nil, nil, ErrAccountNotInLeague
	}

	// Verify to account exists and belongs to this league
	toAccount, err := s.accountRepo.GetByID(ctx, toAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, nil, ErrAccountNotFound
		}
		return nil, nil, err
	}
	if toAccount.LeagueID != leagueID {
		return nil, nil, ErrAccountNotInLeague
	}

	return fromAccount, toAccount, nil
}

// CreateTransaction creates a new transaction (admin endpoint)
func (s *FinanceService) CreateTransaction(ctx context.Context, leagueID uuid.UUID, req *CreateTransactionRequest, userID *uuid.UUID) (*model.Transaction, error) {
	// Validate basic request
	if err := s.validateBasicTransactionRequest(req); err != nil {
		return nil, err
	}

	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	// Validate accounts
	fromAccount, toAccount, err := s.validateAccountsInLeague(ctx, leagueID, req.FromAccountID, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	transaction := &model.Transaction{
		LeagueID:      leagueID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Category:      req.Category,
		Description:   req.Description,
		CreatedBy:     userID,
	}

	// FIA(system) account: apply UseBalance option
	// UseBalance: nil/true=use balance (default), false=no balance (currency issuance)
	useBalance := true
	if fromAccount.OwnerType == model.OwnerTypeSystem && req.UseBalance != nil && !*req.UseBalance {
		useBalance = false
	}

	if err := s.transactionRepo.Create(ctx, transaction, useBalance); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, err
	}

	// Set names for response
	transaction.FromName = fromAccount.OwnerName
	transaction.ToName = toAccount.OwnerName

	return transaction, nil
}

// CreateTransactionByDirector creates a transaction from user's team or participant account
func (s *FinanceService) CreateTransactionByDirector(ctx context.Context, leagueID uuid.UUID, req *CreateTransactionRequest, userID uuid.UUID) (*model.Transaction, error) {
	// Validate basic request
	if err := s.validateBasicTransactionRequest(req); err != nil {
		return nil, err
	}

	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	// Get from account
	fromAccount, err := s.accountRepo.GetByID(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	if fromAccount.LeagueID != leagueID {
		return nil, ErrAccountNotInLeague
	}

	// Check authorization based on account type
	if err := s.validateUserCanUseAccount(ctx, leagueID, userID, fromAccount); err != nil {
		return nil, err
	}

	// Verify to account exists and belongs to this league
	toAccount, err := s.accountRepo.GetByID(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	if toAccount.LeagueID != leagueID {
		return nil, ErrAccountNotInLeague
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

	// Directors/participants cannot use system account, always use balance
	if err := s.transactionRepo.Create(ctx, transaction, true); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, err
	}

	// Set names for response
	transaction.FromName = fromAccount.OwnerName
	transaction.ToName = toAccount.OwnerName

	return transaction, nil
}

// validateUserCanUseAccount checks if user has permission to use the given account
func (s *FinanceService) validateUserCanUseAccount(ctx context.Context, leagueID, userID uuid.UUID, account *model.Account) error {
	switch account.OwnerType {
	case model.OwnerTypeTeam:
		// For team accounts, user must be director of that team
		directorTeamIDs, err := s.participantRepo.GetDirectorTeamIDs(ctx, leagueID, userID)
		if err != nil {
			return err
		}

		if len(directorTeamIDs) == 0 {
			return ErrNotDirector
		}

		// Check if user is director of this team
		isDirectorOfTeam := false
		for _, teamID := range directorTeamIDs {
			if teamID == account.OwnerID {
				isDirectorOfTeam = true
				break
			}
		}

		if !isDirectorOfTeam {
			return ErrNotTeamDirector
		}

	case model.OwnerTypeParticipant:
		// For participant accounts, user must own that participant record
		participant, err := s.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrParticipantNotFound) {
				return ErrParticipantNotFound
			}
			return err
		}

		// Verify the account belongs to this user's participant record
		if account.OwnerID != participant.ID {
			return ErrNotAccountOwner
		}

		// Verify participant is approved
		if participant.Status != model.ParticipantStatusApproved {
			return ErrNotApproved
		}

	case model.OwnerTypeSystem:
		// System accounts cannot be used by regular users
		return ErrSystemAccountForbidden

	default:
		return ErrSystemAccountForbidden
	}

	return nil
}

// ListTransactions returns transactions for a league with pagination
func (s *FinanceService) ListTransactions(ctx context.Context, leagueID uuid.UUID, page, pageSize int) ([]*model.Transaction, int, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, 0, ErrLeagueNotFound
		}
		return nil, 0, err
	}

	transactions, total, err := s.transactionRepo.ListByLeague(ctx, leagueID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if transactions == nil {
		transactions = []*model.Transaction{}
	}

	return transactions, total, nil
}

// ListAccountTransactions returns transactions for a specific account
func (s *FinanceService) ListAccountTransactions(ctx context.Context, accountID uuid.UUID) ([]*model.Transaction, *model.Account, error) {
	// Get account to verify existence and get balance
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, nil, ErrAccountNotFound
		}
		return nil, nil, err
	}

	transactions, err := s.transactionRepo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, nil, err
	}

	if transactions == nil {
		transactions = []*model.Transaction{}
	}

	return transactions, account, nil
}

// GetFinanceStats returns finance statistics for a league
func (s *FinanceService) GetFinanceStats(ctx context.Context, leagueID uuid.UUID) (*FinanceStats, error) {
	// Check if league exists
	if _, err := s.leagueRepo.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return nil, ErrLeagueNotFound
		}
		return nil, err
	}

	stats, err := s.transactionRepo.GetFinanceStats(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	// Ensure non-nil slices
	result := &FinanceStats{
		TotalCirculation: stats.TotalCirculation,
		TeamBalances:     stats.TeamBalances,
		CategoryTotals:   stats.CategoryTotals,
		MonthlyFlow:      stats.MonthlyFlow,
	}

	if result.TeamBalances == nil {
		result.TeamBalances = []model.TeamBalance{}
	}
	if result.MonthlyFlow == nil {
		result.MonthlyFlow = []model.MonthlyFlow{}
	}
	if result.CategoryTotals == nil {
		result.CategoryTotals = make(map[string]int64)
	}

	return result, nil
}

// EnsureParticipantAccount gets or creates a participant's account
// This centralizes account creation logic that was previously duplicated in handlers
func (s *FinanceService) EnsureParticipantAccount(ctx context.Context, leagueID, participantID uuid.UUID) (*model.Account, error) {
	return s.accountRepo.EnsureParticipantAccount(ctx, leagueID, participantID)
}

// CreateTeamAccount creates an account for a team
// This centralizes account creation logic that was previously duplicated in handlers
func (s *FinanceService) CreateTeamAccount(ctx context.Context, leagueID, teamID uuid.UUID) (*model.Account, error) {
	// Check if account already exists
	account, err := s.accountRepo.GetByOwner(ctx, leagueID, teamID, model.OwnerTypeTeam)
	if err == nil {
		return account, nil // Account already exists
	}

	if !errors.Is(err, repository.ErrAccountNotFound) {
		return nil, err
	}

	// Create new account for team
	account = &model.Account{
		LeagueID:  leagueID,
		OwnerID:   teamID,
		OwnerType: model.OwnerTypeTeam,
		Balance:   0,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetMyAccount returns the current user's participant account
func (s *FinanceService) GetMyAccount(ctx context.Context, leagueID, userID uuid.UUID) (*model.Account, error) {
	// Check if user is an approved participant
	participant, err := s.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	if participant.Status != model.ParticipantStatusApproved {
		return nil, ErrNotApproved
	}

	// Ensure participant account exists (creates if missing)
	account, err := s.accountRepo.EnsureParticipantAccount(ctx, leagueID, participant.ID)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// EnsureSystemAccount ensures system account exists for a league and returns it
func (s *FinanceService) EnsureSystemAccount(ctx context.Context, leagueID uuid.UUID) (*model.Account, error) {
	return s.accountRepo.GetOrCreateSystemAccount(ctx, leagueID)
}
