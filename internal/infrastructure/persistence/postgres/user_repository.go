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

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, company_id, email, password_hash, pin_hash, name, role, is_active, profile_picture, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := GetExecutor(ctx, r.db).Exec(ctx, query,
		user.ID, user.CompanyID, user.Email, user.PasswordHash, user.PinHash,
		user.Name, user.Role, user.IsActive, user.ProfilePicture, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, COALESCE(pin_hash, ''), name, role, is_active, profile_picture, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u entity.User
	err := GetExecutor(ctx, r.db).QueryRow(ctx, query, id).Scan(
		&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash, &u.PinHash,
		&u.Name, &u.Role, &u.IsActive, &u.ProfilePicture, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, COALESCE(pin_hash, ''), name, role, is_active, profile_picture, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u entity.User
	err := GetExecutor(ctx, r.db).QueryRow(ctx, query, email).Scan(
		&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash, &u.PinHash,
		&u.Name, &u.Role, &u.IsActive, &u.ProfilePicture, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, name, role, is_active, profile_picture, last_login_at, created_at, updated_at
		FROM users WHERE company_id = $1 ORDER BY name
	`
	rows, err := GetExecutor(ctx, r.db).Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
			&u.Name, &u.Role, &u.IsActive, &u.ProfilePicture, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users 
		SET name = $2, role = $3, is_active = $4, profile_picture = $5, pin_hash = $6, last_login_at = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := GetExecutor(ctx, r.db).Exec(ctx, query,
		user.ID, user.Name, user.Role, user.IsActive, user.ProfilePicture, user.PinHash, user.LastLoginAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := GetExecutor(ctx, r.db).Exec(ctx, query, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "email")
	}

	whereClause := qb.WhereClause()
	limitPos, offsetPos := qb.AddLimitOffset(limit, offset)

	query := fmt.Sprintf(`
		SELECT id, company_id, email, password_hash, name, role, is_active, profile_picture, last_login_at, created_at, updated_at
		FROM users WHERE %s ORDER BY name LIMIT $%d OFFSET $%d
	`, whereClause, limitPos, offsetPos)

	rows, err := GetExecutor(ctx, r.db).Query(ctx, query, qb.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
			&u.Name, &u.Role, &u.IsActive, &u.ProfilePicture, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	qb := querybuilder.New()
	qb.AddCondition("company_id = $%d", companyID)

	if search != "" {
		qb.AddSearch(search, "name", "email")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE %s`, qb.WhereClause())
	var count int
	err := GetExecutor(ctx, r.db).QueryRow(ctx, query, qb.Args()...).Scan(&count)
	return count, err
}
