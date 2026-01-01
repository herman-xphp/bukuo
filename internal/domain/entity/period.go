package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Errors
var (
	ErrPeriodNotOpen = errors.New("period is not open")
	ErrPeriodLocked  = errors.New("period is locked")
)

// PeriodStatus represents the status of an accounting period
type PeriodStatus string

const (
	PeriodStatusOpen   PeriodStatus = "OPEN"
	PeriodStatusClosed PeriodStatus = "CLOSED"
	PeriodStatusLocked PeriodStatus = "LOCKED"
)

// AccountingPeriod represents a fiscal accounting period
type AccountingPeriod struct {
	ID        uuid.UUID    `json:"id"`
	CompanyID uuid.UUID    `json:"company_id"`
	Name      string       `json:"name"`
	StartDate time.Time    `json:"start_date"`
	EndDate   time.Time    `json:"end_date"`
	Status    PeriodStatus `json:"status"`
	ClosedAt  *time.Time   `json:"closed_at"`
	ClosedBy  *uuid.UUID   `json:"closed_by"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewAccountingPeriod creates a new accounting period
func NewAccountingPeriod(companyID uuid.UUID, name string, start, end time.Time) *AccountingPeriod {
	return &AccountingPeriod{
		ID:        uuid.New(),
		CompanyID: companyID,
		Name:      name,
		StartDate: start,
		EndDate:   end,
		Status:    PeriodStatusOpen,
		CreatedAt: time.Now(),
	}
}

// IsOpen checks if the period is open
func (p *AccountingPeriod) IsOpen() bool {
	return p.Status == PeriodStatusOpen
}

// CanPost checks if journals can be posted to this period
func (p *AccountingPeriod) CanPost() error {
	if p.Status == PeriodStatusLocked {
		return ErrPeriodLocked
	}
	if p.Status != PeriodStatusOpen {
		return ErrPeriodNotOpen
	}
	return nil
}

// Close closes the period
func (p *AccountingPeriod) Close(userID uuid.UUID) error {
	if p.Status != PeriodStatusOpen {
		return ErrPeriodNotOpen
	}
	now := time.Now()
	p.Status = PeriodStatusClosed
	p.ClosedAt = &now
	p.ClosedBy = &userID
	return nil
}

// Lock locks the period (cannot be reopened)
func (p *AccountingPeriod) Lock() error {
	if p.Status != PeriodStatusClosed {
		return errors.New("period must be closed before locking")
	}
	p.Status = PeriodStatusLocked
	return nil
}
