package domain

import (
	"time"

	"github.com/google/uuid"
)

type PeriodStatus string

const (
	PeriodStatusOpen   PeriodStatus = "OPEN"
	PeriodStatusClosed PeriodStatus = "CLOSED"
	PeriodStatusLocked PeriodStatus = "LOCKED"
)

type AccountPeriod struct {
	ID        uuid.UUID    `json:"id"`
	CompanyID uuid.UUID    `json:"company_id"`
	Name      string       `json:"name"`
	StartDate time.Time    `json:"start_date"`
	EndDate   time.Time    `json:"end_date"`
	Status    PeriodStatus `json:"status"`
	ClosedAt  *time.Time   `json:"closed_at"`
	CreatedAt time.Time    `json:"created_at"`
}

func (p *AccountPeriod) IsOpen() bool {
	return p.Status == PeriodStatusOpen
}
