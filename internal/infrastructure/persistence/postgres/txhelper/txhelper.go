package txhelper

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tx represents a transaction with helper methods
type Tx struct {
	tx pgx.Tx
}

// RunInTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func RunInTx(ctx context.Context, db *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ExecBatch executes multiple queries in a single transaction.
// Each query is a struct containing the SQL and arguments.
type Query struct {
	SQL  string
	Args []interface{}
}

// ExecBatch executes multiple queries in a single transaction
func ExecBatch(ctx context.Context, db *pgxpool.Pool, queries []Query) error {
	return RunInTx(ctx, db, func(tx pgx.Tx) error {
		for _, q := range queries {
			if _, err := tx.Exec(ctx, q.SQL, q.Args...); err != nil {
				return err
			}
		}
		return nil
	})
}

// InsertMany inserts multiple rows using a single query pattern
func InsertMany[T any](ctx context.Context, tx pgx.Tx, query string, items []T, argsFn func(item T) []interface{}) error {
	for _, item := range items {
		if _, err := tx.Exec(ctx, query, argsFn(item)...); err != nil {
			return err
		}
	}
	return nil
}
