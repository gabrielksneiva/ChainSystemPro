package usecases

import (
	"context"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.GetTransactionStatusFunc = func(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error) {
			return entities.TxStatusConfirmed, nil
		}
		uc := NewGetTransactionStatusUseCase(registry, logger)
		out, err := uc.Execute(ctx, GetTransactionStatusInput{ChainID: "evm-mainnet", TransactionHash: "aaaaaaaa"})
		require.NoError(t, err)
		require.Equal(t, "confirmed", out.Status)
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		logger := mocks.NewMockLogger()
		uc := NewGetTransactionStatusUseCase(registry, logger)
		_, err := uc.Execute(ctx, GetTransactionStatusInput{ChainID: "evm-mainnet", TransactionHash: "aaaaaaaa"})
		require.Error(t, err)
	})

	t.Run("adapter error", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.GetTransactionStatusFunc = func(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error) {
			return entities.TxStatusPending, simpleError{"fail"}
		}
		uc := NewGetTransactionStatusUseCase(registry, logger)
		_, err := uc.Execute(ctx, GetTransactionStatusInput{ChainID: "evm-mainnet", TransactionHash: "aaaaaaaa"})
		require.Error(t, err)
	})

	t.Run("invalid hash format", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		uc := NewGetTransactionStatusUseCase(registry, logger)
		_, err := uc.Execute(ctx, GetTransactionStatusInput{ChainID: "evm-mainnet", TransactionHash: ""})
		require.Error(t, err)
	})
}
