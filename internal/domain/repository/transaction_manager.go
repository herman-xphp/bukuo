package repository

import (
	"context"
)

// TransactionManager handles database transactions
type TransactionManager interface {
	// RunAtomic executes the given function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// If the function returns nil, the transaction is committed.
	RunAtomic(ctx context.Context, fn func(ctx context.Context) error) error
}
