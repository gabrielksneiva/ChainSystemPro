package ledger

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// EntryType represents the type of ledger entry
type EntryType string

const (
	EntryTypeDeposit    EntryType = "deposit"
	EntryTypeWithdrawal EntryType = "withdrawal"
	EntryTypeTransfer   EntryType = "transfer"
	EntryTypeMint       EntryType = "mint"
	EntryTypeBurn       EntryType = "burn"
	EntryTypeFee        EntryType = "fee"
)

// Entry represents a ledger entry
type Entry struct {
	ID           uuid.UUID      `db:"id"`
	EntryType    EntryType      `db:"entry_type"`
	ChainID      string         `db:"chain_id"`
	Address      string         `db:"address"`
	Amount       string         `db:"amount"` // Stored as string to handle big.Int
	Asset        string         `db:"asset"`
	TxHash       sql.NullString `db:"tx_hash"`
	EventID      uuid.UUID      `db:"event_id"`
	Metadata     JSONBMap       `db:"metadata"`
	BalanceAfter sql.NullString `db:"balance_after"`
	CreatedAt    time.Time      `db:"created_at"`
}

// JSONBMap handles JSON marshaling for PostgreSQL JSONB type
type JSONBMap map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, j)
}

// Repository defines ledger repository interface
type Repository interface {
	Create(ctx context.Context, entry *Entry) error
	CreateWithTx(ctx context.Context, tx *sqlx.Tx, entry *Entry) error
	GetByID(ctx context.Context, id uuid.UUID) (*Entry, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) (*Entry, error)
	ListByAddress(ctx context.Context, chainID, address string, limit, offset int) ([]*Entry, error)
	ListByTxHash(ctx context.Context, txHash string) ([]*Entry, error)
	GetBalance(ctx context.Context, chainID, address, asset string) (*big.Int, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new ledger repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const insertLedgerEntryQuery = `
	INSERT INTO ledger_entries (
		id, entry_type, chain_id, address, amount, asset, 
		tx_hash, event_id, metadata, balance_after, created_at
	) VALUES (
		:id, :entry_type, :chain_id, :address, :amount, :asset,
		:tx_hash, :event_id, :metadata, :balance_after, :created_at
	)
`

// Create creates a new ledger entry
func (r *repository) Create(ctx context.Context, entry *Entry) error {
	query := insertLedgerEntryQuery

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.db.NamedExecContext(ctx, query, entry)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	return nil
}

// CreateWithTx creates a new ledger entry within a transaction
func (r *repository) CreateWithTx(ctx context.Context, tx *sqlx.Tx, entry *Entry) error {
	query := `
		INSERT INTO ledger_entries (
			id, entry_type, chain_id, address, amount, asset, 
			tx_hash, event_id, metadata, balance_after, created_at
		) VALUES (
			:id, :entry_type, :chain_id, :address, :amount, :asset,
			:tx_hash, :event_id, :metadata, :balance_after, :created_at
		)
	`

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := tx.NamedExecContext(ctx, query, entry)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	return nil
}

// GetByID retrieves a ledger entry by ID
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Entry, error) {
	var entry Entry
	query := `
		SELECT id, entry_type, chain_id, address, amount, asset,
		       tx_hash, event_id, metadata, balance_after, created_at
		FROM ledger_entries
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &entry, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ledger entry not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get ledger entry: %w", err)
	}

	return &entry, nil
}

// GetByEventID retrieves a ledger entry by event ID
func (r *repository) GetByEventID(ctx context.Context, eventID uuid.UUID) (*Entry, error) {
	var entry Entry
	query := `
		SELECT id, entry_type, chain_id, address, amount, asset,
		       tx_hash, event_id, metadata, balance_after, created_at
		FROM ledger_entries
		WHERE event_id = $1
	`

	err := r.db.GetContext(ctx, &entry, query, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ledger entry not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get ledger entry: %w", err)
	}

	return &entry, nil
}

// ListByAddress lists ledger entries for an address
func (r *repository) ListByAddress(ctx context.Context, chainID, address string, limit, offset int) ([]*Entry, error) {
	var entries []*Entry
	query := `
		SELECT id, entry_type, chain_id, address, amount, asset,
		       tx_hash, event_id, metadata, balance_after, created_at
		FROM ledger_entries
		WHERE chain_id = $1 AND address = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &entries, query, chainID, address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list ledger entries: %w", err)
	}

	return entries, nil
}

// ListByTxHash lists ledger entries for a transaction hash
func (r *repository) ListByTxHash(ctx context.Context, txHash string) ([]*Entry, error) {
	var entries []*Entry
	query := `
		SELECT id, entry_type, chain_id, address, amount, asset,
		       tx_hash, event_id, metadata, balance_after, created_at
		FROM ledger_entries
		WHERE tx_hash = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &entries, query, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to list ledger entries: %w", err)
	}

	return entries, nil
}

// GetBalance calculates the current balance for an address
func (r *repository) GetBalance(ctx context.Context, chainID, address, asset string) (*big.Int, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN entry_type IN ('deposit', 'transfer', 'mint') THEN amount::numeric
				WHEN entry_type IN ('withdrawal', 'burn', 'fee') THEN -amount::numeric
				ELSE 0
			END
		), 0) as balance
		FROM ledger_entries
		WHERE chain_id = $1 AND address = $2 AND asset = $3
	`

	var balanceStr string
	err := r.db.GetContext(ctx, &balanceStr, query, chainID, address, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	balance := new(big.Int)
	balance, ok := balance.SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance: %s", balanceStr)
	}

	return balance, nil
}
