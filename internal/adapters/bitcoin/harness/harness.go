package harness

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/gabrielksneiva/ChainSystemPro/internal/adapters/bitcoin"
)

// BitcoinHarness simulates a Bitcoin node for testing
type BitcoinHarness struct {
	mu           sync.RWMutex
	blockHeight  int64
	balances     map[string]*big.Int
	utxos        map[string][]bitcoin.UTXO
	transactions map[string]*bitcoin.Transaction
	mempool      map[string]*bitcoin.Transaction
	feeRate      *big.Int
}

// NewBitcoinHarness creates a new Bitcoin test harness
func NewBitcoinHarness() *BitcoinHarness {
	return &BitcoinHarness{
		blockHeight:  700000,
		balances:     make(map[string]*big.Int),
		utxos:        make(map[string][]bitcoin.UTXO),
		transactions: make(map[string]*bitcoin.Transaction),
		mempool:      make(map[string]*bitcoin.Transaction),
		feeRate:      big.NewInt(1000), // 1000 sats/byte
	}
}

// GetBlockCount returns the current block height
func (h *BitcoinHarness) GetBlockCount(ctx context.Context) (int64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.blockHeight, nil
}

// GetBalance returns the balance for an address
func (h *BitcoinHarness) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	balance, exists := h.balances[address]
	if !exists {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(balance), nil
}

// ListUnspent returns unspent outputs for an address
func (h *BitcoinHarness) ListUnspent(ctx context.Context, address string) ([]bitcoin.UTXO, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	utxos, exists := h.utxos[address]
	if !exists {
		return []bitcoin.UTXO{}, nil
	}

	result := make([]bitcoin.UTXO, len(utxos))
	copy(result, utxos)
	return result, nil
}

// GetRawTransaction returns a transaction by hash
func (h *BitcoinHarness) GetRawTransaction(ctx context.Context, txHash string) (*bitcoin.Transaction, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check mempool first
	if tx, exists := h.mempool[txHash]; exists {
		return tx, nil
	}

	// Check confirmed transactions
	tx, exists := h.transactions[txHash]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txHash)
	}

	return tx, nil
}

// SendRawTransaction broadcasts a transaction
func (h *BitcoinHarness) SendRawTransaction(ctx context.Context, rawTx string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Generate transaction hash
	txHash := fmt.Sprintf("btc_%s", rawTx[:16])

	// Create mock transaction
	tx := &bitcoin.Transaction{
		TxID:          txHash,
		Hash:          txHash,
		Confirmations: 0,
	}

	h.mempool[txHash] = tx
	return txHash, nil
}

// EstimateFee estimates fee for confirmation in N blocks
func (h *BitcoinHarness) EstimateFee(ctx context.Context, blocks int) (*big.Int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return new(big.Int).Set(h.feeRate), nil
}

// MineBlock simulates mining a block
func (h *BitcoinHarness) MineBlock() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.blockHeight++

	// Move transactions from mempool to confirmed
	for hash, tx := range h.mempool {
		tx.Confirmations = 1
		tx.BlockHash = fmt.Sprintf("block_%d", h.blockHeight)
		h.transactions[hash] = tx
		delete(h.mempool, hash)
	}
}

// SetBalance sets the balance for an address (test helper)
func (h *BitcoinHarness) SetBalance(address string, balance *big.Int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.balances[address] = new(big.Int).Set(balance)
}

// AddUTXO adds a UTXO for an address (test helper)
func (h *BitcoinHarness) AddUTXO(address string, utxo bitcoin.UTXO) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.utxos[address] = append(h.utxos[address], utxo)
}
