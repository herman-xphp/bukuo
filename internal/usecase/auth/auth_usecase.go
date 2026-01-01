package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

var (
	ErrEmailExists = errors.New("email already registered")
)

// AuthUsecase handles authentication business logic
type AuthUsecase struct {
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	jwtService  *JWTService
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(
	ur repository.UserRepository,
	cr repository.CompanyRepository,
	jwt *JWTService,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    ur,
		companyRepo: cr,
		jwtService:  jwt,
	}
}

// RegisterInput represents registration input
type RegisterInput struct {
	CompanyName string
	Email       string
	Password    string
	Name        string
}

// RegisterOutput represents registration output
type RegisterOutput struct {
	User    *entity.User    `json:"user"`
	Company *entity.Company `json:"company"`
	Token   string          `json:"token"`
}

// Register creates a new company and owner user
func (uc *AuthUsecase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// Check if email exists
	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrEmailExists
	}

	// Create company
	company := entity.NewCompany(input.CompanyName, "")
	if err := uc.companyRepo.Create(ctx, company); err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Create owner user
	user, err := entity.NewUser(company.ID, input.Email, input.Password, input.Name, entity.UserRoleOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Generate token
	token, err := uc.jwtService.GenerateToken(user.ID, company.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterOutput{
		User:    user,
		Company: company,
		Token:   token,
	}, nil
}

// LoginInput represents login input
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput represents login output
type LoginOutput struct {
	User  *entity.User `json:"user"`
	Token string       `json:"token"`
}

// Login authenticates a user
func (uc *AuthUsecase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, entity.ErrInvalidCredentials
	}

	if err := user.CheckPassword(input.Password); err != nil {
		return nil, err
	}

	// Update last login
	user.UpdateLastLogin()
	_ = uc.userRepo.Update(ctx, user)

	// Generate token
	token, err := uc.jwtService.GenerateToken(user.ID, user.CompanyID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		User:  user,
		Token: token,
	}, nil
}

// GetUserByID retrieves user by ID
func (uc *AuthUsecase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
