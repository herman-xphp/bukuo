package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// AccountUsecase handles account business logic
type AccountUsecase struct {
	accountRepo repository.AccountRepository
}

// NewAccountUsecase creates a new AccountUsecase
func NewAccountUsecase(ar repository.AccountRepository) *AccountUsecase {
	return &AccountUsecase{accountRepo: ar}
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
