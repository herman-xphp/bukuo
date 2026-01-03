package helper

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard response structures
type successResponse struct {
	Data interface{} `json:"data"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// Success sends a 200 OK response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, successResponse{Data: data})
}

// Created sends a 201 Created response with data
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, successResponse{Data: data})
}

// Deleted sends a 200 OK response with delete confirmation message
func Deleted(c *gin.Context, resource string) {
	c.JSON(http.StatusOK, messageResponse{Message: fmt.Sprintf("%s deleted", resource)})
}

// Message sends a 200 OK response with a custom message
func Message(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, messageResponse{Message: msg})
}

// Error sends an error response with specified status code
func Error(c *gin.Context, status int, err error) {
	c.JSON(status, errorResponse{Error: err.Error()})
}

// ErrorMessage sends an error response with specified status code and message
func ErrorMessage(c *gin.Context, status int, message string) {
	c.JSON(status, errorResponse{Error: message})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
}

// BadRequestMessage sends a 400 Bad Request response with message
func BadRequestMessage(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, errorResponse{Error: message})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, errorResponse{Error: fmt.Sprintf("%s not found", resource)})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, err error) {
	c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, errorResponse{Error: message})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, errorResponse{Error: message})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
}

// InternalErrorMessage sends a 500 Internal Server Error response with message
func InternalErrorMessage(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, errorResponse{Error: message})
}

// Paginated sends a 200 OK response with paginated data
func Paginated(c *gin.Context, result PaginatedResult) {
	c.JSON(http.StatusOK, successResponse{Data: result})
}

// PaginatedItems is a convenience method to create and send paginated response
func PaginatedItems(c *gin.Context, items interface{}, total int64, p Pagination) {
	result := NewPaginatedResult(items, total, p)
	Paginated(c, result)
}

// InvalidID sends a 400 Bad Request for invalid ID parameter
func InvalidID(c *gin.Context, resource string) {
	c.JSON(http.StatusBadRequest, errorResponse{Error: fmt.Sprintf("invalid %s id", resource)})
}
