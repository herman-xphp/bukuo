package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/config"
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
	auditRepo   repository.AuditLogRepository
	jwtService  *JWTService
	securityCfg config.SecurityConfig
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(
	ur repository.UserRepository,
	cr repository.CompanyRepository,
	jwt *JWTService,
	audit repository.AuditLogRepository,
	secCfg config.SecurityConfig,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    ur,
		companyRepo: cr,
		auditRepo:   audit,
		jwtService:  jwt,
		securityCfg: secCfg,
	}
}

// RegisterInput represents registration input
type RegisterInput struct {
	CompanyName string
	Email       string
	Password    string
	Name        string
	IPAddress   string
	UserAgent   string
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

	// Create owner user (password validation happens in NewUser)
	user, err := entity.NewUser(company.ID, input.Email, input.Password, input.Name, entity.UserRoleOwner)
	if err != nil {
		return nil, err // Returns ErrWeakPassword or ErrInvalidEmail
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Audit log
	auditLog := entity.NewAuditLog(
		company.ID, &user.ID, user.Email,
		entity.AuditActionCreate, "USER", &user.ID,
		"User registered: "+user.Name,
		input.IPAddress, input.UserAgent,
	)
	_ = uc.auditRepo.Create(ctx, auditLog)

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
	Email     string
	Password  string
	IPAddress string
	UserAgent string
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
		// Log failed attempt (user not found)
		uc.logFailedLogin(ctx, uuid.Nil, input.Email, input.IPAddress, input.UserAgent, "user not found")
		return nil, entity.ErrInvalidCredentials
	}

	// Check if account is locked
	if user.IsLocked() {
		uc.logFailedLogin(ctx, user.CompanyID, input.Email, input.IPAddress, input.UserAgent, "account locked")
		return nil, entity.ErrAccountLocked
	}

	// Check password
	if err := user.CheckPassword(input.Password); err != nil {
		// Record failed attempt
		user.RecordFailedAttempt(uc.securityCfg.MaxLoginAttempts, uc.securityCfg.LockoutDuration)
		_ = uc.userRepo.Update(ctx, user)

		uc.logFailedLogin(ctx, user.CompanyID, input.Email, input.IPAddress, input.UserAgent, "wrong password")
		return nil, err
	}

	// Success - reset failed attempts and update last login
	user.UpdateLastLogin()
	_ = uc.userRepo.Update(ctx, user)

	// Audit log - successful login
	auditLog := entity.NewAuditLog(
		user.CompanyID, &user.ID, user.Email,
		entity.AuditActionLogin, "USER", &user.ID,
		"User logged in",
		input.IPAddress, input.UserAgent,
	)
	_ = uc.auditRepo.Create(ctx, auditLog)

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

// logFailedLogin records a failed login attempt
func (uc *AuthUsecase) logFailedLogin(ctx context.Context, companyID uuid.UUID, email, ip, ua, reason string) {
	auditLog := entity.NewAuditLog(
		companyID, nil, email,
		entity.AuditActionFailed, "AUTH", nil,
		"Failed login: "+reason,
		ip, ua,
	)
	_ = uc.auditRepo.Create(ctx, auditLog)
}

// GetUserByID retrieves user by ID
func (uc *AuthUsecase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
