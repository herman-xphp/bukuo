package banking

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

// BankAccountRepository interface
type BankAccountRepository interface {
	Create(ctx context.Context, acc *entity.BankAccount) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.BankAccount, error)
	List(ctx context.Context, companyID uuid.UUID) ([]entity.BankAccount, error)
	Update(ctx context.Context, acc *entity.BankAccount) error
}

// BankTransactionRepository interface
type BankTransactionRepository interface {
	Create(ctx context.Context, txn *entity.BankTransaction) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.BankTransaction, error)
	ListByAccount(ctx context.Context, companyID, accountID uuid.UUID, limit, offset int) ([]entity.BankTransaction, int64, error)
	MarkReconciled(ctx context.Context, companyID, id uuid.UUID) error
}

// BankingUsecase handles banking business logic
type BankingUsecase struct {
	accountRepo BankAccountRepository
	txnRepo     BankTransactionRepository
	journalRepo repository.JournalRepository
	txManager   repository.TransactionManager
}

// NewBankingUsecase creates a new BankingUsecase
func NewBankingUsecase(
	ar BankAccountRepository,
	tr BankTransactionRepository,
	jr repository.JournalRepository,
	tm repository.TransactionManager,
) *BankingUsecase {
	return &BankingUsecase{
		accountRepo: ar,
		txnRepo:     tr,
		journalRepo: jr,
		txManager:   tm,
	}
}

// CreateBankAccountInput holds input for bank account creation
type CreateBankAccountInput struct {
	AccountID      uuid.UUID // GL Account
	BankName       string
	AccountNumber  string
	AccountName    string
	CurrencyID     uuid.UUID
	OpeningBalance decimal.Decimal
}

// CreateBankAccount creates a new bank account
func (uc *BankingUsecase) CreateBankAccount(ctx context.Context, companyID uuid.UUID, input CreateBankAccountInput) (*entity.BankAccount, error) {
	acc := &entity.BankAccount{
		ID:             uuid.New(),
		CompanyID:      companyID,
		AccountID:      input.AccountID,
		BankName:       input.BankName,
		AccountNumber:  input.AccountNumber,
		AccountName:    input.AccountName,
		CurrencyID:     input.CurrencyID,
		CurrentBalance: input.OpeningBalance,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := uc.accountRepo.Create(ctx, acc); err != nil {
		return nil, common.WrapErr("create bank account", err)
	}

	return acc, nil
}

// ListBankAccounts lists all bank accounts
func (uc *BankingUsecase) ListBankAccounts(ctx context.Context, companyID uuid.UUID) ([]entity.BankAccount, error) {
	return uc.accountRepo.List(ctx, companyID)
}

// GetBankAccount retrieves a bank account by ID
func (uc *BankingUsecase) GetBankAccount(ctx context.Context, companyID, id uuid.UUID) (*entity.BankAccount, error) {
	return uc.accountRepo.GetByID(ctx, companyID, id)
}

// CreateTransactionInput holds input for bank transaction creation
type CreateTransactionInput struct {
	BankAccountID   uuid.UUID
	TransactionDate time.Time
	Type            entity.BankTransactionType
	Amount          decimal.Decimal
	Description     string
	Reference       string
}

// CreateTransaction creates a bank transaction
func (uc *BankingUsecase) CreateTransaction(ctx context.Context, companyID uuid.UUID, input CreateTransactionInput) (*entity.BankTransaction, error) {
	if input.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, common.NewValidationError("amount must be positive")
	}

	txn := &entity.BankTransaction{
		ID:              uuid.New(),
		CompanyID:       companyID,
		BankAccountID:   input.BankAccountID,
		TransactionNo:   fmt.Sprintf("BTX-%s", time.Now().Format("20060102150405")),
		TransactionDate: input.TransactionDate,
		Type:            input.Type,
		Amount:          input.Amount,
		Description:     input.Description,
		Reference:       input.Reference,
		Reconciled:      false,
		CreatedAt:       time.Now(),
	}

	if err := uc.txnRepo.Create(ctx, txn); err != nil {
		return nil, common.WrapErr("create transaction", err)
	}

	return txn, nil
}

// ListTransactions lists transactions for a bank account
func (uc *BankingUsecase) ListTransactions(ctx context.Context, companyID, accountID uuid.UUID, page, pageSize int) ([]entity.BankTransaction, int64, error) {
	offset := (page - 1) * pageSize
	return uc.txnRepo.ListByAccount(ctx, companyID, accountID, pageSize, offset)
}

// ReconcileTransaction marks a transaction as reconciled
func (uc *BankingUsecase) ReconcileTransaction(ctx context.Context, companyID, txnID uuid.UUID) error {
	return uc.txnRepo.MarkReconciled(ctx, companyID, txnID)
}
