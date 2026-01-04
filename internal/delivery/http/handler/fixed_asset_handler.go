package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/delivery/http/helper"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	fixedassetuc "github.com/herman-xphp/bukuo/internal/usecase/fixedasset"
	"github.com/shopspring/decimal"
)

// FixedAssetHandler handles fixed asset HTTP endpoints
type FixedAssetHandler struct {
	usecase *fixedassetuc.FixedAssetUsecase
}

// NewFixedAssetHandler creates a new FixedAssetHandler
func NewFixedAssetHandler(uc *fixedassetuc.FixedAssetUsecase) *FixedAssetHandler {
	return &FixedAssetHandler{usecase: uc}
}

type createAssetRequest struct {
	AssetCode          string  `json:"asset_code" binding:"required"`
	Name               string  `json:"name" binding:"required"`
	Description        string  `json:"description"`
	CategoryID         string  `json:"category_id" binding:"required"`
	AcquisitionDate    string  `json:"acquisition_date" binding:"required"`
	AcquisitionCost    float64 `json:"acquisition_cost" binding:"required"`
	ResidualValue      float64 `json:"residual_value"`
	UsefulLifeMonths   int     `json:"useful_life_months" binding:"required"`
	DepreciationMethod string  `json:"depreciation_method" binding:"required"`
	AssetAccountID     string  `json:"asset_account_id" binding:"required"`
	DeprAccountID      string  `json:"depreciation_account_id" binding:"required"`
	AccumDeprAccountID string  `json:"accum_depreciation_account_id" binding:"required"`
	Location           string  `json:"location"`
}

// Create handles POST /assets
func (h *FixedAssetHandler) Create(c *gin.Context) {
	var req createAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	categoryID, _ := uuid.Parse(req.CategoryID)
	assetAccountID, _ := uuid.Parse(req.AssetAccountID)
	deprAccountID, _ := uuid.Parse(req.DeprAccountID)
	accumDeprAccountID, _ := uuid.Parse(req.AccumDeprAccountID)
	acquisitionDate, _ := time.Parse("2006-01-02", req.AcquisitionDate)

	input := fixedassetuc.CreateAssetInput{
		AssetCode:          req.AssetCode,
		Name:               req.Name,
		Description:        req.Description,
		CategoryID:         categoryID,
		AcquisitionDate:    acquisitionDate,
		AcquisitionCost:    decimal.NewFromFloat(req.AcquisitionCost),
		ResidualValue:      decimal.NewFromFloat(req.ResidualValue),
		UsefulLifeMonths:   req.UsefulLifeMonths,
		DepreciationMethod: entity.DepreciationMethod(req.DepreciationMethod),
		AssetAccountID:     assetAccountID,
		DeprAccountID:      deprAccountID,
		AccumDeprAccountID: accumDeprAccountID,
		Location:           req.Location,
	}

	result, err := h.usecase.CreateAsset(c.Request.Context(), helper.GetCompanyID(c), input)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}

// Get handles GET /assets/:id
func (h *FixedAssetHandler) Get(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "asset")
		return
	}

	result, err := h.usecase.GetAsset(c.Request.Context(), helper.GetCompanyID(c), id)
	if err != nil {
		helper.NotFound(c, "asset")
		return
	}

	helper.Success(c, result)
}

// List handles GET /assets
func (h *FixedAssetHandler) List(c *gin.Context) {
	page := helper.GetPage(c)
	pageSize := helper.GetPageSize(c)

	result, total, err := h.usecase.ListAssets(c.Request.Context(), helper.GetCompanyID(c), page, pageSize)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.SuccessWithPagination(c, result, total, page, pageSize)
}

type runDepreciationRequest struct {
	PeriodID string `json:"period_id" binding:"required"`
}

// RunDepreciation handles POST /assets/depreciation
func (h *FixedAssetHandler) RunDepreciation(c *gin.Context) {
	var req runDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	periodID, err := uuid.Parse(req.PeriodID)
	if err != nil {
		helper.BadRequestMessage(c, "invalid period_id")
		return
	}

	entries, err := h.usecase.CalculateDepreciation(
		c.Request.Context(),
		helper.GetCompanyID(c),
		periodID,
		helper.GetUserID(c),
	)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "depreciation calculated", "entries": entries})
}

type disposeAssetRequest struct {
	DisposalAmount float64 `json:"disposal_amount"`
}

// Dispose handles POST /assets/:id/dispose
func (h *FixedAssetHandler) Dispose(c *gin.Context) {
	id, err := helper.ParseUUID(c, "id")
	if err != nil {
		helper.InvalidID(c, "asset")
		return
	}

	var req disposeAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	if err := h.usecase.DisposeAsset(
		c.Request.Context(),
		helper.GetCompanyID(c),
		id,
		decimal.NewFromFloat(req.DisposalAmount),
	); err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, gin.H{"message": "asset disposed"})
}

// ListCategories handles GET /assets/categories
func (h *FixedAssetHandler) ListCategories(c *gin.Context) {
	result, err := h.usecase.ListCategories(c.Request.Context(), helper.GetCompanyID(c))
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Success(c, result)
}

type createCategoryRequest struct {
	Name               string `json:"name" binding:"required"`
	DefaultLifeMonths  int    `json:"default_life_months" binding:"required"`
	DefaultMethod      string `json:"default_method" binding:"required"`
	AssetAccountID     string `json:"asset_account_id" binding:"required"`
	DeprAccountID      string `json:"depreciation_account_id" binding:"required"`
	AccumDeprAccountID string `json:"accum_depreciation_account_id" binding:"required"`
}

// CreateCategory handles POST /assets/categories
func (h *FixedAssetHandler) CreateCategory(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, err)
		return
	}

	assetAccID, _ := uuid.Parse(req.AssetAccountID)
	deprAccID, _ := uuid.Parse(req.DeprAccountID)
	accumDeprAccID, _ := uuid.Parse(req.AccumDeprAccountID)

	result, err := h.usecase.CreateCategory(
		c.Request.Context(),
		helper.GetCompanyID(c),
		req.Name,
		req.DefaultLifeMonths,
		entity.DepreciationMethod(req.DefaultMethod),
		assetAccID,
		deprAccID,
		accumDeprAccID,
	)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	helper.Created(c, result)
}
