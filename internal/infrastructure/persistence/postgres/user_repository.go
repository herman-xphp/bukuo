package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.UserRepository = (*UserRepository)(nil)

// UserRepository implements repository.UserRepository for PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, company_id, email, password_hash, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.CompanyID, user.Email, user.PasswordHash,
		user.Name, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, name, role, is_active, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u entity.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
		&u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, name, role, is_active, last_login_at, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u entity.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
		&u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	query := `
		SELECT id, company_id, email, password_hash, name, role, is_active, last_login_at, created_at, updated_at
		FROM users WHERE company_id = $1 ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
			&u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users 
		SET name = $2, role = $3, is_active = $4, last_login_at = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Name, user.Role, user.IsActive, user.LastLoginAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, error) {
	whereClause := `WHERE company_id = $1`
	args := []interface{}{companyID}

	if search != "" {
		whereClause += fmt.Sprintf(` AND (name ILIKE $%d OR email ILIKE $%d)`, len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}

	query := fmt.Sprintf(`
		SELECT id, company_id, email, password_hash, name, role, is_active, last_login_at, created_at, updated_at
		FROM users %s ORDER BY name LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Email, &u.PasswordHash,
			&u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	whereClause := `WHERE company_id = $1`
	args := []interface{}{companyID}

	if search != "" {
		whereClause += fmt.Sprintf(` AND (name ILIKE $%d OR email ILIKE $%d)`, len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM users %s`, whereClause)
	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}
