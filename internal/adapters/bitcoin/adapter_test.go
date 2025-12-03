package bitcoin

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRPCClient is a mock implementation of RPCClient
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) GetBlockCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRPCClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCClient) ListUnspent(ctx context.Context, address string) ([]UTXO, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]UTXO), args.Error(1)
}

func (m *MockRPCClient) GetRawTransaction(ctx context.Context, txHash string) (*Transaction, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockRPCClient) SendRawTransaction(ctx context.Context, rawTx string) (string, error) {
	args := m.Called(ctx, rawTx)
	return args.String(0), args.Error(1)
}

func (m *MockRPCClient) EstimateFee(ctx context.Context, blocks int) (*big.Int, error) {
	args := m.Called(ctx, blocks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func TestNewAdapter(t *testing.T) {
	mockRPC := new(MockRPCClient)
	adapter := NewAdapter(mockRPC, "mainnet")

	assert.NotNil(t, adapter)
	assert.Equal(t, "mainnet", adapter.network)
	assert.Equal(t, mockRPC, adapter.rpcClient)
}

func TestGetChainID(t *testing.T) {
	mockRPC := new(MockRPCClient)
	adapter := NewAdapter(mockRPC, "testnet")

	chainID := adapter.GetChainID()
	assert.Equal(t, "bitcoin-testnet", chainID)
}

func TestGetChainType(t *testing.T) {
	mockRPC := new(MockRPCClient)
	adapter := NewAdapter(mockRPC, "mainnet")

	chainType := adapter.GetChainType()
	assert.Equal(t, entities.ChainTypeBitcoin, chainType)
}

func TestIsConnected(t *testing.T) {
	t.Run("connected successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		mockRPC.On("GetBlockCount", mock.Anything).Return(int64(700000), nil)

		connected := adapter.IsConnected(context.Background())
		assert.True(t, connected)
		mockRPC.AssertExpectations(t)
	})

	t.Run("connection failed", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		mockRPC.On("GetBlockCount", mock.Anything).Return(int64(0), assert.AnError)

		connected := adapter.IsConnected(context.Background())
		assert.False(t, connected)
		mockRPC.AssertExpectations(t)
	})
}

func TestGetBlockNumber(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		mockRPC.On("GetBlockCount", mock.Anything).Return(int64(750000), nil)

		blockNum, err := adapter.GetBlockNumber(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(750000), blockNum)
		mockRPC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		mockRPC.On("GetBlockCount", mock.Anything).Return(int64(0), assert.AnError)

		_, err := adapter.GetBlockNumber(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get block count")
		mockRPC.AssertExpectations(t)
	})
}

func TestGetNativeBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		addr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)

		expectedBalance := big.NewInt(100000000) // 1 BTC
		mockRPC.On("GetBalance", mock.Anything, addr.String()).Return(expectedBalance, nil)

		balance, err := adapter.GetNativeBalance(context.Background(), addr)
		require.NoError(t, err)
		assert.Equal(t, expectedBalance, balance)
		mockRPC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		addr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)

		mockRPC.On("GetBalance", mock.Anything, addr.String()).Return((*big.Int)(nil), assert.AnError)

		_, err = adapter.GetNativeBalance(context.Background(), addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get balance")
		mockRPC.AssertExpectations(t)
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("success with sufficient UTXOs", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		utxos := []UTXO{
			{
				TxID:          "abc123",
				Vout:          0,
				Address:       fromAddr.String(),
				ScriptPubKey:  "76a914...",
				Amount:        200000000, // 2 BTC
				Confirmations: 10,
			},
		}

		mockRPC.On("ListUnspent", mock.Anything, fromAddr.String()).Return(utxos, nil)

		params := entities.TransactionParams{
			ChainID:  "bitcoin-mainnet",
			From:     fromAddr,
			To:       toAddr,
			Value:    big.NewInt(50000000), // 0.5 BTC
			GasPrice: big.NewInt(1000),
		}

		tx, err := adapter.CreateTransaction(context.Background(), params)
		require.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, fromAddr, tx.From())
		assert.Equal(t, toAddr, tx.To())
		assert.Equal(t, big.NewInt(50000000), tx.Value())
		mockRPC.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		utxos := []UTXO{
			{
				TxID:          "abc123",
				Vout:          0,
				Amount:        10000000, // 0.1 BTC
				Confirmations: 10,
			},
		}

		mockRPC.On("ListUnspent", mock.Anything, fromAddr.String()).Return(utxos, nil)

		params := entities.TransactionParams{
			ChainID:  "bitcoin-mainnet",
			From:     fromAddr,
			To:       toAddr,
			Value:    big.NewInt(500000000), // 5 BTC - more than available
			GasPrice: big.NewInt(1000),
		}

		_, err = adapter.CreateTransaction(context.Background(), params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
		mockRPC.AssertExpectations(t)
	})

	t.Run("error listing unspent", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		mockRPC.On("ListUnspent", mock.Anything, fromAddr.String()).Return(([]UTXO)(nil), assert.AnError)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		_, err = adapter.CreateTransaction(context.Background(), params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list unspent")
		mockRPC.AssertExpectations(t)
	})

	t.Run("success with default gas price", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		utxos := []UTXO{
			{
				TxID:          "abc123",
				Vout:          0,
				Amount:        200000000, // 2 BTC
				Confirmations: 10,
			},
		}

		mockRPC.On("ListUnspent", mock.Anything, fromAddr.String()).Return(utxos, nil)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
			// No GasPrice specified
		}

		tx, err := adapter.CreateTransaction(context.Background(), params)
		require.NoError(t, err)
		assert.NotNil(t, tx)
		mockRPC.AssertExpectations(t)
	})
}

func TestSelectCoins(t *testing.T) {
	adapter := &Adapter{}

	t.Run("sufficient funds", func(t *testing.T) {
		utxos := []UTXO{
			{Amount: 100000000}, // 1 BTC
			{Amount: 50000000},  // 0.5 BTC
			{Amount: 25000000},  // 0.25 BTC
		}

		selected, change, err := adapter.selectCoins(utxos, big.NewInt(60000000), big.NewInt(1000))
		require.NoError(t, err)
		assert.NotEmpty(t, selected)
		assert.NotNil(t, change)
		assert.True(t, change.Cmp(big.NewInt(0)) >= 0)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		utxos := []UTXO{
			{Amount: 10000000}, // 0.1 BTC
		}

		_, _, err := adapter.selectCoins(utxos, big.NewInt(500000000), big.NewInt(1000))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("empty UTXOs", func(t *testing.T) {
		utxos := []UTXO{}

		_, _, err := adapter.selectCoins(utxos, big.NewInt(10000000), big.NewInt(1000))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})
}

func TestSignTransaction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		privateKey := []byte("test-private-key")
		err = adapter.SignTransaction(context.Background(), tx, privateKey)
		require.NoError(t, err)
		assert.NotNil(t, tx.Signature())
	})
}

func TestCreateSignature(t *testing.T) {
	mockRPC := new(MockRPCClient)
	adapter := NewAdapter(mockRPC, "mainnet")

	fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
	require.NoError(t, err)
	toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
	require.NoError(t, err)

	params := entities.TransactionParams{
		ChainID: "bitcoin-mainnet",
		From:    fromAddr,
		To:      toAddr,
		Value:   big.NewInt(50000000),
	}

	tx, err := entities.NewTransaction(params)
	require.NoError(t, err)

	privateKey := []byte("test-private-key")
	signature := adapter.createSignature(tx, privateKey)

	assert.NotEmpty(t, signature)
}

func TestBroadcastTransaction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		// Sign the transaction first
		privateKey := []byte("test-private-key")
		err = adapter.SignTransaction(context.Background(), tx, privateKey)
		require.NoError(t, err)

		expectedHash := "abc123def456"
		mockRPC.On("SendRawTransaction", mock.Anything, mock.AnythingOfType("string")).Return(expectedHash, nil)

		hash, err := adapter.BroadcastTransaction(context.Background(), tx)
		require.NoError(t, err)
		assert.NotNil(t, hash)
		assert.Contains(t, hash.String(), expectedHash)
		mockRPC.AssertExpectations(t)
	})

	t.Run("unsigned transaction", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		_, err = adapter.BroadcastTransaction(context.Background(), tx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction not signed")
	})

	t.Run("broadcast error", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		// Sign the transaction
		privateKey := []byte("test-private-key")
		err = adapter.SignTransaction(context.Background(), tx, privateKey)
		require.NoError(t, err)

		mockRPC.On("SendRawTransaction", mock.Anything, mock.AnythingOfType("string")).Return("", assert.AnError)

		_, err = adapter.BroadcastTransaction(context.Background(), tx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to broadcast transaction")
		mockRPC.AssertExpectations(t)
	})
}

func TestBuildRawTransaction(t *testing.T) {
	mockRPC := new(MockRPCClient)
	adapter := NewAdapter(mockRPC, "mainnet")

	fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
	require.NoError(t, err)
	toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
	require.NoError(t, err)

	params := entities.TransactionParams{
		ChainID: "bitcoin-mainnet",
		From:    fromAddr,
		To:      toAddr,
		Value:   big.NewInt(50000000),
	}

	tx, err := entities.NewTransaction(params)
	require.NoError(t, err)

	rawTx := adapter.buildRawTransaction(tx)
	assert.NotEmpty(t, rawTx)

	// Verify it's a valid hex string
	_, err = hex.DecodeString(rawTx)
	assert.NoError(t, err)
}

func TestGetTransactionStatus(t *testing.T) {
	t.Run("confirmed transaction", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		hash, err := valueobjects.NewHash("abc123def456")
		require.NoError(t, err)

		btcTx := &Transaction{
			TxID:          "abc123def456",
			Confirmations: 6,
		}

		mockRPC.On("GetRawTransaction", mock.Anything, hash.String()).Return(btcTx, nil)

		status, err := adapter.GetTransactionStatus(context.Background(), hash)
		require.NoError(t, err)
		assert.Equal(t, entities.TxStatusConfirmed, status)
		mockRPC.AssertExpectations(t)
	})

	t.Run("pending transaction", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		hash, err := valueobjects.NewHash("abc123def456")
		require.NoError(t, err)

		btcTx := &Transaction{
			TxID:          "abc123def456",
			Confirmations: 0,
		}

		mockRPC.On("GetRawTransaction", mock.Anything, hash.String()).Return(btcTx, nil)

		status, err := adapter.GetTransactionStatus(context.Background(), hash)
		require.NoError(t, err)
		assert.Equal(t, entities.TxStatusPending, status)
		mockRPC.AssertExpectations(t)
	})

	t.Run("error fetching transaction", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		hash, err := valueobjects.NewHash("abc123def456")
		require.NoError(t, err)

		mockRPC.On("GetRawTransaction", mock.Anything, hash.String()).Return((*Transaction)(nil), assert.AnError)

		status, err := adapter.GetTransactionStatus(context.Background(), hash)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction")
		assert.Equal(t, entities.TxStatusPending, status)
		mockRPC.AssertExpectations(t)
	})
}

func TestEstimateFee(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		feePerByte := big.NewInt(5000)
		mockRPC.On("EstimateFee", mock.Anything, 6).Return(feePerByte, nil)

		fee, err := adapter.EstimateFee(context.Background(), tx)
		require.NoError(t, err)
		assert.NotNil(t, fee)
		assert.Equal(t, uint64(250), fee.GasLimit())
		assert.Equal(t, feePerByte, fee.GasPrice())
		assert.Equal(t, "BTC", fee.Currency())
		mockRPC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		adapter := NewAdapter(mockRPC, "mainnet")

		fromAddr, err := valueobjects.NewAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "bitcoin-mainnet")
		require.NoError(t, err)
		toAddr, err := valueobjects.NewAddress("1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "bitcoin-mainnet")
		require.NoError(t, err)

		params := entities.TransactionParams{
			ChainID: "bitcoin-mainnet",
			From:    fromAddr,
			To:      toAddr,
			Value:   big.NewInt(50000000),
		}

		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)

		mockRPC.On("EstimateFee", mock.Anything, 6).Return((*big.Int)(nil), assert.AnError)

		_, err = adapter.EstimateFee(context.Background(), tx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to estimate fee")
		mockRPC.AssertExpectations(t)
	})
}
