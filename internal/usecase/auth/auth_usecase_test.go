package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// Implement other methods as no-ops or panics if not used in Register
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return nil, nil
}
func (m *MockUserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockUserRepository) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	return 0, nil
}

type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) Create(ctx context.Context, company *entity.Company) error {
	args := m.Called(ctx, company)
	// Simulate ID assignment
	if company.ID == uuid.Nil {
		company.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockCompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	return nil, nil
}
func (m *MockCompanyRepository) Update(ctx context.Context, company *entity.Company) error {
	return nil
}
func (m *MockCompanyRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *MockAuditLogRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	return nil, nil
}
func (m *MockAuditLogRepository) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.AuditLog, error) {
	return nil, nil
}
func (m *MockAuditLogRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	return nil, nil
}
func (m *MockAuditLogRepository) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.AuditLog, error) {
	return nil, nil
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) RunAtomic(ctx context.Context, fn func(ctx context.Context) error) error {
	// Directly execute the function callback to simulate successful transaction start
	// In a unit test, we don't need actual DB transaction isolation, just logic flow
	return fn(ctx)
}

// --- Tests ---

func TestAuthUsecase_Register(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockCompanyRepo := new(MockCompanyRepository)
	mockAuditRepo := new(MockAuditLogRepository)
	mockTxManager := new(MockTransactionManager)

	// Config
	secCfg := config.SecurityConfig{
		AllowedOrigins: []string{"*"},
	}
	jwtService := auth.NewJWTService("secret", time.Hour)

	uc := auth.NewAuthUsecase(mockUserRepo, mockCompanyRepo, jwtService, mockAuditRepo, mockTxManager, secCfg)

	t.Run("Success", func(t *testing.T) {
		input := auth.RegisterInput{
			CompanyName: "Test Corp",
			Email:       "owner@test.com",
			Password:    "Apassword123", // Strong password
			Name:        "Owner",
			IPAddress:   "127.0.0.1",
			UserAgent:   "GoTest",
		}

		// Expectations
		mockUserRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, nil) // No existing user
		mockCompanyRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Company) bool {
			return c.Name == input.CompanyName
		})).Return(nil)
		mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Email == input.Email && u.Role == entity.UserRoleOwner
		})).Return(nil)
		mockAuditRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *entity.AuditLog) bool {
			return a.Action == entity.AuditActionCreate && a.UserEmail == input.Email
		})).Return(nil)

		// Execute
		output, err := uc.Register(context.Background(), input)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, input.Email, output.User.Email)
		assert.NotEmpty(t, output.Token)

		mockUserRepo.AssertExpectations(t)
		mockCompanyRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		input := auth.RegisterInput{
			CompanyName: "Test Corp",
			Email:       "existing@test.com",
			Password:    "Apassword123",
		}

		existingUser := &entity.User{Email: input.Email}

		// Expectations
		mockUserRepo.On("GetByEmail", mock.Anything, input.Email).Return(existingUser, nil)

		// Execute
		output, err := uc.Register(context.Background(), input)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, auth.ErrEmailExists, err)
	})

	t.Run("WeakPassword", func(t *testing.T) {
		input := auth.RegisterInput{
			CompanyName: "Weak Corp",
			Email:       "weak@test.com",
			Password:    "123", // Weak
		}

		mockUserRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, nil)
		mockCompanyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		// Execute
		output, err := uc.Register(context.Background(), input)

		// Verify
		assert.ErrorIs(t, err, entity.ErrWeakPassword)
		assert.Nil(t, output)
	})
}
