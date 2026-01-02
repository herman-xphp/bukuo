package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// ContactFilter defines filter options for listing contacts
type ContactFilter struct {
	ContactType *entity.ContactType
	Search      string // Search by name or code
	IsActive    *bool
	Page        int
	PageSize    int
}

// ContactRepository defines the interface for contact data access
type ContactRepository interface {
	// Create creates a new contact
	Create(ctx context.Context, contact *entity.Contact) error

	// GetByID retrieves a contact by ID
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Contact, error)

	// GetByCode retrieves a contact by code
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entity.Contact, error)

	// List retrieves contacts with filtering and pagination
	List(ctx context.Context, companyID uuid.UUID, filter ContactFilter) ([]entity.Contact, int64, error)

	// Update updates an existing contact
	Update(ctx context.Context, contact *entity.Contact) error

	// Delete soft-deletes a contact
	Delete(ctx context.Context, companyID, id uuid.UUID) error

	// ExistsByCode checks if a contact with the given code exists
	ExistsByCode(ctx context.Context, companyID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error)
}
