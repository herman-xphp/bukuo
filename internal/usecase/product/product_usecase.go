package product

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
	ErrProductNotFound    = errors.New("product not found")
	ErrProductCodeExists  = errors.New("product code already exists")
	ErrInvalidProductType = errors.New("invalid product type")
)

// ProductUsecase handles product business logic
type ProductUsecase struct {
	productRepo repository.ProductRepository
}

// NewProductUsecase creates a new ProductUsecase
func NewProductUsecase(pr repository.ProductRepository) *ProductUsecase {
	return &ProductUsecase{productRepo: pr}
}

// CreateProductInput represents input for creating a product
type CreateProductInput struct {
	CompanyID          uuid.UUID
	Code               string
	Name               string
	Type               entity.ProductType
	CategoryID         *uuid.UUID
	UnitID             uuid.UUID
	Description        string
	SalesPrice         decimal.Decimal
	PurchasePrice      decimal.Decimal
	SalesAccountID     uuid.UUID
	PurchaseAccountID  uuid.UUID
	InventoryAccountID *uuid.UUID
	MinStock           decimal.Decimal
}

// CreateProduct creates a new product
func (uc *ProductUsecase) CreateProduct(ctx context.Context, input CreateProductInput) (*entity.Product, error) {
	// Validate product type
	if !entity.IsValidProductType(input.Type) {
		return nil, ErrInvalidProductType
	}

	// Check if code exists
	exists, err := uc.productRepo.ExistsByCode(ctx, input.CompanyID, input.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check code: %w", err)
	}
	if exists {
		return nil, ErrProductCodeExists
	}

	product := &entity.Product{
		ID:                 uuid.New(),
		CompanyID:          input.CompanyID,
		Code:               input.Code,
		Name:               input.Name,
		Type:               input.Type,
		CategoryID:         input.CategoryID,
		UnitID:             input.UnitID,
		Description:        input.Description,
		SalesPrice:         input.SalesPrice,
		PurchasePrice:      input.PurchasePrice,
		SalesAccountID:     input.SalesAccountID,
		PurchaseAccountID:  input.PurchaseAccountID,
		InventoryAccountID: input.InventoryAccountID,
		MinStock:           input.MinStock,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetByID retrieves a product by ID
func (uc *ProductUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// ListInput represents input for listing products
type ListInput struct {
	CompanyID   uuid.UUID
	ProductType *entity.ProductType
	CategoryID  *uuid.UUID
	Search      string
	IsActive    *bool
	Page        int
	PageSize    int
}

// ListOutput represents output for listing products
type ListOutput struct {
	Products   []entity.Product
	Total      int64
	Page       int
	PageSize   int
	TotalPages int64
}

// List retrieves products with filtering and pagination
func (uc *ProductUsecase) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := repository.ProductFilter{
		ProductType: input.ProductType,
		CategoryID:  input.CategoryID,
		Search:      input.Search,
		IsActive:    input.IsActive,
		Page:        input.Page,
		PageSize:    input.PageSize,
	}

	products, total, err := uc.productRepo.List(ctx, input.CompanyID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	totalPages := total / int64(input.PageSize)
	if total%int64(input.PageSize) > 0 {
		totalPages++
	}

	return &ListOutput{
		Products:   products,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateProductInput represents input for updating a product
type UpdateProductInput struct {
	CompanyID          uuid.UUID
	ID                 uuid.UUID
	Name               string
	Type               entity.ProductType
	CategoryID         *uuid.UUID
	UnitID             uuid.UUID
	Description        string
	SalesPrice         decimal.Decimal
	PurchasePrice      decimal.Decimal
	SalesAccountID     uuid.UUID
	PurchaseAccountID  uuid.UUID
	InventoryAccountID *uuid.UUID
	MinStock           decimal.Decimal
	IsActive           bool
}

// UpdateProduct updates an existing product
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, input UpdateProductInput) (*entity.Product, error) {
	// Validate product type
	if !entity.IsValidProductType(input.Type) {
		return nil, ErrInvalidProductType
	}

	product, err := uc.productRepo.GetByID(ctx, input.CompanyID, input.ID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	product.Name = input.Name
	product.Type = input.Type
	product.CategoryID = input.CategoryID
	product.UnitID = input.UnitID
	product.Description = input.Description
	product.SalesPrice = input.SalesPrice
	product.PurchasePrice = input.PurchasePrice
	product.SalesAccountID = input.SalesAccountID
	product.PurchaseAccountID = input.PurchaseAccountID
	product.InventoryAccountID = input.InventoryAccountID
	product.MinStock = input.MinStock
	product.IsActive = input.IsActive
	product.UpdatedAt = time.Now()

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteProduct soft-deletes a product
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, companyID, id uuid.UUID) error {
	_, err := uc.productRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrProductNotFound
	}

	if err := uc.productRepo.Delete(ctx, companyID, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
