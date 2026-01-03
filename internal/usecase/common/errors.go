package common

import "fmt"

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
