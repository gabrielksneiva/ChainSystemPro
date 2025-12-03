package entities

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/google/uuid"
)

// ChainType represents the blockchain type
type ChainType string

const (
	ChainTypeEVM     ChainType = "evm"
	ChainTypeTron    ChainType = "tron"
	ChainTypeBitcoin ChainType = "bitcoin"
)

// Chain represents a blockchain network
type Chain struct {
	id          string
	name        string
	chainType   ChainType
	networkID   string
	isTestnet   bool
	nativeToken string
	createdAt   time.Time
	updatedAt   time.Time
}

// NewChain creates a new Chain entity
func NewChain(name string, chainType ChainType, networkID string, isTestnet bool, nativeToken string) (*Chain, error) {
	if name == "" {
		return nil, fmt.Errorf("chain name cannot be empty")
	}
	if chainType == "" {
		return nil, fmt.Errorf("chain type cannot be empty")
	}
	if networkID == "" {
		return nil, fmt.Errorf("network ID cannot be empty")
	}
	if nativeToken == "" {
		return nil, fmt.Errorf("native token cannot be empty")
	}

	now := time.Now()
	return &Chain{
		id:          uuid.New().String(),
		name:        name,
		chainType:   chainType,
		networkID:   networkID,
		isTestnet:   isTestnet,
		nativeToken: nativeToken,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Getters
func (c *Chain) ID() string           { return c.id }
func (c *Chain) Name() string         { return c.name }
func (c *Chain) ChainType() ChainType { return c.chainType }
func (c *Chain) NetworkID() string    { return c.networkID }
func (c *Chain) IsTestnet() bool      { return c.isTestnet }
func (c *Chain) NativeToken() string  { return c.nativeToken }
func (c *Chain) CreatedAt() time.Time { return c.createdAt }
func (c *Chain) UpdatedAt() time.Time { return c.updatedAt }

// TxStatus represents transaction status
type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusDropped   TxStatus = "dropped"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	id             string
	chainID        string
	hash           *valueobjects.Hash
	from           *valueobjects.Address
	to             *valueobjects.Address
	value          *big.Int
	data           []byte
	nonce          *valueobjects.Nonce
	gasLimit       uint64
	gasPrice       *big.Int
	maxFeePerGas   *big.Int
	maxPriorityFee *big.Int
	signature      *valueobjects.Signature
	status         TxStatus
	blockNumber    uint64
	confirmations  uint64
	metadata       map[string]interface{}
	createdAt      time.Time
	updatedAt      time.Time
}

// TransactionParams holds transaction creation parameters
type TransactionParams struct {
	ChainID        string
	From           *valueobjects.Address
	To             *valueobjects.Address
	Value          *big.Int
	Data           []byte
	Nonce          *valueobjects.Nonce
	GasLimit       uint64
	GasPrice       *big.Int
	MaxFeePerGas   *big.Int
	MaxPriorityFee *big.Int
}

// NewTransaction creates a new Transaction entity
func NewTransaction(params TransactionParams) (*Transaction, error) {
	if params.ChainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if params.From == nil {
		return nil, fmt.Errorf("from address cannot be nil")
	}
	if params.To == nil {
		return nil, fmt.Errorf("to address cannot be nil")
	}
	if params.Value == nil {
		params.Value = big.NewInt(0)
	}
	if params.Value.Sign() < 0 {
		return nil, fmt.Errorf("value cannot be negative")
	}

	now := time.Now()
	return &Transaction{
		id:             uuid.New().String(),
		chainID:        params.ChainID,
		from:           params.From,
		to:             params.To,
		value:          new(big.Int).Set(params.Value),
		data:           params.Data,
		nonce:          params.Nonce,
		gasLimit:       params.GasLimit,
		gasPrice:       params.GasPrice,
		maxFeePerGas:   params.MaxFeePerGas,
		maxPriorityFee: params.MaxPriorityFee,
		status:         TxStatusPending,
		metadata:       make(map[string]interface{}),
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// Getters
func (t *Transaction) ID() string                         { return t.id }
func (t *Transaction) ChainID() string                    { return t.chainID }
func (t *Transaction) Hash() *valueobjects.Hash           { return t.hash }
func (t *Transaction) From() *valueobjects.Address        { return t.from }
func (t *Transaction) To() *valueobjects.Address          { return t.to }
func (t *Transaction) Value() *big.Int                    { return new(big.Int).Set(t.value) }
func (t *Transaction) Data() []byte                       { return t.data }
func (t *Transaction) Nonce() *valueobjects.Nonce         { return t.nonce }
func (t *Transaction) GasLimit() uint64                   { return t.gasLimit }
func (t *Transaction) GasPrice() *big.Int                 { return t.gasPrice }
func (t *Transaction) MaxFeePerGas() *big.Int             { return t.maxFeePerGas }
func (t *Transaction) MaxPriorityFee() *big.Int           { return t.maxPriorityFee }
func (t *Transaction) Signature() *valueobjects.Signature { return t.signature }
func (t *Transaction) Status() TxStatus                   { return t.status }
func (t *Transaction) BlockNumber() uint64                { return t.blockNumber }
func (t *Transaction) Confirmations() uint64              { return t.confirmations }
func (t *Transaction) Metadata() map[string]interface{}   { return t.metadata }
func (t *Transaction) CreatedAt() time.Time               { return t.createdAt }
func (t *Transaction) UpdatedAt() time.Time               { return t.updatedAt }

// SetHash sets the transaction hash
func (t *Transaction) SetHash(hash *valueobjects.Hash) error {
	if hash == nil {
		return fmt.Errorf("hash cannot be nil")
	}
	t.hash = hash
	t.updatedAt = time.Now()
	return nil
}

// SetSignature sets the transaction signature
func (t *Transaction) SetSignature(sig *valueobjects.Signature) error {
	if sig == nil {
		return fmt.Errorf("signature cannot be nil")
	}
	t.signature = sig
	t.updatedAt = time.Now()
	return nil
}

// UpdateStatus updates the transaction status
func (t *Transaction) UpdateStatus(status TxStatus) {
	t.status = status
	t.updatedAt = time.Now()
}

// SetBlockNumber sets the block number
func (t *Transaction) SetBlockNumber(blockNumber uint64) {
	t.blockNumber = blockNumber
	t.updatedAt = time.Now()
}

// SetConfirmations sets the number of confirmations
func (t *Transaction) SetConfirmations(confirmations uint64) {
	t.confirmations = confirmations
	t.updatedAt = time.Now()
}

// SetMetadata sets a metadata value
func (t *Transaction) SetMetadata(key string, value interface{}) {
	t.metadata[key] = value
	t.updatedAt = time.Now()
}

// Wallet represents a blockchain wallet/account
type Wallet struct {
	id        string
	address   *valueobjects.Address
	chainID   string
	label     string
	metadata  map[string]interface{}
	createdAt time.Time
	updatedAt time.Time
}

// NewWallet creates a new Wallet entity
func NewWallet(address *valueobjects.Address, chainID, label string) (*Wallet, error) {
	if address == nil {
		return nil, fmt.Errorf("address cannot be nil")
	}
	if chainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}

	now := time.Now()
	return &Wallet{
		id:        uuid.New().String(),
		address:   address,
		chainID:   chainID,
		label:     label,
		metadata:  make(map[string]interface{}),
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (w *Wallet) ID() string                       { return w.id }
func (w *Wallet) Address() *valueobjects.Address   { return w.address }
func (w *Wallet) ChainID() string                  { return w.chainID }
func (w *Wallet) Label() string                    { return w.label }
func (w *Wallet) Metadata() map[string]interface{} { return w.metadata }
func (w *Wallet) CreatedAt() time.Time             { return w.createdAt }
func (w *Wallet) UpdatedAt() time.Time             { return w.updatedAt }

// SetMetadata sets a metadata value
func (w *Wallet) SetMetadata(key string, value interface{}) {
	w.metadata[key] = value
	w.updatedAt = time.Now()
}

// UpdateLabel updates the wallet label
func (w *Wallet) UpdateLabel(label string) {
	w.label = label
	w.updatedAt = time.Now()
}

// Fee represents transaction fee information
type Fee struct {
	gasLimit       uint64
	gasPrice       *big.Int
	maxFeePerGas   *big.Int
	maxPriorityFee *big.Int
	total          *big.Int
	currency       string
}

// NewFee creates a new Fee entity
func NewFee(gasLimit uint64, gasPrice *big.Int, currency string) (*Fee, error) {
	if gasLimit == 0 {
		return nil, fmt.Errorf("gas limit cannot be zero")
	}
	if gasPrice == nil || gasPrice.Sign() <= 0 {
		return nil, fmt.Errorf("gas price must be positive")
	}
	if currency == "" {
		return nil, fmt.Errorf("currency cannot be empty")
	}
	if gasLimit > 9223372036854775807 {
		return nil, fmt.Errorf("gas limit exceeds maximum safe value")
	}

	total := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)

	return &Fee{
		gasLimit: gasLimit,
		gasPrice: new(big.Int).Set(gasPrice),
		total:    total,
		currency: currency,
	}, nil
}

// NewEIP1559Fee creates a new Fee with EIP-1559 parameters
func NewEIP1559Fee(gasLimit uint64, maxFeePerGas, maxPriorityFee *big.Int, currency string) (*Fee, error) {
	if gasLimit == 0 {
		return nil, fmt.Errorf("gas limit cannot be zero")
	}
	if maxFeePerGas == nil || maxFeePerGas.Sign() <= 0 {
		return nil, fmt.Errorf("max fee per gas must be positive")
	}
	if maxPriorityFee == nil || maxPriorityFee.Sign() < 0 {
		return nil, fmt.Errorf("max priority fee cannot be negative")
	}
	if currency == "" {
		return nil, fmt.Errorf("currency cannot be empty")
	}
	if gasLimit > 9223372036854775807 {
		return nil, fmt.Errorf("gas limit exceeds maximum safe value")
	}

	total := new(big.Int).Mul(big.NewInt(int64(gasLimit)), maxFeePerGas)

	return &Fee{
		gasLimit:       gasLimit,
		maxFeePerGas:   new(big.Int).Set(maxFeePerGas),
		maxPriorityFee: new(big.Int).Set(maxPriorityFee),
		total:          total,
		currency:       currency,
	}, nil
}

// Getters
func (f *Fee) GasLimit() uint64         { return f.gasLimit }
func (f *Fee) GasPrice() *big.Int       { return f.gasPrice }
func (f *Fee) MaxFeePerGas() *big.Int   { return f.maxFeePerGas }
func (f *Fee) MaxPriorityFee() *big.Int { return f.maxPriorityFee }
func (f *Fee) Total() *big.Int          { return new(big.Int).Set(f.total) }
func (f *Fee) Currency() string         { return f.currency }

// Network represents network information
type Network struct {
	id          string
	chainID     string
	name        string
	rpcURL      string
	explorerURL string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewNetwork creates a new Network entity
func NewNetwork(chainID, name, rpcURL string) (*Network, error) {
	if chainID == "" {
		return nil, fmt.Errorf("chain ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL cannot be empty")
	}

	now := time.Now()
	return &Network{
		id:        uuid.New().String(),
		chainID:   chainID,
		name:      name,
		rpcURL:    rpcURL,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (n *Network) ID() string           { return n.id }
func (n *Network) ChainID() string      { return n.chainID }
func (n *Network) Name() string         { return n.name }
func (n *Network) RPCURL() string       { return n.rpcURL }
func (n *Network) ExplorerURL() string  { return n.explorerURL }
func (n *Network) IsActive() bool       { return n.isActive }
func (n *Network) CreatedAt() time.Time { return n.createdAt }
func (n *Network) UpdatedAt() time.Time { return n.updatedAt }

// SetExplorerURL sets the explorer URL
func (n *Network) SetExplorerURL(url string) {
	n.explorerURL = url
	n.updatedAt = time.Now()
}

// Activate activates the network
func (n *Network) Activate() {
	n.isActive = true
	n.updatedAt = time.Now()
}

// Deactivate deactivates the network
func (n *Network) Deactivate() {
	n.isActive = false
	n.updatedAt = time.Now()
}
