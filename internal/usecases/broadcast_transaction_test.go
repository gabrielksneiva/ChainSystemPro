package usecases

import (
	"context"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

func TestBroadcastTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	from, _ := valueobjects.NewAddress("0xabc", "evm-mainnet")
	to, _ := valueobjects.NewAddress("0xdef", "evm-mainnet")
	tx, _ := entities.NewTransaction(entities.TransactionParams{ChainID: "evm-mainnet", From: from, To: to})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.BroadcastTransactionFunc = func(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error) {
			return valueobjects.NewHash("aaaaaaaa")
		}
		uc := NewBroadcastTransactionUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, BroadcastTransactionInput{ChainID: "evm-mainnet", Transaction: tx})
		require.NoError(t, err)
		require.Equal(t, "0xaaaaaaaa", out.Hash)
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewBroadcastTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, BroadcastTransactionInput{ChainID: "evm-mainnet", Transaction: tx})
		require.Error(t, err)
	})

	t.Run("adapter error", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.BroadcastTransactionFunc = func(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error) {
			return nil, simpleError{"broadcast failed"}
		}
		uc := NewBroadcastTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, BroadcastTransactionInput{ChainID: "evm-mainnet", Transaction: tx})
		require.Error(t, err)
	})

	t.Run("nil transaction", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		uc := NewBroadcastTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, BroadcastTransactionInput{ChainID: "evm-mainnet", Transaction: nil})
		require.Error(t, err)
	})
}
