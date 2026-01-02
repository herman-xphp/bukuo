package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface implementation at compile time
var _ repository.ContactRepository = (*ContactRepository)(nil)

// ContactRepository implements repository.ContactRepository for PostgreSQL
type ContactRepository struct {
	db *pgxpool.Pool
}

// NewContactRepository creates a new ContactRepository
func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, contact *entity.Contact) error {
	query := `
		INSERT INTO contacts (
			id, company_id, code, name, contact_type, email, phone, 
			address, city, tax_id, credit_limit, payment_term_days, 
			is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.Exec(ctx, query,
		contact.ID, contact.CompanyID, contact.Code, contact.Name, contact.ContactType,
		contact.Email, contact.Phone, contact.Address, contact.City, contact.TaxID,
		contact.CreditLimit, contact.PaymentTermDays, contact.IsActive,
		contact.CreatedAt, contact.UpdatedAt,
	)
	return err
}

func (r *ContactRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Contact, error) {
	query := `
		SELECT id, company_id, code, name, contact_type, email, phone, 
			   address, city, tax_id, credit_limit, payment_term_days, 
			   is_active, created_at, updated_at
		FROM contacts 
		WHERE company_id = $1 AND id = $2
	`
	var contact entity.Contact
	err := r.db.QueryRow(ctx, query, companyID, id).Scan(
		&contact.ID, &contact.CompanyID, &contact.Code, &contact.Name, &contact.ContactType,
		&contact.Email, &contact.Phone, &contact.Address, &contact.City, &contact.TaxID,
		&contact.CreditLimit, &contact.PaymentTermDays, &contact.IsActive,
		&contact.CreatedAt, &contact.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Contact, error) {
	query := `
		SELECT id, company_id, code, name, contact_type, email, phone, 
			   address, city, tax_id, credit_limit, payment_term_days, 
			   is_active, created_at, updated_at
		FROM contacts 
		WHERE company_id = $1 AND code = $2
	`
	var contact entity.Contact
	err := r.db.QueryRow(ctx, query, companyID, code).Scan(
		&contact.ID, &contact.CompanyID, &contact.Code, &contact.Name, &contact.ContactType,
		&contact.Email, &contact.Phone, &contact.Address, &contact.City, &contact.TaxID,
		&contact.CreditLimit, &contact.PaymentTermDays, &contact.IsActive,
		&contact.CreatedAt, &contact.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.ContactFilter) ([]entity.Contact, int64, error) {
	var conditions []string
	var args []interface{}
	argPos := 1

	conditions = append(conditions, fmt.Sprintf("company_id = $%d", argPos))
	args = append(args, companyID)
	argPos++

	if filter.ContactType != nil {
		conditions = append(conditions, fmt.Sprintf("contact_type = $%d", argPos))
		args = append(args, *filter.ContactType)
		argPos++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *filter.IsActive)
		argPos++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM contacts WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, company_id, code, name, contact_type, email, phone, 
			   address, city, tax_id, credit_limit, payment_term_days, 
			   is_active, created_at, updated_at
		FROM contacts 
		WHERE %s 
		ORDER BY code
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []entity.Contact
	for rows.Next() {
		var contact entity.Contact
		err := rows.Scan(
			&contact.ID, &contact.CompanyID, &contact.Code, &contact.Name, &contact.ContactType,
			&contact.Email, &contact.Phone, &contact.Address, &contact.City, &contact.TaxID,
			&contact.CreditLimit, &contact.PaymentTermDays, &contact.IsActive,
			&contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, total, nil
}

func (r *ContactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	contact.UpdatedAt = time.Now()
	query := `
		UPDATE contacts 
		SET name = $3, contact_type = $4, email = $5, phone = $6, 
			address = $7, city = $8, tax_id = $9, credit_limit = $10, 
			payment_term_days = $11, is_active = $12, updated_at = $13
		WHERE company_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		contact.CompanyID, contact.ID, contact.Name, contact.ContactType,
		contact.Email, contact.Phone, contact.Address, contact.City, contact.TaxID,
		contact.CreditLimit, contact.PaymentTermDays, contact.IsActive, contact.UpdatedAt,
	)
	return err
}

func (r *ContactRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	// Soft delete by setting is_active = false
	query := `UPDATE contacts SET is_active = false, updated_at = $3 WHERE company_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, companyID, id, time.Now())
	return err
}

func (r *ContactRepository) ExistsByCode(ctx context.Context, companyID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM contacts WHERE company_id = $1 AND code = $2`
	args := []interface{}{companyID, code}

	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	query += `)`

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}
