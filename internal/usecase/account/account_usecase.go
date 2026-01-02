package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/shopspring/decimal"
)

// AccountUsecase handles account business logic
type AccountUsecase struct {
	accountRepo repository.AccountRepository
	journalRepo repository.JournalRepository
}

// NewAccountUsecase creates a new AccountUsecase
func NewAccountUsecase(ar repository.AccountRepository, jr repository.JournalRepository) *AccountUsecase {
	return &AccountUsecase{accountRepo: ar, journalRepo: jr}
}

// CreateAccountInput represents input for creating an account
type CreateAccountInput struct {
	CompanyID   uuid.UUID
	Code        string
	Name        string
	Type        entity.AccountType
	ParentID    *uuid.UUID
	IsPostable  bool
	Description string
}

// CreateAccount creates a new account
func (uc *AccountUsecase) CreateAccount(ctx context.Context, input CreateAccountInput) (*entity.Account, error) {
	// Check if code exists
	existing, _ := uc.accountRepo.GetByCode(ctx, input.CompanyID, input.Code)
	if existing != nil {
		return nil, fmt.Errorf("account code %s already exists", input.Code)
	}

	account := entity.NewAccount(input.CompanyID, input.Code, input.Name, input.Type)
	account.ParentID = input.ParentID
	account.IsPostable = input.IsPostable
	account.Description = input.Description

	if err := uc.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// GetByID retrieves an account by ID
func (uc *AccountUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	return uc.accountRepo.GetByID(ctx, id)
}

// GetByCompany retrieves all accounts for a company
func (uc *AccountUsecase) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.Account, error) {
	return uc.accountRepo.GetByCompany(ctx, companyID)
}

// UpdateAccountInput represents input for updating an account
type UpdateAccountInput struct {
	ID          uuid.UUID
	Name        string
	Type        entity.AccountType
	ParentID    *uuid.UUID
	IsPostable  bool
	IsActive    bool
	Description string
}

// UpdateAccount updates an account
func (uc *AccountUsecase) UpdateAccount(ctx context.Context, input UpdateAccountInput) (*entity.Account, error) {
	account, err := uc.accountRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	account.Name = input.Name
	account.Type = input.Type
	account.ParentID = input.ParentID
	account.IsPostable = input.IsPostable
	account.IsActive = input.IsActive
	account.Description = input.Description

	if err := uc.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// DeleteAccount deletes an account
func (uc *AccountUsecase) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return uc.accountRepo.Delete(ctx, id)
}

// AccountWithBalance represents an account with its calculated balance
type AccountWithBalance struct {
	entity.Account
	Balance decimal.Decimal `json:"balance"`
}

// GetAccountsWithBalances returns all accounts with their calculated balances
func (uc *AccountUsecase) GetAccountsWithBalances(ctx context.Context, companyID uuid.UUID) ([]AccountWithBalance, error) {
	// Get all accounts
	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Get all posted journals from the beginning of time to now
	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, time.Time{}, time.Now())
	if err != nil {
		return nil, err
	}

	// Calculate balances from posted journals
	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

	// Create account list with balances
	result := make([]AccountWithBalance, len(accounts))
	for i, acc := range accounts {
		balance := balances[acc.ID]
		// Adjust for normal balance (Credit accounts like LIABILITY, EQUITY, REVENUE should show positive when negative)
		if acc.NormalBalance() == "CREDIT" {
			balance = balance.Neg()
		}
		result[i] = AccountWithBalance{
			Account: acc,
			Balance: balance,
		}
	}

	return result, nil
}

// List returns all accounts for a company with pagination and balances
func (uc *AccountUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]AccountWithBalance, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	accounts, err := uc.accountRepo.List(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.accountRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, err
	}

	result := make([]AccountWithBalance, len(accounts))
	for i, acc := range accounts {
		balance, err := uc.journalRepo.GetBalance(ctx, acc.ID)
		if err != nil {
			return nil, 0, err
		}

		// Adjust for normal balance
		if acc.NormalBalance() == "CREDIT" {
			balance = balance.Neg()
		}

		result[i] = AccountWithBalance{
			Account: acc,
			Balance: balance,
		}
	}

	return result, total, nil
}
