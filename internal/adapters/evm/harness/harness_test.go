package harness

import (
	"context"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/require"
)

func TestHarnessBasicFlows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewEVMHarness("evm-mainnet")

	// balances
	addr, _ := valueobjects.NewAddress("0xabc", "evm-mainnet")
	h.SetBalance(addr.String(), big.NewInt(100))
	bal, err := h.GetBalance(ctx, "evm-mainnet", addr)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(100), bal)

	// build + sign + broadcast
	to, _ := valueobjects.NewAddress("0xdef", "evm-mainnet")
	tx, err := h.BuildTransaction(ctx, entities.TransactionParams{ChainID: "evm-mainnet", From: addr, To: to, Value: big.NewInt(1)})
	require.NoError(t, err)
	require.NotNil(t, tx)

	err = h.SignTransaction(ctx, tx, []byte("pk"))
	require.NoError(t, err)
	require.NotNil(t, tx.Hash())

	hash, err := h.BroadcastTransaction(ctx, tx)
	require.NoError(t, err)
	require.NotEmpty(t, hash.Hex())

	// status and receipt
	status, err := h.GetTransactionStatus(ctx, hash)
	require.NoError(t, err)
	require.NotEmpty(t, status)

	receipt, err := h.GetTransactionReceipt(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, receipt)

	// fee estimate
	fee, err := h.EstimateFee(ctx, tx)
	require.NoError(t, err)
	require.NotNil(t, fee)
}

func TestHarnessAllMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewEVMHarness("evm-test")

	// chain info
	require.Equal(t, "evm-test", h.GetChainID())
	require.Equal(t, entities.ChainTypeEVM, h.GetChainType())
	require.True(t, h.IsConnected(ctx))

	// blocks
	bn, err := h.GetBlockNumber(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), bn)
	latest, err := h.GetLatestBlock(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), latest)

	// gas pricing
	gp, err := h.GetGasPrice(ctx)
	require.NoError(t, err)
	require.NotNil(t, gp)
	pf, err := h.GetMaxPriorityFee(ctx)
	require.NoError(t, err)
	require.NotNil(t, pf)

	// network info
	ni, err := h.GetNetworkInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, ni)
	peers, err := h.GetPeers(ctx)
	require.NoError(t, err)
	require.Greater(t, peers, 0)

	// balances with native and token
	addr, _ := valueobjects.NewAddress("0x123", "evm-test")
	h.SetBalance(addr.String(), big.NewInt(500))
	nb, err := h.GetNativeBalance(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(500), nb)

	tokenAddr, _ := valueobjects.NewAddress("0xtoken", "evm-test")
	h.SetTokenBalance(addr.String(), tokenAddr.String(), big.NewInt(999))
	tb, err := h.GetTokenBalance(ctx, "evm-test", addr, tokenAddr)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(999), tb)

	// transaction flow with all steps
	to, _ := valueobjects.NewAddress("0xto", "evm-test")
	tx, _ := h.BuildTransaction(ctx, entities.TransactionParams{ChainID: "evm-test", From: addr, To: to, Value: big.NewInt(10)})

	gas, err := h.EstimateGas(ctx, tx)
	require.NoError(t, err)
	require.Greater(t, gas, uint64(0))

	err = h.SetNonce(ctx, tx)
	require.NoError(t, err)

	err = h.SignTransaction(ctx, tx, []byte("key"))
	require.NoError(t, err)

	valid, err := h.VerifySignature(ctx, tx)
	require.NoError(t, err)
	require.True(t, valid)

	hash, _ := h.BroadcastTransaction(ctx, tx)
	err = h.WaitForConfirmation(ctx, hash, 1)
	require.NoError(t, err)
}
