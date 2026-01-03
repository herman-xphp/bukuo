package helper

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GetCompanyID extracts company_id from gin context
// Returns uuid.Nil if not found or invalid
func GetCompanyID(c *gin.Context) uuid.UUID {
	id, _ := uuid.Parse(c.GetString("company_id"))
	return id
}

// GetUserID extracts user_id from gin context
// Returns uuid.Nil if not found or invalid
func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := uuid.Parse(c.GetString("user_id"))
	return id
}

// ParseUUID parses a UUID from URL parameter
// Returns the parsed UUID and error if invalid
func ParseUUID(c *gin.Context, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(paramName))
}

// ParseOptionalUUID parses an optional UUID from string pointer
// Returns nil if input is nil or empty, parsed UUID pointer otherwise
func ParseOptionalUUID(value *string) *uuid.UUID {
	if value == nil || *value == "" {
		return nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return nil
	}
	return &id
}

// ParseUUIDString parses a UUID from string
// Returns uuid.Nil if empty or invalid
func ParseUUIDString(value string) uuid.UUID {
	if value == "" {
		return uuid.Nil
	}
	id, _ := uuid.Parse(value)
	return id
}

// ParsePagination extracts pagination params from query string
// Supports: page, page_size, limit, offset
func ParsePagination(c *gin.Context) Pagination {
	p := DefaultPagination()

	if page := c.Query("page"); page != "" {
		if parsed, err := strconv.Atoi(page); err == nil && parsed > 0 {
			p.Page = parsed
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if parsed, err := strconv.Atoi(pageSize); err == nil && parsed > 0 {
			p.PageSize = parsed
		}
	}

	// Also support "limit" as alias for page_size
	if limit := c.Query("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			p.PageSize = parsed
		}
	}

	// Allow direct offset override
	if offset := c.Query("offset"); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil && parsed >= 0 {
			p.Offset = parsed
			// Don't recalculate if offset is explicitly set
			p.Limit = p.PageSize
			return p
		}
	}

	p.Calculate()
	return p
}

// ParseQueryInt parses an integer from query string with default value
func ParseQueryInt(c *gin.Context, key string, defaultVal int) int {
	if val := c.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

// ParseQueryBool parses an optional boolean from query string
// Returns nil if not present, pointer to bool otherwise
func ParseQueryBool(c *gin.Context, key string) *bool {
	if val := c.Query(key); val != "" {
		result := val == "true" || val == "1"
		return &result
	}
	return nil
}

// ParseDecimal parses a decimal from string
// Returns decimal.Zero if empty or invalid
func ParseDecimal(value string) decimal.Decimal {
	if value == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// ParseOptionalDecimal parses decimal only if non-empty
// Returns nil if empty, pointer to decimal otherwise
func ParseOptionalDecimal(value string) *decimal.Decimal {
	if value == "" {
		return nil
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return nil
	}
	return &d
}
