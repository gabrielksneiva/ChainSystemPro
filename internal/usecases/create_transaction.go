package usecases

import (
	"context"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/events"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

// CreateTransactionInput represents the input for CreateTransaction use case
type CreateTransactionInput struct {
	ChainID  string
	From     string
	To       string
	Value    string
	Data     []byte
	GasLimit uint64
}

// CreateTransactionOutput represents the output for CreateTransaction use case
type CreateTransactionOutput struct {
	TransactionID string
	ChainID       string
	From          string
	To            string
	Value         string
	Nonce         uint64
	GasLimit      uint64
}

// CreateTransactionUseCase handles transaction creation
type CreateTransactionUseCase struct {
	registry ports.ChainRegistry
	eventBus ports.EventPublisher
	logger   ports.Logger
}

// NewCreateTransactionUseCase creates a new CreateTransactionUseCase
func NewCreateTransactionUseCase(
	registry ports.ChainRegistry,
	eventBus ports.EventPublisher,
	logger ports.Logger,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute executes the create transaction use case
func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*CreateTransactionOutput, error) {
	uc.logger.Info("executing CreateTransaction use case", map[string]interface{}{
		"chain_id": input.ChainID,
		"from":     input.From,
		"to":       input.To,
	})

	if input.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if input.From == "" {
		return nil, fmt.Errorf("from address cannot be empty")
	}
	if input.To == "" {
		return nil, fmt.Errorf("to address cannot be empty")
	}

	adapter, err := uc.registry.Get(input.ChainID)
	if err != nil {
		uc.logger.Error("failed to get chain adapter", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to get chain adapter: %w", err)
	}

	from, err := valueobjects.NewAddress(input.From, input.ChainID)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	to, err := valueobjects.NewAddress(input.To, input.ChainID)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}

	value, ok := parseBigInt(input.Value)
	if !ok {
		return nil, fmt.Errorf("invalid value: %s", input.Value)
	}

	params := entities.TransactionParams{
		ChainID:  input.ChainID,
		From:     from,
		To:       to,
		Value:    value,
		Data:     input.Data,
		GasLimit: input.GasLimit,
	}

	tx, err := adapter.BuildTransaction(ctx, params)
	if err != nil {
		uc.logger.Error("failed to build transaction", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	event := events.NewTransactionCreatedEvent(tx)
	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Warn("failed to publish transaction created event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	uc.logger.Info("transaction created successfully", map[string]interface{}{
		"chain_id":       input.ChainID,
		"transaction_id": tx.ID(),
	})

	var nonce uint64
	if tx.Nonce() != nil {
		nonce = tx.Nonce().Value()
	}

	return &CreateTransactionOutput{
		TransactionID: tx.ID(),
		ChainID:       tx.ChainID(),
		From:          tx.From().String(),
		To:            tx.To().String(),
		Value:         tx.Value().String(),
		Nonce:         nonce,
		GasLimit:      tx.GasLimit(),
	}, nil
}
