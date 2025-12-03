package usecases

import (
	"context"
	"fmt"
	"math/big"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/events"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

// GetBalanceInput represents the input for GetBalance use case
type GetBalanceInput struct {
	ChainID      string
	Address      string
	TokenAddress string // optional, for token balance
}

// GetBalanceOutput represents the output for GetBalance use case
type GetBalanceOutput struct {
	ChainID      string
	Address      string
	Balance      *big.Int
	TokenAddress string
}

// GetBalanceUseCase handles balance queries
type GetBalanceUseCase struct {
	registry ports.ChainRegistry
	eventBus ports.EventPublisher
	logger   ports.Logger
}

// NewGetBalanceUseCase creates a new GetBalanceUseCase
func NewGetBalanceUseCase(
	registry ports.ChainRegistry,
	eventBus ports.EventPublisher,
	logger ports.Logger,
) *GetBalanceUseCase {
	return &GetBalanceUseCase{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute executes the get balance use case
func (uc *GetBalanceUseCase) Execute(ctx context.Context, input GetBalanceInput) (*GetBalanceOutput, error) {
	uc.logger.Info("executing GetBalance use case", map[string]interface{}{
		"chain_id": input.ChainID,
		"address":  input.Address,
	})

	if input.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if input.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	adapter, err := uc.registry.Get(input.ChainID)
	if err != nil {
		uc.logger.Error("failed to get chain adapter", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to get chain adapter: %w", err)
	}

	address, err := valueobjects.NewAddress(input.Address, input.ChainID)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	var balance *big.Int
	var tokenAddress *valueobjects.Address

	if input.TokenAddress != "" {
		tokenAddress, err = valueobjects.NewAddress(input.TokenAddress, input.ChainID)
		if err != nil {
			return nil, fmt.Errorf("invalid token address: %w", err)
		}

		balance, err = adapter.GetTokenBalance(ctx, input.ChainID, address, tokenAddress)
		if err != nil {
			uc.logger.Error("failed to get token balance", err, map[string]interface{}{
				"chain_id":      input.ChainID,
				"address":       input.Address,
				"token_address": input.TokenAddress,
			})
			return nil, fmt.Errorf("failed to get token balance: %w", err)
		}
	} else {
		balance, err = adapter.GetBalance(ctx, input.ChainID, address)
		if err != nil {
			uc.logger.Error("failed to get balance", err, map[string]interface{}{
				"chain_id": input.ChainID,
				"address":  input.Address,
			})
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	}

	event := events.NewBalanceQueriedEvent(input.ChainID, address, balance, tokenAddress)
	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Warn("failed to publish balance queried event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	uc.logger.Info("balance retrieved successfully", map[string]interface{}{
		"chain_id": input.ChainID,
		"address":  input.Address,
		"balance":  balance.String(),
	})

	return &GetBalanceOutput{
		ChainID:      input.ChainID,
		Address:      input.Address,
		Balance:      balance,
		TokenAddress: input.TokenAddress,
	}, nil
}
