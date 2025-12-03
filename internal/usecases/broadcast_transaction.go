package usecases

import (
	"context"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/events"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// BroadcastTransactionInput represents the input for BroadcastTransaction use case
type BroadcastTransactionInput struct {
	ChainID     string
	Transaction *entities.Transaction
}

// BroadcastTransactionOutput represents the output for BroadcastTransaction use case
type BroadcastTransactionOutput struct {
	TransactionID string
	Hash          string
	Status        string
}

// BroadcastTransactionUseCase handles transaction broadcasting
type BroadcastTransactionUseCase struct {
	registry ports.ChainRegistry
	eventBus ports.EventPublisher
	logger   ports.Logger
}

// NewBroadcastTransactionUseCase creates a new BroadcastTransactionUseCase
func NewBroadcastTransactionUseCase(
	registry ports.ChainRegistry,
	eventBus ports.EventPublisher,
	logger ports.Logger,
) *BroadcastTransactionUseCase {
	return &BroadcastTransactionUseCase{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute executes the broadcast transaction use case
func (uc *BroadcastTransactionUseCase) Execute(ctx context.Context, input BroadcastTransactionInput) (*BroadcastTransactionOutput, error) {
	uc.logger.Info("executing BroadcastTransaction use case", map[string]interface{}{
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

	hash, err := adapter.BroadcastTransaction(ctx, input.Transaction)
	if err != nil {
		uc.logger.Error("failed to broadcast transaction", err, map[string]interface{}{
			"chain_id":       input.ChainID,
			"transaction_id": input.Transaction.ID(),
		})

		event := events.NewTransactionFailedEvent(
			input.ChainID,
			input.Transaction.ID(),
			"",
			err.Error(),
			"BROADCAST_ERROR",
		)
		if pubErr := uc.eventBus.Publish(ctx, event); pubErr != nil {
			// Log the publish error but don't override the original broadcast error
			fmt.Printf("failed to publish broadcast error event: %v\n", pubErr)
		}

		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	event := events.NewTransactionBroadcastedEvent(input.ChainID, input.Transaction.ID(), hash)
	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Warn("failed to publish transaction broadcasted event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	uc.logger.Info("transaction broadcasted successfully", map[string]interface{}{
		"chain_id":       input.ChainID,
		"transaction_id": input.Transaction.ID(),
		"hash":           hash.Hex(),
	})

	return &BroadcastTransactionOutput{
		TransactionID: input.Transaction.ID(),
		Hash:          hash.Hex(),
		Status:        string(entities.TxStatusPending),
	}, nil
}
