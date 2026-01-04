package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// AuditUsecase handles audit trail business logic
type AuditUsecase struct {
	repo repository.AuditLogRepository
}

// NewAuditUsecase creates a new AuditUsecase
func NewAuditUsecase(r repository.AuditLogRepository) *AuditUsecase {
	return &AuditUsecase{repo: r}
}

// GetCompanyLogs retrieves audit logs for a company with pagination
func (uc *AuditUsecase) GetCompanyLogs(ctx context.Context, companyID uuid.UUID, page, pageSize int) ([]entity.AuditLog, error) {
	offset := (page - 1) * pageSize
	return uc.repo.GetByCompany(ctx, companyID, pageSize, offset)
}

// GetEntityLogs retrieves audit logs for a specific entity
func (uc *AuditUsecase) GetEntityLogs(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.AuditLog, error) {
	return uc.repo.GetByEntity(ctx, entityType, entityID)
}

// GetUserLogs retrieves audit logs for a specific user
func (uc *AuditUsecase) GetUserLogs(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]entity.AuditLog, error) {
	offset := (page - 1) * pageSize
	return uc.repo.GetByUser(ctx, userID, pageSize, offset)
}

// GetLogsByDateRange retrieves audit logs within a date range
func (uc *AuditUsecase) GetLogsByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.AuditLog, error) {
	return uc.repo.GetByDateRange(ctx, companyID, start, end)
}

// GetAuditSummary provides a summary of recent audit activity
func (uc *AuditUsecase) GetAuditSummary(ctx context.Context, companyID uuid.UUID) (map[string]interface{}, error) {
	// Get last 24 hours of logs
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	logs, err := uc.repo.GetByDateRange(ctx, companyID, dayAgo, now)
	if err != nil {
		return nil, err
	}

	// Aggregate by action
	actionCounts := make(map[string]int)
	entityCounts := make(map[string]int)
	userActivity := make(map[string]int)

	for _, log := range logs {
		actionCounts[string(log.Action)]++
		entityCounts[log.EntityType]++
		userActivity[log.UserEmail]++
	}

	return map[string]interface{}{
		"total_events_24h": len(logs),
		"by_action":        actionCounts,
		"by_entity":        entityCounts,
		"by_user":          userActivity,
		"period_start":     dayAgo,
		"period_end":       now,
	}, nil
}
