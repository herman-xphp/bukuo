package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// SalesFilter defines filter options
type SalesFilter struct {
	CustomerID *uuid.UUID
	Status     *entity.SalesStatus
	StartDate  *string
	EndDate    *string
	Page       int
	PageSize   int
}

// SalesQuotationRepository defines quotation data access
type SalesQuotationRepository interface {
	Create(ctx context.Context, q *entity.SalesQuotation, lines []entity.SalesQuotationLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesQuotation, error)
	List(ctx context.Context, companyID uuid.UUID, filter SalesFilter) ([]entity.SalesQuotation, int64, error)
	Update(ctx context.Context, q *entity.SalesQuotation) error
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.SalesStatus) error
	Delete(ctx context.Context, companyID, id uuid.UUID) error
}

// SalesOrderRepository defines order data access
type SalesOrderRepository interface {
	Create(ctx context.Context, o *entity.SalesOrder, lines []entity.SalesOrderLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesOrder, error)
	List(ctx context.Context, companyID uuid.UUID, filter SalesFilter) ([]entity.SalesOrder, int64, error)
	Update(ctx context.Context, o *entity.SalesOrder) error
	UpdateStatus(ctx context.Context, companyID, id uuid.UUID, status entity.SalesStatus) error
}

// SalesInvoiceRepository defines invoice data access
type SalesInvoiceRepository interface {
	Create(ctx context.Context, inv *entity.SalesInvoice, lines []entity.SalesInvoiceLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesInvoice, error)
	List(ctx context.Context, companyID uuid.UUID, filter SalesFilter) ([]entity.SalesInvoice, int64, error)
	Update(ctx context.Context, inv *entity.SalesInvoice) error
	UpdatePaidAmount(ctx context.Context, companyID, id uuid.UUID, paidAmount interface{}) error
}

// DeliveryOrderRepository defines delivery data access
type DeliveryOrderRepository interface {
	Create(ctx context.Context, d *entity.DeliveryOrder, lines []entity.DeliveryOrderLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.DeliveryOrder, error)
	List(ctx context.Context, companyID uuid.UUID, filter SalesFilter) ([]entity.DeliveryOrder, int64, error)
}

// SalesReturnRepository defines return data access
type SalesReturnRepository interface {
	Create(ctx context.Context, r *entity.SalesReturn, lines []entity.SalesReturnLine) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.SalesReturn, error)
	List(ctx context.Context, companyID uuid.UUID, filter SalesFilter) ([]entity.SalesReturn, int64, error)
}
