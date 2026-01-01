package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action
type AuditAction string

const (
	AuditActionCreate  AuditAction = "CREATE"
	AuditActionUpdate  AuditAction = "UPDATE"
	AuditActionDelete  AuditAction = "DELETE"
	AuditActionPost    AuditAction = "POST"
	AuditActionReverse AuditAction = "REVERSE"
	AuditActionLogin   AuditAction = "LOGIN"
	AuditActionLogout  AuditAction = "LOGOUT"
	AuditActionFailed  AuditAction = "FAILED_LOGIN"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          uuid.UUID   `json:"id"`
	CompanyID   uuid.UUID   `json:"company_id"`
	UserID      *uuid.UUID  `json:"user_id"` // nil for failed logins
	UserEmail   string      `json:"user_email"`
	Action      AuditAction `json:"action"`
	EntityType  string      `json:"entity_type"` // "JOURNAL", "ACCOUNT", "USER", etc.
	EntityID    *uuid.UUID  `json:"entity_id"`
	Description string      `json:"description"`
	OldValue    string      `json:"old_value,omitempty"` // JSON of old state
	NewValue    string      `json:"new_value,omitempty"` // JSON of new state
	IPAddress   string      `json:"ip_address"`
	UserAgent   string      `json:"user_agent"`
	CreatedAt   time.Time   `json:"created_at"`
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	companyID uuid.UUID,
	userID *uuid.UUID,
	userEmail string,
	action AuditAction,
	entityType string,
	entityID *uuid.UUID,
	description string,
	ipAddress string,
	userAgent string,
) *AuditLog {
	return &AuditLog{
		ID:          uuid.New(),
		CompanyID:   companyID,
		UserID:      userID,
		UserEmail:   userEmail,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}
}

// WithValues sets the old and new values for the audit log
func (a *AuditLog) WithValues(oldValue, newValue string) *AuditLog {
	a.OldValue = oldValue
	a.NewValue = newValue
	return a
}
