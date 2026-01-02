package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/domain/repository"
)

// Errors
var (
	ErrCategoryNotFound = errors.New("category not found")
)

// CategoryUsecase handles product category business logic
type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryUsecase creates a new CategoryUsecase
func NewCategoryUsecase(cr repository.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{categoryRepo: cr}
}

// CreateCategoryInput represents input for creating a category
type CreateCategoryInput struct {
	CompanyID uuid.UUID
	Name      string
	ParentID  *uuid.UUID
}

// CreateCategory creates a new product category
func (uc *CategoryUsecase) CreateCategory(ctx context.Context, input CreateCategoryInput) (*entity.ProductCategory, error) {
	category := entity.NewProductCategory(input.CompanyID, input.Name, input.ParentID)

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// List retrieves all categories for a company
func (uc *CategoryUsecase) List(ctx context.Context, companyID uuid.UUID) ([]entity.ProductCategory, error) {
	return uc.categoryRepo.List(ctx, companyID)
}

// GetByID retrieves a category by ID
func (uc *CategoryUsecase) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.ProductCategory, error) {
	cat, err := uc.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	return cat, nil
}

// UpdateCategoryInput represents input for updating a category
type UpdateCategoryInput struct {
	CompanyID uuid.UUID
	ID        uuid.UUID
	Name      string
	ParentID  *uuid.UUID
}

// UpdateCategory updates an existing category
func (uc *CategoryUsecase) UpdateCategory(ctx context.Context, input UpdateCategoryInput) (*entity.ProductCategory, error) {
	cat, err := uc.categoryRepo.GetByID(ctx, input.CompanyID, input.ID)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	cat.Name = input.Name
	cat.ParentID = input.ParentID

	if err := uc.categoryRepo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return cat, nil
}

// Delete deletes a category
func (uc *CategoryUsecase) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	_, err := uc.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return ErrCategoryNotFound
	}
	return uc.categoryRepo.Delete(ctx, companyID, id)
}
