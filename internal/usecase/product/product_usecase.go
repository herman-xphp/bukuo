package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
	"github.com/herman-xphp/bukuo/internal/usecase/common"
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
	productRepo   repository.ProductRepository
	inventoryRepo repository.InventoryRepository
}

// NewProductUsecase creates a new ProductUsecase
func NewProductUsecase(pr repository.ProductRepository, ir repository.InventoryRepository) *ProductUsecase {
	return &ProductUsecase{productRepo: pr, inventoryRepo: ir}
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
	if !entity.IsValidProductType(input.Type) {
		return nil, ErrInvalidProductType
	}

	exists, err := uc.productRepo.ExistsByCode(ctx, input.CompanyID, input.Code, nil)
	if err != nil {
		return nil, common.WrapErr("check code", err)
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
		return nil, common.WrapErr("create product", err)
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
	p := common.ValidatePagination(input.Page, input.PageSize)

	filter := repository.ProductFilter{
		ProductType: input.ProductType,
		CategoryID:  input.CategoryID,
		Search:      input.Search,
		IsActive:    input.IsActive,
		Page:        p.Page,
		PageSize:    p.PageSize,
	}

	products, total, err := uc.productRepo.List(ctx, input.CompanyID, filter)
	if err != nil {
		return nil, common.WrapErr("list products", err)
	}

	return &ListOutput{
		Products:   products,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: common.CalculateTotalPages(total, p.PageSize),
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
	ImageURL           *string
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
	product.ImageURL = input.ImageURL
	product.SalesPrice = input.SalesPrice
	product.PurchasePrice = input.PurchasePrice
	product.SalesAccountID = input.SalesAccountID
	product.PurchaseAccountID = input.PurchaseAccountID
	product.InventoryAccountID = input.InventoryAccountID
	product.MinStock = input.MinStock
	product.IsActive = input.IsActive
	product.UpdatedAt = time.Now()

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, common.WrapErr("update product", err)
	}

	return product, nil
}

// DeleteProduct soft-deletes a product
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, companyID, id uuid.UUID) error {
	if _, err := uc.productRepo.GetByID(ctx, companyID, id); err != nil {
		return ErrProductNotFound
	}

	// Check for existing inventory
	stock, err := uc.inventoryRepo.GetTotalStock(ctx, companyID, id)
	if err == nil && stock != nil && !stock.Quantity.IsZero() {
		return fmt.Errorf("cannot delete product with existing stock: %s", stock.Quantity)
	}

	if err := uc.productRepo.Delete(ctx, companyID, id); err != nil {
		return common.WrapErr("delete product", err)
	}

	return nil
}
