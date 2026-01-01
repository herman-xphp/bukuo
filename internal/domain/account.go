package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET" // Aset
	AccountTypeLiability AccountType = "LIABILITY // Utang"
	AccountTypeEquity    AccountType = "EQUITY"  // Modal
	AccountTypeRevenue   AccountType = "REVENUE" // Pendapatan
	AccountTypeExpense   AccountType = "EXPENSE" // Beban
)

type Account struct {
	ID          uuid.UUID   `json:"id"`
	CompanyID   uuid.UUID   `json:"company_id"`
	Code        string      `json:"code"` // "1-1100"
	Name        string      `json:"name"` // "Kas"
	Type        AccountType `json:"type"`
	ParentID    *uuid.UUID  `json:"parent_id"`
	IsPostable  bool        `json:"is_postable"` // bisa diposting?
	IsActive    bool        `json:"is_active"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// NormalBalance - Debit atau Credit?
func (a *Account) NormalBalance() string {
	switch a.Type {
	case AccountTypeAsset, AccountTypeExpense:
		return "DEBIT"
	default:
		return "CREDIT"
	}
}
