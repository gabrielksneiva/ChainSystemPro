package usecases

import (
	"context"
	"math/big"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/require"
)

func TestEstimateFee(t *testing.T) {
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
		adapter.EstimateFeeFunc = func(ctx context.Context, inTx *entities.Transaction) (*entities.Fee, error) {
			return entities.NewFee(21000, big.NewInt(20000000000), "ETH")
		}
		uc := NewEstimateFeeUseCase(registry, publisher, logger)
		out, err := uc.Execute(ctx, EstimateFeeInput{ChainID: "evm-mainnet", Transaction: tx})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, uint64(21000), out.GasLimit)
		require.Equal(t, "ETH", out.Currency)
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		uc := NewEstimateFeeUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, EstimateFeeInput{ChainID: "evm-mainnet", Transaction: tx})
		require.Error(t, err)
	})

	t.Run("adapter error", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		adapter.EstimateFeeFunc = func(ctx context.Context, inTx *entities.Transaction) (*entities.Fee, error) {
			return nil, simpleError{"fail"}
		}
		uc := NewEstimateFeeUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, EstimateFeeInput{ChainID: "evm-mainnet", Transaction: tx})
		require.Error(t, err)
	})

	t.Run("nil transaction", func(t *testing.T) {
		t.Parallel()
		adapter := &mocks.MockChainAdapter{}
		registry := mocks.NewMockChainRegistry()
		publisher := mocks.NewMockEventPublisher()
		logger := mocks.NewMockLogger()
		_ = registry.Register("evm-mainnet", adapter)
		uc := NewEstimateFeeUseCase(registry, publisher, logger)
		_, err := uc.Execute(ctx, EstimateFeeInput{ChainID: "evm-mainnet", Transaction: nil})
		require.Error(t, err)
	})
}
