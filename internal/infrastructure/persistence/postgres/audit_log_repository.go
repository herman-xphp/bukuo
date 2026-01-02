package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.AuditLogRepository = (*AuditLogRepository)(nil)

type AuditLogRepository struct {
	db *pgxpool.Pool
}

func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_logs 
		(id, company_id, user_id, user_email, action, entity_type, entity_id, 
		 description, old_value, new_value, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := GetExecutor(ctx, r.db).Exec(ctx, query,
		log.ID, log.CompanyID, log.UserID, log.UserEmail, log.Action,
		log.EntityType, log.EntityID, log.Description,
		nullableString(log.OldValue), nullableString(log.NewValue),
		log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	return err
}

func (r *AuditLogRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	query := `
		SELECT id, company_id, user_id, user_email, action, entity_type, entity_id,
		       description, old_value, new_value, ip_address, user_agent, created_at
		FROM audit_logs 
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryLogs(ctx, query, companyID, limit, offset)
}

func (r *AuditLogRepository) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.AuditLog, error) {
	query := `
		SELECT id, company_id, user_id, user_email, action, entity_type, entity_id,
		       description, old_value, new_value, ip_address, user_agent, created_at
		FROM audit_logs 
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`
	return r.queryLogs(ctx, query, entityType, entityID)
}

func (r *AuditLogRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	query := `
		SELECT id, company_id, user_id, user_email, action, entity_type, entity_id,
		       description, old_value, new_value, ip_address, user_agent, created_at
		FROM audit_logs 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryLogs(ctx, query, userID, limit, offset)
}

func (r *AuditLogRepository) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.AuditLog, error) {
	query := `
		SELECT id, company_id, user_id, user_email, action, entity_type, entity_id,
		       description, old_value, new_value, ip_address, user_agent, created_at
		FROM audit_logs 
		WHERE company_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
	`
	return r.queryLogs(ctx, query, companyID, start, end)
}

func (r *AuditLogRepository) queryLogs(ctx context.Context, query string, args ...interface{}) ([]entity.AuditLog, error) {
	rows, err := GetExecutor(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.AuditLog
	for rows.Next() {
		var log entity.AuditLog
		var oldValue, newValue *string
		err := rows.Scan(
			&log.ID, &log.CompanyID, &log.UserID, &log.UserEmail, &log.Action,
			&log.EntityType, &log.EntityID, &log.Description,
			&oldValue, &newValue, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if oldValue != nil {
			log.OldValue = *oldValue
		}
		if newValue != nil {
			log.NewValue = *newValue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
