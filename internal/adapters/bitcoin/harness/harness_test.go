package harness

import (
	"context"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/adapters/bitcoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBitcoinHarness(t *testing.T) {
	h := NewBitcoinHarness()

	assert.NotNil(t, h)
	assert.Equal(t, int64(700000), h.blockHeight)
	assert.NotNil(t, h.balances)
	assert.NotNil(t, h.utxos)
	assert.NotNil(t, h.transactions)
	assert.NotNil(t, h.mempool)
	assert.Equal(t, big.NewInt(1000), h.feeRate)
}

func TestGetBlockCount(t *testing.T) {
	h := NewBitcoinHarness()

	count, err := h.GetBlockCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(700000), count)
}

func TestGetBalance(t *testing.T) {
	h := NewBitcoinHarness()

	t.Run("address with balance", func(t *testing.T) {
		address := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
		expectedBalance := big.NewInt(100000000)
		h.SetBalance(address, expectedBalance)

		balance, err := h.GetBalance(context.Background(), address)
		require.NoError(t, err)
		assert.Equal(t, expectedBalance, balance)
	})

	t.Run("address without balance", func(t *testing.T) {
		address := "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"

		balance, err := h.GetBalance(context.Background(), address)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(0), balance)
	})
}

func TestListUnspent(t *testing.T) {
	h := NewBitcoinHarness()

	t.Run("address with UTXOs", func(t *testing.T) {
		address := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
		utxo := bitcoin.UTXO{
			TxID:          "abc123",
			Vout:          0,
			Address:       address,
			ScriptPubKey:  "76a914...",
			Amount:        100000000,
			Confirmations: 6,
		}
		h.AddUTXO(address, utxo)

		utxos, err := h.ListUnspent(context.Background(), address)
		require.NoError(t, err)
		assert.Len(t, utxos, 1)
		assert.Equal(t, utxo.TxID, utxos[0].TxID)
		assert.Equal(t, utxo.Amount, utxos[0].Amount)
	})

	t.Run("address without UTXOs", func(t *testing.T) {
		address := "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"

		utxos, err := h.ListUnspent(context.Background(), address)
		require.NoError(t, err)
		assert.Empty(t, utxos)
	})

	t.Run("multiple UTXOs", func(t *testing.T) {
		address := "1MultipleUTXOs"
		utxo1 := bitcoin.UTXO{TxID: "tx1", Amount: 50000000}
		utxo2 := bitcoin.UTXO{TxID: "tx2", Amount: 75000000}

		h.AddUTXO(address, utxo1)
		h.AddUTXO(address, utxo2)

		utxos, err := h.ListUnspent(context.Background(), address)
		require.NoError(t, err)
		assert.Len(t, utxos, 2)
	})
}

func TestGetRawTransaction(t *testing.T) {
	h := NewBitcoinHarness()

	t.Run("transaction in mempool", func(t *testing.T) {
		txHash := "mempool_tx"
		tx := &bitcoin.Transaction{
			TxID:          txHash,
			Hash:          txHash,
			Confirmations: 0,
		}
		h.mempool[txHash] = tx

		result, err := h.GetRawTransaction(context.Background(), txHash)
		require.NoError(t, err)
		assert.Equal(t, txHash, result.TxID)
		assert.Equal(t, int64(0), result.Confirmations)
	})

	t.Run("confirmed transaction", func(t *testing.T) {
		txHash := "confirmed_tx"
		tx := &bitcoin.Transaction{
			TxID:          txHash,
			Hash:          txHash,
			Confirmations: 6,
			BlockHash:     "block_123",
		}
		h.transactions[txHash] = tx

		result, err := h.GetRawTransaction(context.Background(), txHash)
		require.NoError(t, err)
		assert.Equal(t, txHash, result.TxID)
		assert.Equal(t, int64(6), result.Confirmations)
		assert.Equal(t, "block_123", result.BlockHash)
	})

	t.Run("transaction not found", func(t *testing.T) {
		txHash := "nonexistent_tx"

		_, err := h.GetRawTransaction(context.Background(), txHash)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction not found")
	})
}

func TestSendRawTransaction(t *testing.T) {
	h := NewBitcoinHarness()

	rawTx := "0100000001abcdef1234567890"

	txHash, err := h.SendRawTransaction(context.Background(), rawTx)
	require.NoError(t, err)
	assert.NotEmpty(t, txHash)
	assert.Contains(t, txHash, "btc_")

	// Verify transaction is in mempool
	tx, err := h.GetRawTransaction(context.Background(), txHash)
	require.NoError(t, err)
	assert.Equal(t, int64(0), tx.Confirmations)
}

func TestEstimateFee(t *testing.T) {
	h := NewBitcoinHarness()

	t.Run("default fee rate", func(t *testing.T) {
		fee, err := h.EstimateFee(context.Background(), 6)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(1000), fee)
	})

	t.Run("different confirmation blocks", func(t *testing.T) {
		fee1, err := h.EstimateFee(context.Background(), 1)
		require.NoError(t, err)

		fee6, err := h.EstimateFee(context.Background(), 6)
		require.NoError(t, err)

		// Both should return same fee in this simple implementation
		assert.Equal(t, fee1, fee6)
	})
}

func TestMineBlock(t *testing.T) {
	h := NewBitcoinHarness()

	initialHeight := h.blockHeight

	// Add transaction to mempool
	rawTx := "0100000001abcdef1234567890"
	txHash, err := h.SendRawTransaction(context.Background(), rawTx)
	require.NoError(t, err)

	// Verify transaction is in mempool
	tx, err := h.GetRawTransaction(context.Background(), txHash)
	require.NoError(t, err)
	assert.Equal(t, int64(0), tx.Confirmations)

	// Mine a block
	h.MineBlock()

	// Verify block height increased
	newHeight, err := h.GetBlockCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, initialHeight+1, newHeight)

	// Verify transaction is now confirmed
	tx, err = h.GetRawTransaction(context.Background(), txHash)
	require.NoError(t, err)
	assert.Equal(t, int64(1), tx.Confirmations)
	assert.Contains(t, tx.BlockHash, "block_")

	// Verify mempool is empty
	assert.Empty(t, h.mempool)
}

func TestSetBalance(t *testing.T) {
	h := NewBitcoinHarness()

	address := "1TestAddress"
	balance := big.NewInt(500000000)

	h.SetBalance(address, balance)

	result, err := h.GetBalance(context.Background(), address)
	require.NoError(t, err)
	assert.Equal(t, balance, result)
}

func TestAddUTXO(t *testing.T) {
	h := NewBitcoinHarness()

	address := "1TestAddress"
	utxo := bitcoin.UTXO{
		TxID:          "test_tx",
		Vout:          0,
		Amount:        100000000,
		Confirmations: 10,
	}

	h.AddUTXO(address, utxo)

	utxos, err := h.ListUnspent(context.Background(), address)
	require.NoError(t, err)
	assert.Len(t, utxos, 1)
	assert.Equal(t, utxo.TxID, utxos[0].TxID)
}

func TestConcurrentAccess(t *testing.T) {
	h := NewBitcoinHarness()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = h.GetBlockCount(context.Background())
			_, _ = h.GetBalance(context.Background(), "test_addr")
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(idx int) {
			h.SetBalance("addr", big.NewInt(int64(idx)))
			h.MineBlock()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
