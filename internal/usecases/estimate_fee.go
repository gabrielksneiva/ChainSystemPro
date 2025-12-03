package usecases

import (
	"context"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/events"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// EstimateFeeInput represents the input for EstimateFee use case
type EstimateFeeInput struct {
	ChainID     string
	Transaction *entities.Transaction
}

// EstimateFeeOutput represents the output for EstimateFee use case
type EstimateFeeOutput struct {
	GasLimit       uint64
	GasPrice       string
	MaxFeePerGas   string
	MaxPriorityFee string
	Total          string
	Currency       string
}

// EstimateFeeUseCase handles fee estimation
type EstimateFeeUseCase struct {
	registry ports.ChainRegistry
	eventBus ports.EventPublisher
	logger   ports.Logger
}

// NewEstimateFeeUseCase creates a new EstimateFeeUseCase
func NewEstimateFeeUseCase(
	registry ports.ChainRegistry,
	eventBus ports.EventPublisher,
	logger ports.Logger,
) *EstimateFeeUseCase {
	return &EstimateFeeUseCase{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute executes the estimate fee use case
func (uc *EstimateFeeUseCase) Execute(ctx context.Context, input EstimateFeeInput) (*EstimateFeeOutput, error) {
	uc.logger.Info("executing EstimateFee use case", map[string]interface{}{
		"chain_id": input.ChainID,
	})

	if input.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if input.Transaction == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	adapter, err := uc.registry.Get(input.ChainID)
	if err != nil {
		uc.logger.Error("failed to get chain adapter", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to get chain adapter: %w", err)
	}

	fee, err := adapter.EstimateFee(ctx, input.Transaction)
	if err != nil {
		uc.logger.Error("failed to estimate fee", err, map[string]interface{}{
			"chain_id":       input.ChainID,
			"transaction_id": input.Transaction.ID(),
		})
		return nil, fmt.Errorf("failed to estimate fee: %w", err)
	}

	event := events.NewFeeEstimatedEvent(input.ChainID, fee, input.Transaction.ID())
	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Warn("failed to publish fee estimated event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	uc.logger.Info("fee estimated successfully", map[string]interface{}{
		"chain_id": input.ChainID,
		"total":    fee.Total().String(),
	})

	output := &EstimateFeeOutput{
		GasLimit: fee.GasLimit(),
		Total:    fee.Total().String(),
		Currency: fee.Currency(),
	}

	if fee.GasPrice() != nil {
		output.GasPrice = fee.GasPrice().String()
	}
	if fee.MaxFeePerGas() != nil {
		output.MaxFeePerGas = fee.MaxFeePerGas().String()
	}
	if fee.MaxPriorityFee() != nil {
		output.MaxPriorityFee = fee.MaxPriorityFee().String()
	}

	return output, nil
}
