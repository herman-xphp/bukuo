package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const (
	txKey contextKey = "db_tx"
)

// TransactionManager implements repository.TransactionManager for PostgreSQL
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a new TransactionManager
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// RunAtomic executes the function within a transaction
func (tm *TransactionManager) RunAtomic(ctx context.Context, fn func(ctx context.Context) error) error {
	// If a transaction is already in progress, reuse it (nested transaction support could be added here, but simple reuse is safer for now)
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback (will be ignored if committed)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Inject tx into context
	txCtx := context.WithValue(ctx, txKey, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DBExecutor defines the common interface for pgxpool.Pool and pgx.Tx
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// getExecutor returns the transaction from context if it exists, otherwise the pool
func (tm *TransactionManager) getExecutor(ctx context.Context) DBExecutor {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return tm.pool
}

// GetExecutorExported allows repositories in the same package to access the helper
// Since they are in the same package 'postgres', they can access unexported methods if I attach them to a shared struct or just use a standalone function.
// But better pattern:
// Repositories should use a helper function.
func GetExecutor(ctx context.Context, pool *pgxpool.Pool) DBExecutor {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}
