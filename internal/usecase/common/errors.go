package common

import (
	"errors"
	"fmt"
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// WrapErr wraps an error with operation context.
// Usage: return nil, WrapErr("create product", err)
// Output: "failed to create product: <original error>"
func WrapErr(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// WrapErrIf wraps error only if err is not nil, returns nil otherwise.
// This is useful for inline returns.
func WrapErrIf(err error, operation string) error {
	return WrapErr(operation, err)
}

// CheckErr is a helper for operations that return only error.
// Usage: if err := CheckErr(repo.Delete(ctx, id), "delete product"); err != nil { return err }
func CheckErr(err error, operation string) error {
	return WrapErr(operation, err)
}
