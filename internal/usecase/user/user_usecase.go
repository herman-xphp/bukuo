package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
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
	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("email %s is already registered", input.Email)
	}

	user, err := entity.NewUser(input.CompanyID, input.Email, input.Password, input.Name, input.Role)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, common.WrapErr("save user", err)
	}

	return user, nil
}

func (uc *UserUsecase) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	users, err := uc.userRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, common.WrapErr("list users", err)
	}
	return users, nil
}

func (uc *UserUsecase) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, int, error) {
	p := common.ValidatePagination(1, limit)
	if limit <= 0 {
		limit = p.PageSize
	}

	users, err := uc.userRepo.List(ctx, companyID, limit, offset, search)
	if err != nil {
		return nil, 0, common.WrapErr("list users", err)
	}

	total, err := uc.userRepo.Count(ctx, companyID, search)
	if err != nil {
		return nil, 0, common.WrapErr("count users", err)
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

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return common.WrapErr("delete user", err)
	}
	return nil
}

type UpdateUserInput struct {
	Name           string
	Role           entity.UserRole
	IsActive       bool
	ProfilePicture *string
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id uuid.UUID, companyID uuid.UUID, input UpdateUserInput) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.CompanyID != companyID {
		return nil, fmt.Errorf("user not found in this company")
	}

	if user.Role == entity.UserRoleOwner && input.Role != entity.UserRoleOwner {
		return nil, fmt.Errorf("cannot change owner's role")
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	user.IsActive = input.IsActive
	if input.ProfilePicture != nil {
		user.ProfilePicture = input.ProfilePicture
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, common.WrapErr("update user", err)
	}

	return user, nil
}

func (uc *UserUsecase) ResetPassword(ctx context.Context, id uuid.UUID, companyID uuid.UUID, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.CompanyID != companyID {
		return fmt.Errorf("user not found in this company")
	}

	if err := entity.ValidatePassword(newPassword); err != nil {
		return err
	}

	if err := user.SetPassword(newPassword); err != nil {
		return common.WrapErr("hash password", err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return common.WrapErr("update password", err)
	}

	return nil
}

func (uc *UserUsecase) SetUserPin(ctx context.Context, id uuid.UUID, companyID uuid.UUID, pin string) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.CompanyID != companyID {
		return fmt.Errorf("user not found in this company")
	}

	if err := user.SetPin(pin); err != nil {
		return err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return common.WrapErr("update user pin", err)
	}

	return nil
}

func (uc *UserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
