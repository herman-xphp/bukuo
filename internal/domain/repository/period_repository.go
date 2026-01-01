package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// PeriodRepository defines the interface for accounting period persistence
type PeriodRepository interface {
	Create(ctx context.Context, period *entity.AccountingPeriod) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AccountingPeriod, error)
	GetByDate(ctx context.Context, companyID uuid.UUID, date time.Time) (*entity.AccountingPeriod, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error)
	GetOpenPeriods(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error)
	Update(ctx context.Context, period *entity.AccountingPeriod) error
}
