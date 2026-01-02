package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

type UserUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(ur repository.UserRepository) *UserUsecase {
	return &UserUsecase{userRepo: ur}
}

type CreateUserInput struct {
	CompanyID uuid.UUID
	Email     string
	Password  string
	Name      string
	Role      entity.UserRole
}

func (uc *UserUsecase) CreateUser(ctx context.Context, input CreateUserInput) (*entity.User, error) {
	// Check if email already exists
	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("email %s is already registered", input.Email)
	}

	user, err := entity.NewUser(input.CompanyID, input.Email, input.Password, input.Name, input.Role)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	return uc.userRepo.GetByCompany(ctx, companyID)
}

func (uc *UserUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	users, err := uc.userRepo.List(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.userRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.CompanyID != companyID {
		return fmt.Errorf("user not found in this company")
	}

	if user.Role == entity.UserRoleOwner {
		return fmt.Errorf("cannot delete an owner")
	}

	return uc.userRepo.Delete(ctx, id)
}
