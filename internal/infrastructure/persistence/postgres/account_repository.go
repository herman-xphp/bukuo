package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres/querybuilder"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.AccountRepository = (*AccountRepository)(nil)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *entity.Account) error {
	query := `
		INSERT INTO accounts (id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		acc.ID, acc.CompanyID, acc.Code, acc.Name, acc.Type,
		acc.ParentID, acc.IsPostable, acc.IsActive, acc.Description,
		acc.CreatedAt, acc.UpdatedAt,
	)
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	query := `
		SELECT id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at
		FROM accounts WHERE id = $1
	`
	var acc entity.Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type,
		&acc.ParentID, &acc.IsPostable, &acc.IsActive, &acc.Description,
		&acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Account, error) {
	query := `
		SELECT id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at
		FROM accounts WHERE company_id = $1 AND code = $2
	`
	var acc entity.Account
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(
		&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type,
		&acc.ParentID, &acc.IsPostable, &acc.IsActive, &acc.Description,
		&acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.Account, error) {
	query := `
		SELECT id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at
		FROM accounts WHERE company_id = $1 ORDER BY code
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []entity.Account
	for rows.Next() {
		var acc entity.Account
		if err := rows.Scan(
			&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type,
			&acc.ParentID, &acc.IsPostable, &acc.IsActive, &acc.Description,
			&acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *AccountRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*entity.Account, error) {
	query := `
		SELECT id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at
		FROM accounts WHERE id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make(map[uuid.UUID]*entity.Account)
	for rows.Next() {
		var acc entity.Account
		if err := rows.Scan(
			&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type,
			&acc.ParentID, &acc.IsPostable, &acc.IsActive, &acc.Description,
			&acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts[acc.ID] = &acc
	}
	return accounts, nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *entity.Account) error {
	query := `
		UPDATE accounts 
		SET name = $2, type = $3, parent_id = $4, is_postable = $5, is_active = $6, description = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		acc.ID, acc.Name, acc.Type, acc.ParentID,
		acc.IsPostable, acc.IsActive, acc.Description, acc.UpdatedAt,
	)
	return err
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *AccountRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.Account, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "code")
	}

	whereClause := qb.WhereClause()
	limitPos, offsetPos := qb.AddLimitOffset(limit, offset)

	query := fmt.Sprintf(`
		SELECT id, company_id, code, name, type, parent_id, is_postable, is_active, description, created_at, updated_at
		FROM accounts WHERE %s ORDER BY code LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := r.db.Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []entity.Account
	for rows.Next() {
		var acc entity.Account
		if err := rows.Scan(
			&acc.ID, &acc.CompanyID, &acc.Code, &acc.Name, &acc.Type,
			&acc.ParentID, &acc.IsPostable, &acc.IsActive, &acc.Description,
			&acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *AccountRepository) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "code")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM accounts WHERE %s`, qb.WhereClause())
	var count int
	err := r.db.QueryRow(ctx, query, qb.Args()...).Scan(&count)
	return count, err
}
