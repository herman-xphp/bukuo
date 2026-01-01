package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *domain.Account) error {
	query := `
INSERT INTO accounts (id, company_id, code, name, type, parent_id, is_postable, description)	
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query, acc.ID, acc.CompanyID, acc.Name, acc.Type, acc.ParentID, acc.IsPostable, acc.Description)
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	query := `SELECT id, company_id, code, name, type, parent_id, is_postable, is_active FROM accounts WHERE id = $1`

	var acc domain.Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type, &acc.ParentID, &acc.IsPostable, &acc.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &acc, nil
}

func (r *AccountRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Account, error) {
	query := `SELECT id, company_id, code, name, type, parent_id, is_postable, is_active
	FROM accounts WHERE company_id = $1 ORDER BY code`

	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var acc domain.Account
		err := rows.Scan(&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type, &acc.ParentID, &acc.IsPostable, &acc.IsActive)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}
