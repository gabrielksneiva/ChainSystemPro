package usecases

import (
	"context"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/events"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// SignTransactionInput represents the input for SignTransaction use case
type SignTransactionInput struct {
	ChainID     string
	Transaction *entities.Transaction
	PrivateKey  []byte
}

// SignTransactionOutput represents the output for SignTransaction use case
type SignTransactionOutput struct {
	TransactionID string
	Hash          string
	Signature     string
}

// SignTransactionUseCase handles transaction signing
type SignTransactionUseCase struct {
	registry ports.ChainRegistry
	eventBus ports.EventPublisher
	logger   ports.Logger
}

// NewSignTransactionUseCase creates a new SignTransactionUseCase
func NewSignTransactionUseCase(
	registry ports.ChainRegistry,
	eventBus ports.EventPublisher,
	logger ports.Logger,
) *SignTransactionUseCase {
	return &SignTransactionUseCase{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute executes the sign transaction use case
func (uc *SignTransactionUseCase) Execute(ctx context.Context, input SignTransactionInput) (*SignTransactionOutput, error) {
	uc.logger.Info("executing SignTransaction use case", map[string]interface{}{
		"chain_id": input.ChainID,
	})

	if input.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if input.Transaction == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}
	if len(input.PrivateKey) == 0 {
		return nil, fmt.Errorf("private key cannot be empty")
	}

	adapter, err := uc.registry.Get(input.ChainID)
	if err != nil {
		uc.logger.Error("failed to get chain adapter", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to get chain adapter: %w", err)
	}

	if err := adapter.SignTransaction(ctx, input.Transaction, input.PrivateKey); err != nil {
		uc.logger.Error("failed to sign transaction", err, map[string]interface{}{
			"chain_id":       input.ChainID,
			"transaction_id": input.Transaction.ID(),
		})
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	event := events.NewTransactionSignedEvent(input.Transaction)
	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Warn("failed to publish transaction signed event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	uc.logger.Info("transaction signed successfully", map[string]interface{}{
		"chain_id":       input.ChainID,
		"transaction_id": input.Transaction.ID(),
	})

	var hash, signature string
	if input.Transaction.Hash() != nil {
		hash = input.Transaction.Hash().Hex()
	}
	if input.Transaction.Signature() != nil {
		signature = input.Transaction.Signature().Hex()
	}

	return &SignTransactionOutput{
		TransactionID: input.Transaction.ID(),
		Hash:          hash,
		Signature:     signature,
	}, nil
}
