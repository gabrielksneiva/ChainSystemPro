package usecases

import (
	"context"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

// GetTransactionStatusInput represents the input for GetTransactionStatus use case
type GetTransactionStatusInput struct {
	ChainID         string
	TransactionHash string
}

// GetTransactionStatusOutput represents the output for GetTransactionStatus use case
type GetTransactionStatusOutput struct {
	Hash          string
	Status        string
	BlockNumber   uint64
	Confirmations uint64
}

// GetTransactionStatusUseCase handles transaction status queries
type GetTransactionStatusUseCase struct {
	registry ports.ChainRegistry
	logger   ports.Logger
}

// NewGetTransactionStatusUseCase creates a new GetTransactionStatusUseCase
func NewGetTransactionStatusUseCase(
	registry ports.ChainRegistry,
	logger ports.Logger,
) *GetTransactionStatusUseCase {
	return &GetTransactionStatusUseCase{
		registry: registry,
		logger:   logger,
	}
}

// Execute executes the get transaction status use case
func (uc *GetTransactionStatusUseCase) Execute(ctx context.Context, input GetTransactionStatusInput) (*GetTransactionStatusOutput, error) {
	uc.logger.Info("executing GetTransactionStatus use case", map[string]interface{}{
		"chain_id": input.ChainID,
		"hash":     input.TransactionHash,
	})

	if input.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if input.TransactionHash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	adapter, err := uc.registry.Get(input.ChainID)
	if err != nil {
		uc.logger.Error("failed to get chain adapter", err, map[string]interface{}{
			"chain_id": input.ChainID,
		})
		return nil, fmt.Errorf("failed to get chain adapter: %w", err)
	}

	hash, err := valueobjects.NewHash(input.TransactionHash)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction hash: %w", err)
	}

	status, err := adapter.GetTransactionStatus(ctx, hash)
	if err != nil {
		uc.logger.Error("failed to get transaction status", err, map[string]interface{}{
			"chain_id": input.ChainID,
			"hash":     input.TransactionHash,
		})
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	receipt, err := adapter.GetTransactionReceipt(ctx, hash)
	if err != nil {
		uc.logger.Warn("failed to get transaction receipt", map[string]interface{}{
			"chain_id": input.ChainID,
			"hash":     input.TransactionHash,
			"error":    err.Error(),
		})
	}

	uc.logger.Info("transaction status retrieved successfully", map[string]interface{}{
		"chain_id": input.ChainID,
		"hash":     input.TransactionHash,
		"status":   string(status),
	})

	output := &GetTransactionStatusOutput{
		Hash:   hash.Hex(),
		Status: string(status),
	}

	if receipt != nil {
		output.BlockNumber = receipt.BlockNumber()
		output.Confirmations = receipt.Confirmations()
	}

	return output, nil
}
