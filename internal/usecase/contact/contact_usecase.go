package contact

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
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
	// Validate contact type
	if !entity.IsValidContactType(input.ContactType) {
		return nil, ErrInvalidContactType
	}

	// Check if code exists
	exists, err := uc.contactRepo.ExistsByCode(ctx, input.CompanyID, input.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check code: %w", err)
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
		return nil, fmt.Errorf("failed to create contact: %w", err)
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
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := repository.ContactFilter{
		ContactType: input.ContactType,
		Search:      input.Search,
		IsActive:    input.IsActive,
		Page:        input.Page,
		PageSize:    input.PageSize,
	}

	contacts, total, err := uc.contactRepo.List(ctx, input.CompanyID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	totalPages := total / int64(input.PageSize)
	if total%int64(input.PageSize) > 0 {
		totalPages++
	}

	return &ListOutput{
		Contacts:   contacts,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
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
	// Validate contact type
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
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return contact, nil
}

// DeleteContact soft-deletes a contact
func (uc *ContactUsecase) DeleteContact(ctx context.Context, companyID, id uuid.UUID) error {
	// Check if contact exists
	_, err := uc.contactRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrContactNotFound
	}

	if err := uc.contactRepo.Delete(ctx, companyID, id); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
