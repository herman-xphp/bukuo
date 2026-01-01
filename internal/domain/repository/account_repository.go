package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// AccountRepository defines the interface for account persistence
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error)
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Account, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.Account, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}
