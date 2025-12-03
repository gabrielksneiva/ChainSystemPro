package events

import (
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseEvent(t *testing.T) {
	event := NewBaseEvent(EventTypeTransactionCreated, "ethereum")

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionCreated, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewTransactionCreatedEvent(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	nonce := valueobjects.NewNonce(1)

	tx, err := entities.NewTransaction(entities.TransactionParams{
		ChainID:  "ethereum",
		From:     from,
		To:       to,
		Value:    big.NewInt(1000),
		Nonce:    nonce,
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000),
	})
	require.NoError(t, err)

	event := NewTransactionCreatedEvent(tx)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionCreated, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.Equal(t, tx.ID(), event.TransactionID)
	assert.Equal(t, from.String(), event.From)
	assert.Equal(t, to.String(), event.To)
	assert.Equal(t, "1000", event.Value)
	assert.Equal(t, uint64(1), event.Nonce)
	assert.Equal(t, uint64(21000), event.GasLimit)
	assert.Equal(t, "20000000000", event.GasPrice)
}

func TestNewTransactionSignedEvent(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")

	tx, err := entities.NewTransaction(entities.TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})
	require.NoError(t, err)

	hash, _ := valueobjects.NewHash("0xabcd")
	sig, _ := valueobjects.NewSignature("0x1234")
	_ = tx.SetHash(hash)
	_ = tx.SetSignature(sig)

	event := NewTransactionSignedEvent(tx)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionSigned, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.Equal(t, tx.ID(), event.TransactionID)
	assert.Equal(t, hash.Hex(), event.Hash)
	assert.Equal(t, sig.Hex(), event.Signature)
}

func TestNewTransactionBroadcastedEvent(t *testing.T) {
	hash, _ := valueobjects.NewHash("0xabcdef")

	event := NewTransactionBroadcastedEvent("ethereum", "tx-123", hash)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionBroadcasted, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.Equal(t, "tx-123", event.TransactionID)
	assert.Equal(t, hash.Hex(), event.Hash)
}

func TestNewTransactionConfirmedEvent(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")

	tx, err := entities.NewTransaction(entities.TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})
	require.NoError(t, err)

	hash, _ := valueobjects.NewHash("0xabcd")
	_ = tx.SetHash(hash)
	tx.SetBlockNumber(12345)
	tx.SetConfirmations(10)

	event := NewTransactionConfirmedEvent(tx)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionConfirmed, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.Equal(t, tx.ID(), event.TransactionID)
	assert.Equal(t, hash.Hex(), event.Hash)
	assert.Equal(t, uint64(12345), event.BlockNumber)
	assert.Equal(t, uint64(10), event.Confirmations)
}

func TestNewTransactionFailedEvent(t *testing.T) {
	event := NewTransactionFailedEvent("ethereum", "tx-123", "0xabcdef", "out of gas", "E001")

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeTransactionFailed, event.Type)
	assert.Equal(t, "ethereum", event.ChainID)
	assert.Equal(t, "tx-123", event.TransactionID)
	assert.Equal(t, "0xabcdef", event.Hash)
	assert.Equal(t, "out of gas", event.Reason)
	assert.Equal(t, "E001", event.ErrorCode)
}

func TestNewBalanceQueriedEvent(t *testing.T) {
	addr, _ := valueobjects.NewAddress("0xwallet", "ethereum")
	balance := big.NewInt(1000000000000000000)

	t.Run("without token address", func(t *testing.T) {
		event := NewBalanceQueriedEvent("ethereum", addr, balance, nil)

		assert.NotEmpty(t, event.ID)
		assert.Equal(t, EventTypeBalanceQueried, event.Type)
		assert.Equal(t, "ethereum", event.ChainID)
		assert.Equal(t, addr.String(), event.Address)
		assert.Equal(t, balance.String(), event.Balance)
		assert.Empty(t, event.TokenAddress)
	})

	t.Run("with token address", func(t *testing.T) {
		tokenAddr, _ := valueobjects.NewAddress("0xtoken", "ethereum")
		event := NewBalanceQueriedEvent("ethereum", addr, balance, tokenAddr)

		assert.Equal(t, tokenAddr.String(), event.TokenAddress)
	})
}

func TestNewFeeEstimatedEvent(t *testing.T) {
	t.Run("with standard fee", func(t *testing.T) {
		fee, err := entities.NewFee(21000, big.NewInt(20000000000), "ETH")
		require.NoError(t, err)

		event := NewFeeEstimatedEvent("ethereum", fee, "tx-123")

		assert.NotEmpty(t, event.ID)
		assert.Equal(t, EventTypeFeeEstimated, event.Type)
		assert.Equal(t, "ethereum", event.ChainID)
		assert.Equal(t, "tx-123", event.TransactionID)
		assert.Equal(t, uint64(21000), event.GasLimit)
		assert.Equal(t, "20000000000", event.GasPrice)
		assert.NotEmpty(t, event.Total)
		assert.Equal(t, "ETH", event.Currency)
	})

	t.Run("with EIP-1559 fee", func(t *testing.T) {
		fee, err := entities.NewEIP1559Fee(21000, big.NewInt(30000000000), big.NewInt(2000000000), "ETH")
		require.NoError(t, err)

		event := NewFeeEstimatedEvent("ethereum", fee, "tx-123")

		assert.Equal(t, "30000000000", event.MaxFeePerGas)
		assert.Equal(t, "2000000000", event.MaxPriorityFee)
	})
}
