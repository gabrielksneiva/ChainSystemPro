package ports

import (
	"context"
	"math/big"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

// ChainConnector defines the interface for blockchain connectors
type ChainConnector interface {
	// GetChainID returns the chain identifier
	GetChainID() string

	// GetChainType returns the chain type
	GetChainType() entities.ChainType

	// IsConnected checks if the connector is connected to the chain
	IsConnected(ctx context.Context) bool

	// GetBlockNumber returns the current block number
	GetBlockNumber(ctx context.Context) (uint64, error)

	// GetNativeBalance returns the native token balance for an address
	GetNativeBalance(ctx context.Context, address *valueobjects.Address) (*big.Int, error)
}

// BalanceProvider defines the interface for querying balances
type BalanceProvider interface {
	// GetBalance returns the balance for a given address
	GetBalance(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error)

	// GetTokenBalance returns the token balance for a given address and token
	GetTokenBalance(ctx context.Context, chainID string, address *valueobjects.Address, tokenAddress *valueobjects.Address) (*big.Int, error)
}

// TransactionBuilder defines the interface for building transactions
type TransactionBuilder interface {
	// BuildTransaction creates a new transaction
	BuildTransaction(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error)

	// EstimateGas estimates the gas required for a transaction
	EstimateGas(ctx context.Context, tx *entities.Transaction) (uint64, error)

	// SetNonce sets the nonce for a transaction
	SetNonce(ctx context.Context, tx *entities.Transaction) error
}

// TransactionSigner defines the interface for signing transactions
type TransactionSigner interface {
	// SignTransaction signs a transaction
	SignTransaction(ctx context.Context, tx *entities.Transaction, privateKey []byte) error

	// VerifySignature verifies a transaction signature
	VerifySignature(ctx context.Context, tx *entities.Transaction) (bool, error)
}

// TxBroadcaster defines the interface for broadcasting transactions
type TxBroadcaster interface {
	// BroadcastTransaction broadcasts a signed transaction to the network
	BroadcastTransaction(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error)

	// GetTransactionStatus returns the status of a transaction
	GetTransactionStatus(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error)

	// GetTransactionReceipt returns the transaction receipt
	GetTransactionReceipt(ctx context.Context, hash *valueobjects.Hash) (*entities.Transaction, error)

	// WaitForConfirmation waits for a transaction to be confirmed
	WaitForConfirmation(ctx context.Context, hash *valueobjects.Hash, confirmations uint64) error
}

// FeeEstimator defines the interface for estimating transaction fees
type FeeEstimator interface {
	// EstimateFee estimates the fee for a transaction
	EstimateFee(ctx context.Context, tx *entities.Transaction) (*entities.Fee, error)

	// GetGasPrice returns the current gas price
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// GetMaxPriorityFee returns the max priority fee (EIP-1559)
	GetMaxPriorityFee(ctx context.Context) (*big.Int, error)
}

// NetworkInfoProvider defines the interface for network information
type NetworkInfoProvider interface {
	// GetNetworkInfo returns the network information
	GetNetworkInfo(ctx context.Context) (*entities.Network, error)

	// GetPeers returns the number of connected peers
	GetPeers(ctx context.Context) (int, error)

	// GetLatestBlock returns the latest block number
	GetLatestBlock(ctx context.Context) (uint64, error)
}

// ChainAdapter combines all chain-specific interfaces
type ChainAdapter interface {
	ChainConnector
	BalanceProvider
	TransactionBuilder
	TransactionSigner
	TxBroadcaster
	FeeEstimator
	NetworkInfoProvider
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// Publish publishes an event
	Publish(ctx context.Context, event interface{}) error

	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []interface{}) error
}

// EventSubscriber defines the interface for subscribing to domain events
type EventSubscriber interface {
	// Subscribe subscribes to events of a specific type
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error

	// Unsubscribe unsubscribes from events of a specific type
	Unsubscribe(ctx context.Context, eventType string) error
}

// EventHandler handles domain events
type EventHandler func(ctx context.Context, event interface{}) error

// EventBus combines publisher and subscriber
type EventBus interface {
	EventPublisher
	EventSubscriber

	// Start starts the event bus
	Start(ctx context.Context) error

	// Stop stops the event bus
	Stop(ctx context.Context) error
}

// ChainRegistry defines the interface for managing chain adapters
type ChainRegistry interface {
	// Register registers a chain adapter
	Register(chainID string, adapter ChainAdapter) error

	// Unregister unregisters a chain adapter
	Unregister(chainID string) error

	// Get returns a chain adapter by ID
	Get(chainID string) (ChainAdapter, error)

	// List returns all registered chain IDs
	List() []string

	// Has checks if a chain adapter is registered
	Has(chainID string) bool
}

// Logger defines the interface for structured logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields map[string]interface{})

	// Info logs an info message
	Info(msg string, fields map[string]interface{})

	// Warn logs a warning message
	Warn(msg string, fields map[string]interface{})

	// Error logs an error message
	Error(msg string, err error, fields map[string]interface{})

	// Fatal logs a fatal message and exits
	Fatal(msg string, err error, fields map[string]interface{})
}

// ConfigProvider defines the interface for configuration management
type ConfigProvider interface {
	// GetString returns a string config value
	GetString(key string) string

	// GetInt returns an int config value
	GetInt(key string) int

	// GetBool returns a bool config value
	GetBool(key string) bool

	// GetStringSlice returns a string slice config value
	GetStringSlice(key string) []string

	// IsSet checks if a config key is set
	IsSet(key string) bool
}
