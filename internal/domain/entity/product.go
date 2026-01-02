package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductType represents the type of product
type ProductType string

const (
	ProductTypeGoods   ProductType = "GOODS"
	ProductTypeService ProductType = "SERVICE"
)

// Product represents a product or service item
type Product struct {
	ID                 uuid.UUID       `json:"id"`
	CompanyID          uuid.UUID       `json:"company_id"`
	Code               string          `json:"code"`
	Name               string          `json:"name"`
	Type               ProductType     `json:"type"`
	CategoryID         *uuid.UUID      `json:"category_id,omitempty"`
	UnitID             uuid.UUID       `json:"unit_id"`
	Description        string          `json:"description,omitempty"`
	SalesPrice         decimal.Decimal `json:"sales_price"`
	PurchasePrice      decimal.Decimal `json:"purchase_price"`
	SalesAccountID     uuid.UUID       `json:"sales_account_id"`     // Revenue account
	PurchaseAccountID  uuid.UUID       `json:"purchase_account_id"`  // Expense/COGS account
	InventoryAccountID *uuid.UUID      `json:"inventory_account_id"` // For GOODS only
	IsActive           bool            `json:"is_active"`
	MinStock           decimal.Decimal `json:"min_stock"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// NewProduct creates a new product entity
func NewProduct(companyID uuid.UUID, code, name string, productType ProductType, unitID, salesAccountID, purchaseAccountID uuid.UUID) *Product {
	return &Product{
		ID:                companyID,
		CompanyID:         companyID,
		Code:              code,
		Name:              name,
		Type:              productType,
		UnitID:            unitID,
		SalesAccountID:    salesAccountID,
		PurchaseAccountID: purchaseAccountID,
		SalesPrice:        decimal.Zero,
		PurchasePrice:     decimal.Zero,
		MinStock:          decimal.Zero,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// IsGoods returns true if the product is a physical goods
func (p *Product) IsGoods() bool {
	return p.Type == ProductTypeGoods
}

// IsService returns true if the product is a service
func (p *Product) IsService() bool {
	return p.Type == ProductTypeService
}

// ValidProductTypes returns all valid product types
func ValidProductTypes() []ProductType {
	return []ProductType{ProductTypeGoods, ProductTypeService}
}

// IsValidProductType checks if the product type is valid
func IsValidProductType(t ProductType) bool {
	for _, valid := range ValidProductTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
