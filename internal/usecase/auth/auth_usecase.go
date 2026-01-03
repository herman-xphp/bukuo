package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
)

var (
	ErrEmailExists = errors.New("email already registered")
)

// AuthUsecase handles authentication business logic
type AuthUsecase struct {
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	auditRepo   repository.AuditLogRepository
	txManager   repository.TransactionManager
	jwtService  *JWTService
	securityCfg config.SecurityConfig
	audit       *common.AuditLogger
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(
	ur repository.UserRepository,
	cr repository.CompanyRepository,
	jwt *JWTService,
	auditRepo repository.AuditLogRepository,
	tm repository.TransactionManager,
	secCfg config.SecurityConfig,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    ur,
		companyRepo: cr,
		auditRepo:   auditRepo,
		txManager:   tm,
		jwtService:  jwt,
		securityCfg: secCfg,
		audit:       common.NewAuditLogger(auditRepo),
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
	var user *entity.User
	var company *entity.Company

	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrEmailExists
	}

	err := uc.txManager.RunAtomic(ctx, func(ctx context.Context) error {
		var err error
		company = entity.NewCompany(input.CompanyName, "")
		if err = uc.companyRepo.Create(ctx, company); err != nil {
			return common.WrapErr("create company", err)
		}

		user, err = entity.NewUserWithCost(company.ID, input.Email, input.Password, input.Name, entity.UserRoleOwner, uc.securityCfg.BcryptCost)
		if err != nil {
			return err
		}

		if err := uc.userRepo.Create(ctx, user); err != nil {
			return common.WrapErr("save user", err)
		}

		uc.audit.Log(ctx, common.LogInput{
			CompanyID: company.ID, UserID: &user.ID, UserEmail: user.Email,
			Action: entity.AuditActionCreate, EntityType: "USER", EntityID: &user.ID,
			Description: "User registered", IPAddress: input.IPAddress, UserAgent: input.UserAgent,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	token, err := uc.jwtService.GenerateToken(user.ID, company.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, common.WrapErr("generate token", err)
	}

	return &RegisterOutput{User: user, Company: company, Token: token}, nil
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
		uc.audit.LogFailedLogin(ctx, uuid.Nil, input.Email, "user not found", input.IPAddress, input.UserAgent)
		return nil, entity.ErrInvalidCredentials
	}

	if user.IsLocked() {
		uc.audit.LogFailedLogin(ctx, user.CompanyID, input.Email, "account locked", input.IPAddress, input.UserAgent)
		return nil, entity.ErrAccountLocked
	}

	if err := user.CheckPassword(input.Password); err != nil {
		user.RecordFailedAttempt(uc.securityCfg.MaxLoginAttempts, uc.securityCfg.LockoutDuration)
		_ = uc.userRepo.Update(ctx, user)
		uc.audit.LogFailedLogin(ctx, user.CompanyID, input.Email, "wrong password", input.IPAddress, input.UserAgent)
		return nil, err
	}

	user.UpdateLastLogin()
	_ = uc.userRepo.Update(ctx, user)

	uc.audit.LogLogin(ctx, user.CompanyID, &user.ID, user.Email, input.IPAddress, input.UserAgent)

	token, err := uc.jwtService.GenerateToken(user.ID, user.CompanyID, user.Email, string(user.Role))
	if err != nil {
		return nil, common.WrapErr("generate token", err)
	}

	return &LoginOutput{User: user, Token: token}, nil
}

// GetUserByID retrieves user by ID
func (uc *AuthUsecase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

// UpdateProfileInput represents update profile input
type UpdateProfileInput struct {
	Name           string
	ProfilePicture *string
}

// UpdateProfile updates user's own profile
func (uc *AuthUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.ProfilePicture != nil {
		user.ProfilePicture = input.ProfilePicture
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, common.WrapErr("update profile", err)
	}

	return user, nil
}

// ChangePassword changes user's own password
func (uc *AuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.CheckPassword(oldPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
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

func (uc *AuthUsecase) VerifyPin(ctx context.Context, userID uuid.UUID, pin string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	return user.CheckPin(pin)
}
