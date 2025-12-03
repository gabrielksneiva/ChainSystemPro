package entities

import (
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/require"
)

func TestTransactionGetters(t *testing.T) {
	t.Parallel()

	from, _ := valueobjects.NewAddress("0xfrom", "chain")
	to, _ := valueobjects.NewAddress("0xto", "chain")
	tx, _ := NewTransaction(TransactionParams{
		ChainID:  "chain",
		From:     from,
		To:       to,
		Value:    big.NewInt(100),
		Data:     []byte("data"),
		GasLimit: 21000,
		GasPrice: big.NewInt(20),
	})

	require.Equal(t, from, tx.From())
	require.Equal(t, to, tx.To())
	require.Equal(t, big.NewInt(100), tx.Value())
	require.Equal(t, []byte("data"), tx.Data())
	require.Nil(t, tx.Nonce())
	require.Equal(t, uint64(21000), tx.GasLimit())
	require.Equal(t, big.NewInt(20), tx.GasPrice())
	require.Nil(t, tx.MaxFeePerGas())
	require.Nil(t, tx.MaxPriorityFee())
	require.NotZero(t, tx.CreatedAt())
}

func TestFeeGetters(t *testing.T) {
	t.Parallel()

	fee, _ := NewFee(21000, big.NewInt(20), "ETH")
	require.Equal(t, uint64(21000), fee.GasLimit())
	require.Equal(t, big.NewInt(20), fee.GasPrice())
	require.Nil(t, fee.MaxFeePerGas())
	require.Nil(t, fee.MaxPriorityFee())
	require.NotNil(t, fee.Total())
	require.Equal(t, "ETH", fee.Currency())

	eip1559Fee, _ := NewEIP1559Fee(21000, big.NewInt(30), big.NewInt(2), "ETH")
	require.Equal(t, big.NewInt(30), eip1559Fee.MaxFeePerGas())
	require.Equal(t, big.NewInt(2), eip1559Fee.MaxPriorityFee())
}

func TestWalletGetters(t *testing.T) {
	t.Parallel()

	addr, _ := valueobjects.NewAddress("0xwallet", "chain")
	wallet, _ := NewWallet(addr, "chain", "my-wallet")
	require.NotZero(t, wallet.CreatedAt())
	require.NotZero(t, wallet.UpdatedAt())
}

func TestNetworkGetters(t *testing.T) {
	t.Parallel()

	net, _ := NewNetwork("chain", "Test Network", "http://localhost")
	require.NotZero(t, net.CreatedAt())
	require.NotZero(t, net.UpdatedAt())
}
