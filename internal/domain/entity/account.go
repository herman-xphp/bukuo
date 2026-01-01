package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Errors
var (
	ErrAccountNotPostable = errors.New("account is not postable")
)

// AccountType represents the type of account per PSAK
type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"
	AccountTypeLiability AccountType = "LIABILITY"
	AccountTypeEquity    AccountType = "EQUITY"
	AccountTypeRevenue   AccountType = "REVENUE"
	AccountTypeExpense   AccountType = "EXPENSE"
)

// Account represents a chart of accounts entry
type Account struct {
	ID          uuid.UUID   `json:"id"`
	CompanyID   uuid.UUID   `json:"company_id"`
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	Type        AccountType `json:"type"`
	ParentID    *uuid.UUID  `json:"parent_id"`
	IsPostable  bool        `json:"is_postable"`
	IsActive    bool        `json:"is_active"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// NewAccount creates a new account entity
func NewAccount(companyID uuid.UUID, code, name string, accType AccountType) *Account {
	return &Account{
		ID:         uuid.New(),
		CompanyID:  companyID,
		Code:       code,
		Name:       name,
		Type:       accType,
		IsPostable: true,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// NormalBalance returns whether this account normally has a debit or credit balance
func (a *Account) NormalBalance() string {
	switch a.Type {
	case AccountTypeAsset, AccountTypeExpense:
		return "DEBIT"
	default:
		return "CREDIT"
	}
}

// CanPost checks if the account can be posted to
func (a *Account) CanPost() error {
	if !a.IsPostable {
		return ErrAccountNotPostable
	}
	return nil
}
