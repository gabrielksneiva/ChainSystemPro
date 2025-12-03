package database

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewDatabase(t *testing.T) {
	ctx := context.Background()

	// Start postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if termErr := testcontainers.TerminateContainer(pgContainer); termErr != nil {
			t.Logf("failed to terminate container: %s", termErr)
		}
	}()

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := Config{
		Host:     host,
		Port:     port.Int(),
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	t.Run("Ping", func(t *testing.T) {
		err := db.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("WithTransaction Success", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec("CREATE TABLE test_table (id SERIAL PRIMARY KEY)")
			return err
		})
		assert.NoError(t, err)

		var exists bool
		err = db.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_table')")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("WithTransaction Rollback", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec("CREATE TABLE test_rollback (id SERIAL PRIMARY KEY)")
			if err != nil {
				return err
			}
			return assert.AnError // Force rollback
		})
		assert.Error(t, err)

		var exists bool
		err = db.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_rollback')")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("WithTransaction Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but got none")
			}
		}()

		_ = db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
			panic("test panic")
		})
	})

	t.Run("Close", func(t *testing.T) {
		// Create a new connection just for this test
		cfg := Config{
			Host:     host,
			Port:     port.Int(),
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		}
		db2, err := New(cfg)
		require.NoError(t, err)

		err = db2.Close()
		assert.NoError(t, err)
	})
}
func TestNewDatabase_InvalidConfig(t *testing.T) {
	cfg := Config{
		Host:     "invalid-host",
		Port:     9999,
		User:     "invalid",
		Password: "invalid",
		DBName:   "invalid",
		SSLMode:  "disable",
	}

	_, err := New(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestRunMigrations(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("migratedb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if termErr := testcontainers.TerminateContainer(pgContainer); termErr != nil {
			t.Logf("failed to terminate container: %s", termErr)
		}
	}()

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := Config{
		Host:     host,
		Port:     port.Int(),
		User:     "testuser",
		Password: "testpass",
		DBName:   "migratedb",
		SSLMode:  "disable",
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	err = db.RunMigrations("../../../migrations")
	require.NoError(t, err)

	// Run migrations again (should handle ErrNoChange)
	err = db.RunMigrations("../../../migrations")
	require.NoError(t, err)

	// Verify tables exist
	tables := []string{
		"ledger_entries",
		"transactions",
		"events",
		"balance_snapshots",
		"deposits",
		"withdrawals",
		"idempotency_keys",
		"event_bus_messages",
	}

	for _, table := range tables {
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", table)
		assert.NoError(t, err)
		assert.True(t, exists, "table %s should exist", table)
	}
}
