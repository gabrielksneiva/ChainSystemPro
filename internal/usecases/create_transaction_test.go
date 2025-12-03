package usecases

import (
	"context"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()

		require.NoError(t, registry.Register("evm-mainnet", adapter))

		adapter.BuildTransactionFunc = func(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error) {
			return entities.NewTransaction(params)
		}

		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID:  "evm-mainnet",
			From:     "0xabc",
			To:       "0xdef",
			Value:    "1000000000000000000",
			GasLimit: 21000,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, "evm-mainnet", out.ChainID)
		require.NotEmpty(t, out.TransactionID)
	})

	t.Run("validation error: missing chainID", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID: "",
			From:    "0xabc",
			To:      "0xdef",
			Value:   "1",
		})
		require.Error(t, err)
	})

	t.Run("registry error: chain not found", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID: "unknown",
			From:    "0xabc",
			To:      "0xdef",
			Value:   "1",
		})
		require.Error(t, err)
	})

	t.Run("validation error: missing from", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID: "evm-mainnet",
			From:    "",
			To:      "0xdef",
			Value:   "1",
		})
		require.Error(t, err)
	})

	t.Run("validation error: missing to", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID: "evm-mainnet",
			From:    "0xabc",
			To:      "",
			Value:   "1",
		})
		require.Error(t, err)
	})

	t.Run("invalid value format", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID: "evm-mainnet",
			From:    "0xabc",
			To:      "0xdef",
			Value:   "invalid",
		})
		require.Error(t, err)
	})

	t.Run("adapter error on build", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		require.NoError(t, registry.Register("evm-mainnet", adapter))

		adapter.BuildTransactionFunc = func(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error) {
			return nil, simpleError{"build failed"}
		}

		uc := NewCreateTransactionUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, CreateTransactionInput{
			ChainID:  "evm-mainnet",
			From:     "0xabc",
			To:       "0xdef",
			Value:    "1000",
			GasLimit: 21000,
		})
		require.Error(t, err)
	})
}
