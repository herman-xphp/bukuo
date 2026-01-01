package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// AuditLogRepository defines audit log operations
type AuditLogRepository interface {
	// Create saves a new audit log entry
	Create(ctx context.Context, log *entity.AuditLog) error

	// GetByCompany retrieves audit logs for a company
	GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.AuditLog, error)

	// GetByEntity retrieves audit logs for a specific entity
	GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.AuditLog, error)

	// GetByUser retrieves audit logs for a specific user
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.AuditLog, error)

	// GetByDateRange retrieves audit logs within a date range
	GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.AuditLog, error)
}
