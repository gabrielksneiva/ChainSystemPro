package usecases

import (
	"context"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

// simpleError helper for negative test flows
type simpleError struct{ msg string }

func (e simpleError) Error() string { return e.msg }

func TestSignTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fromAddr, _ := valueobjects.NewAddress("0xabc", "evm-mainnet")
	toAddr, _ := valueobjects.NewAddress("0xdef", "evm-mainnet")
	// build minimal valid tx
	tx, _ := entities.NewTransaction(entities.TransactionParams{ChainID: "evm-mainnet", From: fromAddr, To: toAddr})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.SignTransactionFunc = func(ctx context.Context, inTx *entities.Transaction, pk []byte) error { return nil }
		uc := NewSignTransactionUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, SignTransactionInput{ChainID: "evm-mainnet", Transaction: tx, PrivateKey: []byte("privkey")})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, tx.ID(), out.TransactionID)
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewSignTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, SignTransactionInput{ChainID: "evm-mainnet", Transaction: tx, PrivateKey: []byte("privkey")})
		require.Error(t, err)
	})

	t.Run("adapter error", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.SignTransactionFunc = func(ctx context.Context, inTx *entities.Transaction, pk []byte) error {
			return simpleError{"sign failed"}
		}
		uc := NewSignTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, SignTransactionInput{ChainID: "evm-mainnet", Transaction: tx, PrivateKey: []byte("privkey")})
		require.Error(t, err)
	})

	t.Run("nil transaction", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		uc := NewSignTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, SignTransactionInput{ChainID: "evm-mainnet", Transaction: nil, PrivateKey: []byte("privkey")})
		require.Error(t, err)
	})

	t.Run("nil private key", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		uc := NewSignTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, SignTransactionInput{ChainID: "evm-mainnet", Transaction: tx, PrivateKey: nil})
		require.Error(t, err)
	})
}
