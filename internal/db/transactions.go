package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManager handles database transactions
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

// TransactionFunc represents a function that runs within a transaction
type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Log rollback error but don't override panic
				fmt.Printf("Failed to rollback transaction after panic: %v\n", rollbackErr)
			}
			panic(p) // Re-throw panic
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithReadOnlyTransaction executes a function within a read-only transaction
func (tm *TransactionManager) WithReadOnlyTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read-only transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				fmt.Printf("Failed to rollback read-only transaction after panic: %v\n", rollbackErr)
			}
			panic(p) // Re-throw panic
		}
	}()

	// Set transaction to read-only
	if _, err := tx.Exec(ctx, "SET TRANSACTION READ ONLY"); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction after setting read-only: %w (original error: %v)", rollbackErr, err)
		}
		return fmt.Errorf("failed to set transaction read-only: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("failed to rollback read-only transaction: %w (original error: %v)", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit read-only transaction: %w", err)
	}

	return nil
}

// TransactionalRepository provides a base for repositories that support transactions
type TransactionalRepository struct {
	txManager *TransactionManager
}

// NewTransactionalRepository creates a new transactional repository
func NewTransactionalRepository(txManager *TransactionManager) *TransactionalRepository {
	return &TransactionalRepository{
		txManager: txManager,
	}
}

// WithTransaction executes a function within a transaction
func (tr *TransactionalRepository) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	return tr.txManager.WithTransaction(ctx, fn)
}

// WithReadOnlyTransaction executes a function within a read-only transaction
func (tr *TransactionalRepository) WithReadOnlyTransaction(ctx context.Context, fn TransactionFunc) error {
	return tr.txManager.WithReadOnlyTransaction(ctx, fn)
}

// TransactionalQuery executes a query within a transaction
func (tr *TransactionalRepository) TransactionalQuery(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return tr.txManager.WithTransaction(ctx, fn)
}

// ReadOnlyQuery executes a query within a read-only transaction
func (tr *TransactionalRepository) ReadOnlyQuery(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return tr.txManager.WithReadOnlyTransaction(ctx, fn)
}
