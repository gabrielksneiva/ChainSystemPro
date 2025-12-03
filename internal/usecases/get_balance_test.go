package usecases

import (
	"context"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

// failingPublisher implements ports.EventPublisher and returns error on Publish
type failingPublisher struct{}

func (f failingPublisher) Publish(ctx context.Context, event interface{}) error {
	return simpleError{"publish failed"}
}
func (f failingPublisher) PublishBatch(ctx context.Context, events []interface{}) error { return nil }

func TestGetBalance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success native balance", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)

		adapter.GetBalanceFunc = func(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
			return big.NewInt(1000), nil
		}

		uc := NewGetBalanceUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, GetBalanceInput{ChainID: "evm-mainnet", Address: "0xabc"})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, "1000", out.Balance.String())
	})

	t.Run("success token balance", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)

		adapter.GetTokenBalanceFunc = func(ctx context.Context, chainID string, address, tokenAddress *valueobjects.Address) (*big.Int, error) {
			return big.NewInt(5000), nil
		}

		uc := NewGetBalanceUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, GetBalanceInput{ChainID: "evm-mainnet", Address: "0xabc", TokenAddress: "0xtoken"})
		require.NoError(t, err)
		require.Equal(t, "5000", out.Balance.String())
		require.Equal(t, "0xtoken", out.TokenAddress)
	})

	t.Run("validation error: missing chainID", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewGetBalanceUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, GetBalanceInput{ChainID: "", Address: "0xabc"})
		require.Error(t, err)
	})

	t.Run("validation error: missing address", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewGetBalanceUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, GetBalanceInput{ChainID: "evm-mainnet", Address: ""})
		require.Error(t, err)
	})

	t.Run("registry error: chain not found", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewGetBalanceUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, GetBalanceInput{ChainID: "unknown", Address: "0xabc"})
		require.Error(t, err)
	})

	t.Run("adapter error on balance", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)

		adapter.GetBalanceFunc = func(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
			return nil, simpleError{"balance query failed"}
		}

		uc := NewGetBalanceUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, GetBalanceInput{ChainID: "evm-mainnet", Address: "0xabc"})
		require.Error(t, err)
	})

	t.Run("event publish failure is warned but not fatal", func(t *testing.T) {
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)

		adapter.GetBalanceFunc = func(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
			return big.NewInt(42), nil
		}

		// custom publisher that returns error on Publish
		var publisher ports.EventPublisher = failingPublisher{}

		uc := NewGetBalanceUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, GetBalanceInput{ChainID: "evm-mainnet", Address: "0xabc"})
		require.NoError(t, err)
		require.Equal(t, "42", out.Balance.String())
	})
}
