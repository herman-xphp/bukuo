package common

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// AuditLogger provides convenient methods for creating audit logs
type AuditLogger struct {
	repo repository.AuditLogRepository
}

// NewAuditLogger creates a new AuditLogger
func NewAuditLogger(repo repository.AuditLogRepository) *AuditLogger {
	return &AuditLogger{repo: repo}
}

// LogInput contains common audit log parameters
type LogInput struct {
	CompanyID   uuid.UUID
	UserID      *uuid.UUID
	UserEmail   string
	Action      entity.AuditAction
	EntityType  string
	EntityID    *uuid.UUID
	Description string
	IPAddress   string
	UserAgent   string
}

// Log creates and saves an audit log entry
func (a *AuditLogger) Log(ctx context.Context, input LogInput) {
	if a == nil || a.repo == nil {
		return
	}
	log := entity.NewAuditLog(
		input.CompanyID, input.UserID, input.UserEmail,
		input.Action, input.EntityType, input.EntityID,
		input.Description, input.IPAddress, input.UserAgent,
	)
	_ = a.repo.Create(ctx, log)
}

// LogCreate logs a CREATE action
func (a *AuditLogger) LogCreate(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, entityType string, entityID *uuid.UUID, description string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, Action: entity.AuditActionCreate,
		EntityType: entityType, EntityID: entityID, Description: description,
	})
}

// LogUpdate logs an UPDATE action
func (a *AuditLogger) LogUpdate(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, entityType string, entityID *uuid.UUID, description string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, Action: entity.AuditActionUpdate,
		EntityType: entityType, EntityID: entityID, Description: description,
	})
}

// LogDelete logs a DELETE action
func (a *AuditLogger) LogDelete(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, entityType string, entityID *uuid.UUID, description string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, Action: entity.AuditActionDelete,
		EntityType: entityType, EntityID: entityID, Description: description,
	})
}

// LogPost logs a POST action (for journals)
func (a *AuditLogger) LogPost(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, entityType string, entityID *uuid.UUID, description string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, Action: entity.AuditActionPost,
		EntityType: entityType, EntityID: entityID, Description: description,
	})
}

// LogReverse logs a REVERSE action
func (a *AuditLogger) LogReverse(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, entityType string, entityID *uuid.UUID, description string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, Action: entity.AuditActionReverse,
		EntityType: entityType, EntityID: entityID, Description: description,
	})
}

// LogLogin logs a successful login
func (a *AuditLogger) LogLogin(ctx context.Context, companyID uuid.UUID, userID *uuid.UUID, email, ip, ua string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserID: userID, UserEmail: email,
		Action: entity.AuditActionLogin, EntityType: "USER", EntityID: userID,
		Description: "User logged in", IPAddress: ip, UserAgent: ua,
	})
}

// LogFailedLogin logs a failed login attempt
func (a *AuditLogger) LogFailedLogin(ctx context.Context, companyID uuid.UUID, email, reason, ip, ua string) {
	a.Log(ctx, LogInput{
		CompanyID: companyID, UserEmail: email,
		Action: entity.AuditActionFailed, EntityType: "AUTH",
		Description: "Failed login: " + reason, IPAddress: ip, UserAgent: ua,
	})
}
