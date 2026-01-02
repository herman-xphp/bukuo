package mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

// JournalRepositoryMock mocks the JournalRepository interface
type JournalRepositoryMock struct {
	mock.Mock
}

func (m *JournalRepositoryMock) Create(ctx context.Context, journal *entity.JournalEntry) error {
	args := m.Called(ctx, journal)
	return args.Error(0)
}

func (m *JournalRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.JournalEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) GetByEntryNumber(ctx context.Context, companyID uuid.UUID, number string) (*entity.JournalEntry, error) {
	args := m.Called(ctx, companyID, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) GetByPeriod(ctx context.Context, periodID uuid.UUID) ([]entity.JournalEntry, error) {
	args := m.Called(ctx, periodID)
	return args.Get(0).([]entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.JournalEntry, error) {
	args := m.Called(ctx, companyID, start, end)
	return args.Get(0).([]entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) GetByStatus(ctx context.Context, companyID uuid.UUID, status entity.JournalStatus) ([]entity.JournalEntry, error) {
	args := m.Called(ctx, companyID, status)
	return args.Get(0).([]entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) Update(ctx context.Context, journal *entity.JournalEntry) error {
	args := m.Called(ctx, journal)
	return args.Error(0)
}
func (m *JournalRepositoryMock) UpdateDetails(ctx context.Context, journal *entity.JournalEntry) error {
	args := m.Called(ctx, journal)
	return args.Error(0)
}
func (m *JournalRepositoryMock) CountByYear(ctx context.Context, companyID uuid.UUID, year int) (int, error) {
	args := m.Called(ctx, companyID, year)
	return args.Int(0), args.Error(1)
}

func (m *JournalRepositoryMock) CreateReversalWithTransaction(ctx context.Context, reversal *entity.JournalEntry, originalID uuid.UUID) error {
	args := m.Called(ctx, reversal, originalID)
	return args.Error(0)
}

func (m *JournalRepositoryMock) ClosePeriodWithTransaction(ctx context.Context, closingJournal *entity.JournalEntry, period *entity.AccountingPeriod) error {
	args := m.Called(ctx, closingJournal, period)
	return args.Error(0)
}

func (m *JournalRepositoryMock) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.JournalEntry, error) {
	args := m.Called(ctx, companyID, limit, offset, search)
	return args.Get(0).([]entity.JournalEntry), args.Error(1)
}

func (m *JournalRepositoryMock) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	args := m.Called(ctx, companyID, search)
	return args.Int(0), args.Error(1)
}

func (m *JournalRepositoryMock) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// UserRepositoryMock mocks the UserRepository interface
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.User, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *UserRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AccountRepositoryMock mocks the AccountRepository interface
type AccountRepositoryMock struct {
	mock.Mock
}

func (m *AccountRepositoryMock) Create(ctx context.Context, account *entity.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *AccountRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *AccountRepositoryMock) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Account, error) {
	args := m.Called(ctx, companyID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *AccountRepositoryMock) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.Account, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]entity.Account), args.Error(1)
}

func (m *AccountRepositoryMock) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.Account, error) {
	args := m.Called(ctx, companyID, limit, offset, search)
	return args.Get(0).([]entity.Account), args.Error(1)
}

func (m *AccountRepositoryMock) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	args := m.Called(ctx, companyID, search)
	return args.Int(0), args.Error(1)
}

func (m *AccountRepositoryMock) Update(ctx context.Context, account *entity.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *AccountRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *UserRepositoryMock) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.User, error) {
	args := m.Called(ctx, companyID, limit, offset, search)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *UserRepositoryMock) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	args := m.Called(ctx, companyID, search)
	return args.Int(0), args.Error(1)
}

func (m *AccountRepositoryMock) UpdateBalance(ctx context.Context, id uuid.UUID, amount decimal.Decimal, isDebit bool) error {
	args := m.Called(ctx, id, amount, isDebit)
	return args.Error(0)
}

func (m *AccountRepositoryMock) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*entity.Account, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*entity.Account), args.Error(1)
}

// PeriodRepositoryMock mocks the PeriodRepository interface
type PeriodRepositoryMock struct {
	mock.Mock
}

func (m *PeriodRepositoryMock) Create(ctx context.Context, period *entity.AccountingPeriod) error {
	args := m.Called(ctx, period)
	return args.Error(0)
}

func (m *PeriodRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccountingPeriod, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccountingPeriod), args.Error(1)
}

func (m *PeriodRepositoryMock) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]entity.AccountingPeriod), args.Error(1)
}

func (m *PeriodRepositoryMock) GetByDate(ctx context.Context, companyID uuid.UUID, date time.Time) (*entity.AccountingPeriod, error) {
	args := m.Called(ctx, companyID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccountingPeriod), args.Error(1)
}

func (m *PeriodRepositoryMock) GetOpenPeriods(ctx context.Context, companyID uuid.UUID) ([]entity.AccountingPeriod, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]entity.AccountingPeriod), args.Error(1)
}

func (m *PeriodRepositoryMock) Update(ctx context.Context, period *entity.AccountingPeriod) error {
	args := m.Called(ctx, period)
	return args.Error(0)
}

func (m *PeriodRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *PeriodRepositoryMock) List(ctx context.Context, companyID uuid.UUID, limit, offset int, search string) ([]entity.AccountingPeriod, error) {
	args := m.Called(ctx, companyID, limit, offset, search)
	return args.Get(0).([]entity.AccountingPeriod), args.Error(1)
}

func (m *PeriodRepositoryMock) Count(ctx context.Context, companyID uuid.UUID, search string) (int, error) {
	args := m.Called(ctx, companyID, search)
	return args.Int(0), args.Error(1)
}

// AuditLogRepositoryMock mocks the AuditLogRepository interface
type AuditLogRepositoryMock struct {
	mock.Mock
}

func (m *AuditLogRepositoryMock) Create(ctx context.Context, log *entity.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *AuditLogRepositoryMock) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	args := m.Called(ctx, companyID, limit, offset)
	return args.Get(0).([]entity.AuditLog), args.Error(1)
}

func (m *AuditLogRepositoryMock) GetByDateRange(ctx context.Context, companyID uuid.UUID, start, end time.Time) ([]entity.AuditLog, error) {
	args := m.Called(ctx, companyID, start, end)
	return args.Get(0).([]entity.AuditLog), args.Error(1)
}

func (m *AuditLogRepositoryMock) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.AuditLog, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).([]entity.AuditLog), args.Error(1)
}

func (m *AuditLogRepositoryMock) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]entity.AuditLog), args.Error(1)
}
