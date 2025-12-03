package ledger

import (
	"context"
	"database/sql"
	"math/big"
	"testing"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testChainID = "evm-1"

func setupTestDB(t *testing.T) (db *database.DB, cleanup func()) {
	var err error
	ctx := context.Background()

	var pgContainer *postgres.PostgresContainer
	pgContainer, err = postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ledgertest"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := database.Config{
		Host:     host,
		Port:     port.Int(),
		User:     "testuser",
		Password: "testpass",
		DBName:   "ledgertest",
		SSLMode:  "disable",
	}

	db, err = database.New(cfg)
	require.NoError(t, err)

	// Run migrations
	err = db.RunMigrations("../../../migrations")
	require.NoError(t, err)

	cleanup = func() {
		db.Close()
		if termErr := testcontainers.TerminateContainer(pgContainer); termErr != nil {
			t.Logf("failed to terminate container: %s", termErr)
		}
	}

	return db, cleanup
}

func TestLedgerRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	eventID := uuid.New()
	entry := &Entry{
		EntryType:    EntryTypeDeposit,
		ChainID:      testChainID,
		Address:      "0x1234567890abcdef",
		Amount:       "1000000000000000000", // 1 ETH in wei
		Asset:        "ETH",
		TxHash:       sql.NullString{String: "0xabcdef", Valid: true},
		EventID:      eventID,
		Metadata:     map[string]interface{}{"source": "test"},
		BalanceAfter: sql.NullString{String: "1000000000000000000", Valid: true},
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, entry.ID)
	assert.False(t, entry.CreatedAt.IsZero())
}

func TestLedgerRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	// Create entry
	eventID := uuid.New()
	entry := &Entry{
		EntryType: EntryTypeTransfer,
		ChainID:   testChainID,
		Address:   "0xabcdef",
		Amount:    "500000000000000000",
		Asset:     "ETH",
		EventID:   eventID,
		Metadata:  map[string]interface{}{},
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, retrieved.ID)
	assert.Equal(t, entry.EntryType, retrieved.EntryType)
	assert.Equal(t, entry.ChainID, retrieved.ChainID)
	assert.Equal(t, entry.Address, retrieved.Address)
	assert.Equal(t, entry.Amount, retrieved.Amount)
}

func TestLedgerRepository_GetByEventID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	eventID := uuid.New()
	entry := &Entry{
		EntryType: EntryTypeMint,
		ChainID:   testChainID,
		Address:   "0xminter",
		Amount:    "1000000",
		Asset:     "USDT",
		EventID:   eventID,
		Metadata:  map[string]interface{}{},
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// Get by event ID
	retrieved, err := repo.GetByEventID(ctx, eventID)
	require.NoError(t, err)
	assert.Equal(t, eventID, retrieved.EventID)
}

func TestLedgerRepository_ListByAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	chainID := testChainID
	address := "0xtest"

	// Create multiple entries
	for i := 0; i < 5; i++ {
		entry := &Entry{
			EntryType: EntryTypeDeposit,
			ChainID:   chainID,
			Address:   address,
			Amount:    "1000000",
			Asset:     "ETH",
			EventID:   uuid.New(),
			Metadata:  map[string]interface{}{},
		}
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// List entries
	entries, err := repo.ListByAddress(ctx, chainID, address, 10, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 5)
}

func TestLedgerRepository_GetBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	chainID := testChainID
	address := "0xbalance"
	asset := "ETH"

	// Create deposits
	deposit1 := &Entry{
		EntryType: EntryTypeDeposit,
		ChainID:   chainID,
		Address:   address,
		Amount:    "1000000000000000000", // 1 ETH
		Asset:     asset,
		EventID:   uuid.New(),
		Metadata:  map[string]interface{}{},
	}
	err := repo.Create(ctx, deposit1)
	require.NoError(t, err)

	deposit2 := &Entry{
		EntryType: EntryTypeDeposit,
		ChainID:   chainID,
		Address:   address,
		Amount:    "2000000000000000000", // 2 ETH
		Asset:     asset,
		EventID:   uuid.New(),
		Metadata:  map[string]interface{}{},
	}
	err = repo.Create(ctx, deposit2)
	require.NoError(t, err)

	// Create withdrawal
	withdrawal := &Entry{
		EntryType: EntryTypeWithdrawal,
		ChainID:   chainID,
		Address:   address,
		Amount:    "500000000000000000", // 0.5 ETH
		Asset:     asset,
		EventID:   uuid.New(),
		Metadata:  map[string]interface{}{},
	}
	err = repo.Create(ctx, withdrawal)
	require.NoError(t, err)

	// Get balance
	balance, err := repo.GetBalance(ctx, chainID, address, asset)
	require.NoError(t, err)

	expected := new(big.Int)
	expected.SetString("2500000000000000000", 10) // 2.5 ETH
	assert.Equal(t, expected.String(), balance.String())
}

func TestLedgerRepository_CreateWithTx(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		entry := &Entry{
			EntryType: EntryTypeDeposit,
			ChainID:   testChainID,
			Address:   "0xtxtest",
			Amount:    "1000000",
			Asset:     "ETH",
			EventID:   uuid.New(),
			Metadata:  map[string]interface{}{},
		}
		return repo.CreateWithTx(ctx, tx, entry)
	})
	require.NoError(t, err)

	// Verify entry was created
	entries, err := repo.ListByAddress(ctx, testChainID, "0xtxtest", 10, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestLedgerRepository_ListByTxHash(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db.DB)
	ctx := context.Background()

	txHash := "0xabcdef1234567890"

	// Create multiple entries with the same tx hash
	for i := 0; i < 3; i++ {
		entry := &Entry{
			EntryType: EntryTypeTransfer,
			ChainID:   testChainID,
			Address:   "0xaddr",
			Amount:    "100000",
			Asset:     "ETH",
			TxHash:    sql.NullString{String: txHash, Valid: true},
			EventID:   uuid.New(),
			Metadata:  map[string]interface{}{},
		}
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Create an entry with a different tx hash
	otherEntry := &Entry{
		EntryType: EntryTypeDeposit,
		ChainID:   testChainID,
		Address:   "0xother",
		Amount:    "200000",
		Asset:     "ETH",
		TxHash:    sql.NullString{String: "0xdifferent", Valid: true},
		EventID:   uuid.New(),
		Metadata:  map[string]interface{}{},
	}
	err := repo.Create(ctx, otherEntry)
	require.NoError(t, err)

	// List by tx hash
	entries, err := repo.ListByTxHash(ctx, txHash)
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	// Verify all entries have the correct tx hash
	for _, entry := range entries {
		assert.True(t, entry.TxHash.Valid)
		assert.Equal(t, txHash, entry.TxHash.String)
	}

	// Test with non-existent tx hash
	emptyEntries, err := repo.ListByTxHash(ctx, "0xnonexistent")
	require.NoError(t, err)
	assert.Empty(t, emptyEntries)
}

func TestJSONBMap_Value(t *testing.T) {
	// Test with valid map
	jmap := JSONBMap{"key": "value", "number": float64(123)}
	val, err := jmap.Value()
	require.NoError(t, err)
	assert.NotNil(t, val)

	// Test with nil map
	var nilMap JSONBMap
	val, err = nilMap.Value()
	require.NoError(t, err)
	assert.NotNil(t, val)

	// The result should be valid JSON
	bytes, ok := val.([]byte)
	assert.True(t, ok)
	assert.NotEmpty(t, bytes)
}

func TestJSONBMap_Scan(t *testing.T) {
	// Test with valid JSON bytes
	jsonData := []byte(`{"key":"value","number":123}`)
	var jmap JSONBMap
	err := jmap.Scan(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "value", jmap["key"])
	assert.Equal(t, float64(123), jmap["number"])

	// Test with nil
	var jmapNil JSONBMap
	err = jmapNil.Scan(nil)
	require.NoError(t, err)
	assert.NotNil(t, jmapNil)
	assert.Empty(t, jmapNil)

	// Test with invalid type
	var jmapInvalid JSONBMap
	err = jmapInvalid.Scan("invalid string")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan JSONB")

	// Test with invalid JSON
	var jmapBadJSON JSONBMap
	err = jmapBadJSON.Scan([]byte(`{invalid json`))
	require.Error(t, err)
}
