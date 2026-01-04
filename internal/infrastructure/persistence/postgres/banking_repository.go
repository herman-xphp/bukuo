package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/txhelper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BankAccountRepository struct {
	db *pgxpool.Pool
}

func NewBankAccountRepository(db *pgxpool.Pool) *BankAccountRepository {
	return &BankAccountRepository{db: db}
}

func (r *BankAccountRepository) Create(ctx context.Context, acc *entity.BankAccount) error {
	query := `
		INSERT INTO bank_accounts (
			id, company_id, account_id, bank_name, account_number, account_name,
			currency_id, current_balance, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		acc.ID, acc.CompanyID, acc.AccountID, acc.BankName, acc.AccountNumber, acc.AccountName,
		acc.CurrencyID, acc.CurrentBalance, acc.IsActive, acc.CreatedAt, acc.UpdatedAt,
	)
	return err
}

func (r *BankAccountRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.BankAccount, error) {
	query := `
		SELECT id, company_id, account_id, bank_name, account_number, account_name,
			   currency_id, current_balance, is_active, created_at, updated_at
		FROM bank_accounts
		WHERE company_id = $1 AND id = $2
	`
	var acc entity.BankAccount
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&acc.ID, &acc.CompanyID, &acc.AccountID, &acc.BankName, &acc.AccountNumber, &acc.AccountName,
		&acc.CurrencyID, &acc.CurrentBalance, &acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *BankAccountRepository) List(ctx context.Context, companyID uuid.UUID) ([]entity.BankAccount, error) {
	query := `
		SELECT id, company_id, account_id, bank_name, account_number, account_name,
			   currency_id, current_balance, is_active, created_at, updated_at
		FROM bank_accounts
		WHERE company_id = $1
		ORDER BY bank_name, account_name
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []entity.BankAccount
	for rows.Next() {
		var acc entity.BankAccount
		if err := rows.Scan(
			&acc.ID, &acc.CompanyID, &acc.AccountID, &acc.BankName, &acc.AccountNumber, &acc.AccountName,
			&acc.CurrencyID, &acc.CurrentBalance, &acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *BankAccountRepository) Update(ctx context.Context, acc *entity.BankAccount) error {
	acc.UpdatedAt = time.Now()
	query := `
		UPDATE bank_accounts
		SET bank_name = $3, account_number = $4, account_name = $5, is_active = $6, updated_at = $7
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query, acc.CompanyID, acc.ID, acc.BankName, acc.AccountNumber, acc.AccountName, acc.IsActive, acc.UpdatedAt)
	return err
}

func (r *BankAccountRepository) UpdateBalance(ctx context.Context, companyID, id uuid.UUID, balance entity.BankAccount) error {
	query := "UPDATE bank_accounts SET current_balance = $3, updated_at = $4 WHERE company_id = $1 AND id = $2"
	_, err := r.db.Exec(ctx, query, companyID, id, balance.CurrentBalance, time.Now())
	return err
}

// BankTransactionRepository
type BankTransactionRepository struct {
	db *pgxpool.Pool
}

func NewBankTransactionRepository(db *pgxpool.Pool) *BankTransactionRepository {
	return &BankTransactionRepository{db: db}
}

func (r *BankTransactionRepository) Create(ctx context.Context, txn *entity.BankTransaction) error {
	return txhelper.RunInTx(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO bank_transactions (
				id, company_id, bank_account_id, transaction_no, transaction_date,
				type, amount, description, reference, journal_id, reconciled, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		if _, err := tx.Exec(ctx, query,
			txn.ID, txn.CompanyID, txn.BankAccountID, txn.TransactionNo, txn.TransactionDate,
			txn.Type, txn.Amount, txn.Description, txn.Reference, txn.JournalID, txn.Reconciled, txn.CreatedAt,
		); err != nil {
			return err
		}

		// Update bank account balance
		var balanceUpdate string
		if txn.Type == entity.BankTransactionDeposit || txn.Type == entity.BankTransactionInterest {
			balanceUpdate = "UPDATE bank_accounts SET current_balance = current_balance + $2, updated_at = $3 WHERE id = $1"
		} else {
			balanceUpdate = "UPDATE bank_accounts SET current_balance = current_balance - $2, updated_at = $3 WHERE id = $1"
		}
		_, err := tx.Exec(ctx, balanceUpdate, txn.BankAccountID, txn.Amount, time.Now())
		return err
	})
}

func (r *BankTransactionRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.BankTransaction, error) {
	query := `
		SELECT id, company_id, bank_account_id, transaction_no, transaction_date,
			   type, amount, description, reference, journal_id, reconciled, reconciled_at, created_at
		FROM bank_transactions
		WHERE company_id = $1 AND id = $2
	`
	var txn entity.BankTransaction
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&txn.ID, &txn.CompanyID, &txn.BankAccountID, &txn.TransactionNo, &txn.TransactionDate,
		&txn.Type, &txn.Amount, &txn.Description, &txn.Reference, &txn.JournalID, &txn.Reconciled, &txn.ReconciledAt, &txn.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *BankTransactionRepository) ListByAccount(ctx context.Context, companyID, accountID uuid.UUID, limit, offset int) ([]entity.BankTransaction, int64, error) {
	countQuery := "SELECT COUNT(*) FROM bank_transactions WHERE company_id = $1 AND bank_account_id = $2"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, companyID, accountID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, company_id, bank_account_id, transaction_no, transaction_date,
			   type, amount, description, reference, journal_id, reconciled, reconciled_at, created_at
		FROM bank_transactions
		WHERE company_id = $1 AND bank_account_id = $2
		ORDER BY transaction_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, companyID, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []entity.BankTransaction
	for rows.Next() {
		var txn entity.BankTransaction
		if err := rows.Scan(
			&txn.ID, &txn.CompanyID, &txn.BankAccountID, &txn.TransactionNo, &txn.TransactionDate,
			&txn.Type, &txn.Amount, &txn.Description, &txn.Reference, &txn.JournalID, &txn.Reconciled, &txn.ReconciledAt, &txn.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, txn)
	}
	return transactions, total, nil
}

func (r *BankTransactionRepository) MarkReconciled(ctx context.Context, companyID, id uuid.UUID) error {
	now := time.Now()
	query := "UPDATE bank_transactions SET reconciled = true, reconciled_at = $3 WHERE company_id = $1 AND id = $2"
	_, err := r.db.Exec(ctx, query, companyID, id, now)
	return err
}
