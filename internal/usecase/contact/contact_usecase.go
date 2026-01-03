package contact

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
	"github.com/shopspring/decimal"
)

// Errors
var (
	ErrContactNotFound    = errors.New("contact not found")
	ErrContactCodeExists  = errors.New("contact code already exists")
	ErrInvalidContactType = errors.New("invalid contact type")
)

// ContactUsecase handles contact business logic
type ContactUsecase struct {
	contactRepo repository.ContactRepository
}

// NewContactUsecase creates a new ContactUsecase
func NewContactUsecase(cr repository.ContactRepository) *ContactUsecase {
	return &ContactUsecase{contactRepo: cr}
}

// CreateContactInput represents input for creating a contact
type CreateContactInput struct {
	CompanyID       uuid.UUID
	Code            string
	Name            string
	ContactType     entity.ContactType
	Email           string
	Phone           string
	Address         string
	City            string
	TaxID           string
	CreditLimit     decimal.Decimal
	PaymentTermDays int
}

// CreateContact creates a new contact
func (uc *ContactUsecase) CreateContact(ctx context.Context, input CreateContactInput) (*entity.Contact, error) {
	if !entity.IsValidContactType(input.ContactType) {
		return nil, ErrInvalidContactType
	}

	exists, err := uc.contactRepo.ExistsByCode(ctx, input.CompanyID, input.Code, nil)
	if err != nil {
		return nil, common.WrapErr("check code", err)
	}
	if exists {
		return nil, ErrContactCodeExists
	}

	contact := entity.NewContact(input.CompanyID, input.Code, input.Name, input.ContactType)
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.Address = input.Address
	contact.City = input.City
	contact.TaxID = input.TaxID
	if !input.CreditLimit.IsZero() {
		contact.CreditLimit = input.CreditLimit
	}
	if input.PaymentTermDays > 0 {
		contact.PaymentTermDays = input.PaymentTermDays
	}

	if err := uc.contactRepo.Create(ctx, contact); err != nil {
		return nil, common.WrapErr("create contact", err)
	}

	return contact, nil
}

// GetByID retrieves a contact by ID
func (uc *ContactUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Contact, error) {
	contact, err := uc.contactRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrContactNotFound
	}
	return contact, nil
}

// ListInput represents input for listing contacts
type ListInput struct {
	CompanyID   uuid.UUID
	ContactType *entity.ContactType
	Search      string
	IsActive    *bool
	Page        int
	PageSize    int
}

// ListOutput represents output for listing contacts
type ListOutput struct {
	Contacts   []entity.Contact
	Total      int64
	Page       int
	PageSize   int
	TotalPages int64
}

// List retrieves contacts with filtering and pagination
func (uc *ContactUsecase) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	p := common.ValidatePagination(input.Page, input.PageSize)

	filter := repository.ContactFilter{
		ContactType: input.ContactType,
		Search:      input.Search,
		IsActive:    input.IsActive,
		Page:        p.Page,
		PageSize:    p.PageSize,
	}

	contacts, total, err := uc.contactRepo.List(ctx, input.CompanyID, filter)
	if err != nil {
		return nil, common.WrapErr("list contacts", err)
	}

	return &ListOutput{
		Contacts:   contacts,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: common.CalculateTotalPages(total, p.PageSize),
	}, nil
}

// UpdateContactInput represents input for updating a contact
type UpdateContactInput struct {
	CompanyID       uuid.UUID
	ID              uuid.UUID
	Name            string
	ContactType     entity.ContactType
	Email           string
	Phone           string
	Address         string
	City            string
	TaxID           string
	CreditLimit     decimal.Decimal
	PaymentTermDays int
	IsActive        bool
}

// UpdateContact updates an existing contact
func (uc *ContactUsecase) UpdateContact(ctx context.Context, input UpdateContactInput) (*entity.Contact, error) {
	if !entity.IsValidContactType(input.ContactType) {
		return nil, ErrInvalidContactType
	}

	contact, err := uc.contactRepo.GetByID(ctx, input.CompanyID, input.ID)
	if err != nil {
		return nil, ErrContactNotFound
	}

	contact.Name = input.Name
	contact.ContactType = input.ContactType
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.Address = input.Address
	contact.City = input.City
	contact.TaxID = input.TaxID
	contact.CreditLimit = input.CreditLimit
	contact.PaymentTermDays = input.PaymentTermDays
	contact.IsActive = input.IsActive
	contact.UpdatedAt = time.Now()

	if err := uc.contactRepo.Update(ctx, contact); err != nil {
		return nil, common.WrapErr("update contact", err)
	}

	return contact, nil
}

// DeleteContact soft-deletes a contact
func (uc *ContactUsecase) DeleteContact(ctx context.Context, companyID, id uuid.UUID) error {
	if _, err := uc.contactRepo.GetByID(ctx, companyID, id); err != nil {
		return ErrContactNotFound
	}

	if err := uc.contactRepo.Delete(ctx, companyID, id); err != nil {
		return common.WrapErr("delete contact", err)
	}

	return nil
}
