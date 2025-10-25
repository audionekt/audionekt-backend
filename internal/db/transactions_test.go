package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTransactionManager_WithTransaction(t *testing.T) {
	// This test requires a real database connection
	// In a real test environment, you'd use a test database
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a test database connection
	// Note: This would need to be configured for your test environment
	config, err := pgxpool.ParseConfig("postgres://dev:devpassword@localhost:5432/musicapp_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}
	defer pool.Close()

	txManager := NewTransactionManager(pool)

	t.Run("successful transaction", func(t *testing.T) {
		err := txManager.WithTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			// Simple query to test transaction
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("transaction with error", func(t *testing.T) {
		testErr := errors.New("test error")
		err := txManager.WithTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			return testErr
		})

		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
	})

	t.Run("transaction with panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic to be re-thrown")
			}
		}()

		err := txManager.WithTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			panic("test panic")
		})

		// This should not be reached due to panic
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestTransactionManager_WithReadOnlyTransaction(t *testing.T) {
	// This test requires a real database connection
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config, err := pgxpool.ParseConfig("postgres://dev:devpassword@localhost:5432/musicapp_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}
	defer pool.Close()

	txManager := NewTransactionManager(pool)

	t.Run("successful read-only transaction", func(t *testing.T) {
		err := txManager.WithReadOnlyTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			// Simple read query
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("read-only transaction with error", func(t *testing.T) {
		testErr := errors.New("test error")
		err := txManager.WithReadOnlyTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			return testErr
		})

		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
	})
}

func TestTransactionalRepository(t *testing.T) {
	// This test requires a real database connection
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config, err := pgxpool.ParseConfig("postgres://dev:devpassword@localhost:5432/musicapp_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}
	defer pool.Close()

	txManager := NewTransactionManager(pool)
	repo := NewTransactionalRepository(txManager)

	t.Run("transactional query", func(t *testing.T) {
		err := repo.TransactionalQuery(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("read-only query", func(t *testing.T) {
		err := repo.ReadOnlyQuery(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestTransactionManager_ContextTimeout(t *testing.T) {
	// This test requires a real database connection
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config, err := pgxpool.ParseConfig("postgres://dev:devpassword@localhost:5432/musicapp_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Skip("Skipping test - no test database available")
		return
	}
	defer pool.Close()

	txManager := NewTransactionManager(pool)

	t.Run("transaction with context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Wait a bit to ensure context times out
		time.Sleep(2 * time.Millisecond)

		err := txManager.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})

		if err == nil {
			t.Error("Expected context timeout error")
		}
	})
}
