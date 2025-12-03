package mocks

import (
	"context"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestMockChainAdapter_AllMethods(t *testing.T) {
	t.Parallel()
	m := &MockChainAdapter{}
	// Ensure deterministic mock behavior for balance-related methods
	m.GetBalanceFunc = func(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
		return big.NewInt(123), nil
	}
	m.GetTokenBalanceFunc = func(ctx context.Context, chainID string, address, tokenAddress *valueobjects.Address) (*big.Int, error) {
		return big.NewInt(456), nil
	}
	ctx := context.Background()
	addr, _ := valueobjects.NewAddress("0x1234", "eth")
	from, _ := valueobjects.NewAddress("0x1234", "mock")
	to, _ := valueobjects.NewAddress("0x5678", "mock")
	value := big.NewInt(1)
	txParams := entities.TransactionParams{
		From:    from,
		To:      to,
		Value:   value,
		ChainID: "mock",
	}
	tx, _ := entities.NewTransaction(txParams)
	hash, _ := valueobjects.NewHash("0x1234")

	assert.Equal(t, "mock", m.GetChainID())
	assert.Equal(t, entities.ChainTypeEVM, m.GetChainType())
	assert.True(t, m.IsConnected(ctx))
	bn, err := m.GetBlockNumber(ctx)
	assert.Equal(t, uint64(1), bn)
	assert.NoError(t, err)
	bal, err := m.GetNativeBalance(ctx, addr)
	assert.NotNil(t, bal)
	assert.NoError(t, err)
	bal2, err := m.GetBalance(ctx, "mock", addr)
	assert.NotNil(t, bal2)
	assert.NoError(t, err)
	tbal, err := m.GetTokenBalance(ctx, "mock", addr, addr)
	assert.NotNil(t, tbal)
	assert.NoError(t, err)
	tx2, _ := m.BuildTransaction(ctx, txParams)
	assert.NotNil(t, tx2)
	gas, _ := m.EstimateGas(ctx, tx)
	assert.Equal(t, uint64(21000), gas)
	_ = m.SetNonce(ctx, tx)
	_ = m.SignTransaction(ctx, tx, []byte("key"))
	ok, _ := m.VerifySignature(ctx, tx)
	assert.True(t, ok)
	hash2, _ := m.BroadcastTransaction(ctx, tx)
	assert.NotNil(t, hash2)
	status, _ := m.GetTransactionStatus(ctx, hash)
	assert.Equal(t, entities.TxStatusPending, status)
	tx3, _ := m.GetTransactionReceipt(ctx, hash)
	assert.Nil(t, tx3)
	_ = m.WaitForConfirmation(ctx, hash, 1)
	fee, _ := m.EstimateFee(ctx, tx)
	assert.NotNil(t, fee)
	gp, _ := m.GetGasPrice(ctx)
	assert.NotNil(t, gp)
	mpf, _ := m.GetMaxPriorityFee(ctx)
	assert.NotNil(t, mpf)
	net, _ := m.GetNetworkInfo(ctx)
	assert.NotNil(t, net)
	peers, _ := m.GetPeers(ctx)
	assert.Equal(t, 10, peers)
	lb, _ := m.GetLatestBlock(ctx)
	assert.Equal(t, uint64(12345), lb)
}
