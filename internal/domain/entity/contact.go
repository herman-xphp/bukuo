package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ContactType represents the type of contact
type ContactType string

const (
	ContactTypeCustomer ContactType = "CUSTOMER"
	ContactTypeSupplier ContactType = "SUPPLIER"
	ContactTypeBoth     ContactType = "BOTH"
)

// Contact represents a customer or supplier
type Contact struct {
	ID              uuid.UUID       `json:"id"`
	CompanyID       uuid.UUID       `json:"company_id"`
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	ContactType     ContactType     `json:"contact_type"`
	Email           string          `json:"email,omitempty"`
	Phone           string          `json:"phone,omitempty"`
	Address         string          `json:"address,omitempty"`
	City            string          `json:"city,omitempty"`
	TaxID           string          `json:"tax_id,omitempty"` // NPWP
	CreditLimit     decimal.Decimal `json:"credit_limit"`
	PaymentTermDays int             `json:"payment_term_days"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// NewContact creates a new contact entity
func NewContact(companyID uuid.UUID, code, name string, contactType ContactType) *Contact {
	return &Contact{
		ID:              uuid.New(),
		CompanyID:       companyID,
		Code:            code,
		Name:            name,
		ContactType:     contactType,
		CreditLimit:     decimal.Zero,
		PaymentTermDays: 30, // Default 30 days
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// IsCustomer returns true if contact can be used as customer
func (c *Contact) IsCustomer() bool {
	return c.ContactType == ContactTypeCustomer || c.ContactType == ContactTypeBoth
}

// IsSupplier returns true if contact can be used as supplier
func (c *Contact) IsSupplier() bool {
	return c.ContactType == ContactTypeSupplier || c.ContactType == ContactTypeBoth
}

// Customer represents a customer entity
type Customer struct {
	Contact
}

// Supplier represents a supplier entity
type Supplier struct {
	Contact
}

// ValidContactTypes returns all valid contact types
func ValidContactTypes() []ContactType {
	return []ContactType{ContactTypeCustomer, ContactTypeSupplier, ContactTypeBoth}
}

// IsValidContactType checks if the contact type is valid
func IsValidContactType(t ContactType) bool {
	for _, valid := range ValidContactTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
