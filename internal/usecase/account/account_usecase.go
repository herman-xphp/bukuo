package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
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
	existing, _ := uc.accountRepo.GetByCode(ctx, input.CompanyID, input.Code)
	if existing != nil {
		return nil, fmt.Errorf("account code %s already exists", input.Code)
	}

	account := entity.NewAccount(input.CompanyID, input.Code, input.Name, input.Type)
	account.ParentID = input.ParentID
	account.IsPostable = input.IsPostable
	account.Description = input.Description

	if err := uc.accountRepo.Create(ctx, account); err != nil {
		return nil, common.WrapErr("create account", err)
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
		return nil, common.WrapErr("update account", err)
	}

	return account, nil
}

// DeleteAccount deletes an account
func (uc *AccountUsecase) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	// Check if account has balance
	balance, err := uc.journalRepo.GetBalance(ctx, id)
	if err != nil {
		return common.WrapErr("check balance", err)
	}

	if !balance.IsZero() {
		return fmt.Errorf("cannot delete account with non-zero balance: %s", balance)
	}

	if err := uc.accountRepo.Delete(ctx, id); err != nil {
		return common.WrapErr("delete account", err)
	}
	return nil
}

// AccountWithBalance represents an account with its calculated balance
type AccountWithBalance struct {
	entity.Account
	Balance decimal.Decimal `json:"balance"`
}

// GetAccountsWithBalances returns all accounts with their calculated balances
func (uc *AccountUsecase) GetAccountsWithBalances(ctx context.Context, companyID uuid.UUID) ([]AccountWithBalance, error) {
	accounts, err := uc.accountRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("get accounts", err)
	}

	journals, err := uc.journalRepo.GetByDateRange(ctx, companyID, time.Time{}, time.Now())
	if err != nil {
		return nil, common.WrapErr("get journals", err)
	}

	balances := make(map[uuid.UUID]decimal.Decimal)
	for _, journal := range journals {
		if journal.Status != entity.JournalStatusPosted {
			continue
		}
		for _, line := range journal.Lines {
			balances[line.AccountID] = balances[line.AccountID].Add(line.DebitAmount).Sub(line.CreditAmount)
		}
	}

	result := make([]AccountWithBalance, len(accounts))
	for i, acc := range accounts {
		balance := balances[acc.ID]
		if acc.NormalBalance() == "CREDIT" {
			balance = balance.Neg()
		}
		result[i] = AccountWithBalance{Account: acc, Balance: balance}
	}

	return result, nil
}

// List returns all accounts for a company with pagination and balances
func (uc *AccountUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]AccountWithBalance, int, error) {
	p := common.ValidatePagination(1, limit) // Use limit as pageSize
	if limit <= 0 {
		limit = p.PageSize
	}

	accounts, err := uc.accountRepo.List(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, common.WrapErr("list accounts", err)
	}

	total, err := uc.accountRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, common.WrapErr("count accounts", err)
	}

	result := make([]AccountWithBalance, len(accounts))
	for i, acc := range accounts {
		balance, err := uc.journalRepo.GetBalance(ctx, acc.ID)
		if err != nil {
			return nil, 0, common.WrapErr("get balance", err)
		}

		if acc.NormalBalance() == "CREDIT" {
			balance = balance.Neg()
		}
		result[i] = AccountWithBalance{Account: acc, Balance: balance}
	}

	return result, total, nil
}
