package mocks

import (
	"context"
	"math/big"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

type MockChainAdapter struct {
	GetChainIDFunc            func() string
	GetBalanceFunc            func(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error)
	GetTokenBalanceFunc       func(ctx context.Context, chainID string, address *valueobjects.Address, tokenAddress *valueobjects.Address) (*big.Int, error)
	BuildTransactionFunc      func(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error)
	SignTransactionFunc       func(ctx context.Context, tx *entities.Transaction, privateKey []byte) error
	BroadcastTransactionFunc  func(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error)
	EstimateFeeFunc           func(ctx context.Context, tx *entities.Transaction) (*entities.Fee, error)
	GetTransactionStatusFunc  func(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error)
	GetTransactionReceiptFunc func(ctx context.Context, hash *valueobjects.Hash) (*entities.Transaction, error)
}

func (m *MockChainAdapter) GetChainID() string {
	if m.GetChainIDFunc != nil {
		return m.GetChainIDFunc()
	}
	return "mock"
}

func (m *MockChainAdapter) GetChainType() entities.ChainType {
	return entities.ChainTypeEVM
}

func (m *MockChainAdapter) IsConnected(ctx context.Context) bool {
	return true
}

func (m *MockChainAdapter) GetBlockNumber(ctx context.Context) (uint64, error) {
	return 1, nil
}

func (m *MockChainAdapter) GetNativeBalance(ctx context.Context, address *valueobjects.Address) (*big.Int, error) {
	return m.GetBalance(ctx, "mock", address)
}

func (m *MockChainAdapter) GetBalance(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, chainID, address)
	}
	return big.NewInt(0), nil
}

func (m *MockChainAdapter) GetTokenBalance(ctx context.Context, chainID string, address, tokenAddress *valueobjects.Address) (*big.Int, error) {
	if m.GetTokenBalanceFunc != nil {
		return m.GetTokenBalanceFunc(ctx, chainID, address, tokenAddress)
	}
	return big.NewInt(0), nil
}

func (m *MockChainAdapter) BuildTransaction(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error) {
	if m.BuildTransactionFunc != nil {
		return m.BuildTransactionFunc(ctx, params)
	}
	return entities.NewTransaction(params)
}

func (m *MockChainAdapter) EstimateGas(ctx context.Context, tx *entities.Transaction) (uint64, error) {
	return 21000, nil
}

func (m *MockChainAdapter) SetNonce(ctx context.Context, tx *entities.Transaction) error {
	return nil
}

func (m *MockChainAdapter) SignTransaction(ctx context.Context, tx *entities.Transaction, privateKey []byte) error {
	if m.SignTransactionFunc != nil {
		return m.SignTransactionFunc(ctx, tx, privateKey)
	}
	hash, _ := valueobjects.NewHash("0xabcd")
	sig, _ := valueobjects.NewSignature("0x1234")
	tx.SetHash(hash)
	tx.SetSignature(sig)
	return nil
}

func (m *MockChainAdapter) VerifySignature(ctx context.Context, tx *entities.Transaction) (bool, error) {
	return true, nil
}

func (m *MockChainAdapter) BroadcastTransaction(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error) {
	if m.BroadcastTransactionFunc != nil {
		return m.BroadcastTransactionFunc(ctx, tx)
	}
	return valueobjects.NewHash("0xabcd")
}

func (m *MockChainAdapter) GetTransactionStatus(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error) {
	if m.GetTransactionStatusFunc != nil {
		return m.GetTransactionStatusFunc(ctx, hash)
	}
	return entities.TxStatusPending, nil
}

func (m *MockChainAdapter) GetTransactionReceipt(ctx context.Context, hash *valueobjects.Hash) (*entities.Transaction, error) {
	if m.GetTransactionReceiptFunc != nil {
		return m.GetTransactionReceiptFunc(ctx, hash)
	}
	return nil, nil
}

func (m *MockChainAdapter) WaitForConfirmation(ctx context.Context, hash *valueobjects.Hash, confirmations uint64) error {
	return nil
}

func (m *MockChainAdapter) EstimateFee(ctx context.Context, tx *entities.Transaction) (*entities.Fee, error) {
	if m.EstimateFeeFunc != nil {
		return m.EstimateFeeFunc(ctx, tx)
	}
	return entities.NewFee(21000, big.NewInt(20000000000), "ETH")
}

func (m *MockChainAdapter) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(20000000000), nil
}

func (m *MockChainAdapter) GetMaxPriorityFee(ctx context.Context) (*big.Int, error) {
	return big.NewInt(2000000000), nil
}

func (m *MockChainAdapter) GetNetworkInfo(ctx context.Context) (*entities.Network, error) {
	return entities.NewNetwork("mock", "Mock Chain", "http://localhost")
}

func (m *MockChainAdapter) GetPeers(ctx context.Context) (int, error) {
	return 10, nil
}

func (m *MockChainAdapter) GetLatestBlock(ctx context.Context) (uint64, error) {
	return 12345, nil
}
