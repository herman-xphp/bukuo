package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error)
	List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, error)
	Count(ctx context.Context, companyID uuid.UUID, search string) (int, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
