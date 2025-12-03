package bitcoin

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

// Adapter implements the ChainAdapter interface for Bitcoin
type Adapter struct {
	rpcClient RPCClient
	network   string
}

// RPCClient defines the interface for Bitcoin RPC operations
type RPCClient interface {
	GetBlockCount(ctx context.Context) (int64, error)
	GetBalance(ctx context.Context, address string) (*big.Int, error)
	ListUnspent(ctx context.Context, address string) ([]UTXO, error)
	GetRawTransaction(ctx context.Context, txHash string) (*Transaction, error)
	SendRawTransaction(ctx context.Context, rawTx string) (string, error)
	EstimateFee(ctx context.Context, blocks int) (*big.Int, error)
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID          string
	Vout          uint32
	Address       string
	ScriptPubKey  string
	Amount        int64 // satoshis
	Confirmations int64
}

// Transaction represents a Bitcoin transaction
type Transaction struct {
	TxID          string
	Hash          string
	Version       int32
	Size          int32
	VSize         int32
	Weight        int32
	LockTime      uint32
	Inputs        []TxInput
	Outputs       []TxOutput
	BlockHash     string
	Confirmations int64
	Time          int64
	BlockTime     int64
}

// TxInput represents a transaction input
type TxInput struct {
	TxID      string
	Vout      uint32
	ScriptSig string
	Sequence  uint32
	Witness   []string
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value        int64
	N            uint32
	ScriptPubKey string
	Address      string
}

// NewAdapter creates a new Bitcoin adapter
func NewAdapter(rpcClient RPCClient, network string) *Adapter {
	return &Adapter{
		rpcClient: rpcClient,
		network:   network,
	}
}

// GetChainID returns the chain identifier
func (a *Adapter) GetChainID() string {
	return fmt.Sprintf("bitcoin-%s", a.network)
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() entities.ChainType {
	return entities.ChainTypeBitcoin
}

// IsConnected checks if the adapter is connected to the Bitcoin network
func (a *Adapter) IsConnected(ctx context.Context) bool {
	_, err := a.rpcClient.GetBlockCount(ctx)
	return err == nil
}

// GetBlockNumber returns the current block height
func (a *Adapter) GetBlockNumber(ctx context.Context) (uint64, error) {
	height, err := a.rpcClient.GetBlockCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}
	return uint64(height), nil
}

// GetNativeBalance returns the Bitcoin balance for an address
func (a *Adapter) GetNativeBalance(ctx context.Context, address *valueobjects.Address) (*big.Int, error) {
	balance, err := a.rpcClient.GetBalance(ctx, address.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// CreateTransaction creates a new Bitcoin transaction using UTXO selection
func (a *Adapter) CreateTransaction(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error) {
	// Get UTXOs for the sender address
	utxos, err := a.rpcClient.ListUnspent(ctx, params.From.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list unspent: %w", err)
	}

	// Select UTXOs using coin selection algorithm
	feePerByte := params.GasPrice
	if feePerByte == nil {
		feePerByte = big.NewInt(1000) // Default fee
	}
	selectedUTXOs, changeAmount, err := a.selectCoins(utxos, params.Value, feePerByte)
	if err != nil {
		return nil, fmt.Errorf("coin selection failed: %w", err)
	}

	// Build transaction with params
	txParams := entities.TransactionParams{
		ChainID:  a.GetChainID(),
		From:     params.From,
		To:       params.To,
		Value:    params.Value,
		GasPrice: feePerByte,
	}

	tx, err := entities.NewTransaction(txParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Store UTXO information in metadata for signing
	tx.SetMetadata("utxos", selectedUTXOs)
	tx.SetMetadata("change_amount", changeAmount.String())

	return tx, nil
}

// selectCoins implements a simple coin selection algorithm
func (a *Adapter) selectCoins(utxos []UTXO, amount *big.Int, feePerByte *big.Int) ([]UTXO, *big.Int, error) {
	var selected []UTXO
	total := big.NewInt(0)
	required := new(big.Int).Set(amount)

	// Estimate fee (simplified - assume 250 bytes per input, 34 bytes per output)
	estimatedSize := int64(10 + len(utxos)*250 + 2*34)
	fee := new(big.Int).Mul(feePerByte, big.NewInt(estimatedSize))
	required.Add(required, fee)

	// Simple greedy selection
	for _, utxo := range utxos {
		if total.Cmp(required) >= 0 {
			break
		}
		selected = append(selected, utxo)
		total.Add(total, big.NewInt(utxo.Amount))
	}

	if total.Cmp(required) < 0 {
		return nil, nil, fmt.Errorf("insufficient funds: need %s, have %s", required.String(), total.String())
	}

	change := new(big.Int).Sub(total, required)
	return selected, change, nil
}

// SignTransaction signs a Bitcoin transaction
func (a *Adapter) SignTransaction(ctx context.Context, tx *entities.Transaction, privateKey []byte) error {
	// In production, this would use proper Bitcoin signing with ECDSA
	// For now, we create a simplified signature
	signature := a.createSignature(tx, privateKey)

	sig, err := valueobjects.NewSignature(hex.EncodeToString(signature))
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	if err := tx.SetSignature(sig); err != nil {
		return fmt.Errorf("failed to set signature: %w", err)
	}

	return nil
} // createSignature creates a transaction signature (simplified)
func (a *Adapter) createSignature(tx *entities.Transaction, privateKey []byte) []byte {
	// In production, use proper Bitcoin SIGHASH and ECDSA signing
	// This is a simplified version for demonstration
	data := fmt.Sprintf("%s:%s:%s:%s",
		tx.From().String(),
		tx.To().String(),
		tx.Value().String(),
		tx.ID(),
	)
	return []byte(hex.EncodeToString([]byte(data)))
}

// BroadcastTransaction broadcasts a signed Bitcoin transaction
func (a *Adapter) BroadcastTransaction(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error) {
	if tx.Signature() == nil {
		return nil, fmt.Errorf("transaction not signed")
	}

	// Build raw transaction
	rawTx := a.buildRawTransaction(tx)

	txHash, err := a.rpcClient.SendRawTransaction(ctx, rawTx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	hash, err := valueobjects.NewHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create hash: %w", err)
	}

	return hash, nil
}

// buildRawTransaction builds a raw transaction hex string
func (a *Adapter) buildRawTransaction(tx *entities.Transaction) string {
	// In production, properly encode transaction according to Bitcoin protocol
	// This is simplified for demonstration
	return hex.EncodeToString([]byte(tx.ID()))
}

// GetTransactionStatus returns the status of a transaction
func (a *Adapter) GetTransactionStatus(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error) {
	btcTx, err := a.rpcClient.GetRawTransaction(ctx, hash.String())
	if err != nil {
		return entities.TxStatusPending, fmt.Errorf("failed to get transaction: %w", err)
	}

	if btcTx.Confirmations == 0 {
		return entities.TxStatusPending, nil
	} else if btcTx.Confirmations > 0 {
		return entities.TxStatusConfirmed, nil
	}

	return entities.TxStatusFailed, nil
}

// EstimateFee estimates the transaction fee
func (a *Adapter) EstimateFee(ctx context.Context, tx *entities.Transaction) (*entities.Fee, error) {
	// Estimate fee per byte for 6 block confirmation
	feePerByte, err := a.rpcClient.EstimateFee(ctx, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate fee: %w", err)
	}

	// Estimate transaction size (simplified)
	estimatedSize := uint64(250) // Average transaction size

	return entities.NewFee(estimatedSize, feePerByte, "BTC")
}
