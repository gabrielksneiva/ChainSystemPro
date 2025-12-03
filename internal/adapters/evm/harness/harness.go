package harness

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
)

type EVMHarness struct {
	chainID      string
	accounts     map[string]*big.Int
	transactions map[string]*entities.Transaction
	blockNumber  uint64
	gasPrice     *big.Int
	mu           sync.RWMutex
}

func NewEVMHarness(chainID string) *EVMHarness {
	return &EVMHarness{
		chainID:      chainID,
		accounts:     make(map[string]*big.Int),
		transactions: make(map[string]*entities.Transaction),
		blockNumber:  1,
		gasPrice:     big.NewInt(20000000000),
	}
}

func (h *EVMHarness) GetChainID() string {
	return h.chainID
}

func (h *EVMHarness) GetChainType() entities.ChainType {
	return entities.ChainTypeEVM
}

func (h *EVMHarness) IsConnected(ctx context.Context) bool {
	return true
}

func (h *EVMHarness) GetBlockNumber(ctx context.Context) (uint64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.blockNumber, nil
}

func (h *EVMHarness) GetNativeBalance(ctx context.Context, address *valueobjects.Address) (*big.Int, error) {
	return h.GetBalance(ctx, h.chainID, address)
}

func (h *EVMHarness) GetBalance(ctx context.Context, chainID string, address *valueobjects.Address) (*big.Int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	balance, exists := h.accounts[address.Value()]
	if !exists {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(balance), nil
}

func (h *EVMHarness) GetTokenBalance(ctx context.Context, chainID string, address, tokenAddress *valueobjects.Address) (*big.Int, error) {
	key := fmt.Sprintf("%s:%s", address.Value(), tokenAddress.Value())
	h.mu.RLock()
	defer h.mu.RUnlock()
	balance, exists := h.accounts[key]
	if !exists {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(balance), nil
}

func (h *EVMHarness) BuildTransaction(ctx context.Context, params entities.TransactionParams) (*entities.Transaction, error) {
	h.mu.RLock()
	nonce := uint64(len(h.transactions))
	h.mu.RUnlock()
	params.Nonce = valueobjects.NewNonce(nonce)
	if params.GasLimit == 0 {
		params.GasLimit = 21000
	}
	if params.GasPrice == nil {
		params.GasPrice = new(big.Int).Set(h.gasPrice)
	}
	return entities.NewTransaction(params)
}

func (h *EVMHarness) EstimateGas(ctx context.Context, tx *entities.Transaction) (uint64, error) {
	if len(tx.Data()) > 0 {
		return 50000, nil
	}
	return 21000, nil
}

func (h *EVMHarness) SetNonce(ctx context.Context, tx *entities.Transaction) error {
	return nil
}

func (h *EVMHarness) SignTransaction(ctx context.Context, tx *entities.Transaction, privateKey []byte) error {
	if len(privateKey) == 0 {
		return fmt.Errorf("private key cannot be empty")
	}
	hashBytes := make([]byte, 32)
	if _, err := rand.Read(hashBytes); err != nil {
		return fmt.Errorf("failed to generate random hash: %w", err)
	}
	hash, _ := valueobjects.NewHashFromBytes(hashBytes)
	sigBytes := make([]byte, 65)
	if _, err := rand.Read(sigBytes); err != nil {
		return fmt.Errorf("failed to generate random signature: %w", err)
	}
	sig, _ := valueobjects.NewSignatureFromBytes(sigBytes)
	if err := tx.SetHash(hash); err != nil {
		return err
	}
	return tx.SetSignature(sig)
}

func (h *EVMHarness) VerifySignature(ctx context.Context, tx *entities.Transaction) (bool, error) {
	return tx.Signature() != nil, nil
}

func (h *EVMHarness) BroadcastTransaction(ctx context.Context, tx *entities.Transaction) (*valueobjects.Hash, error) {
	if tx.Hash() == nil {
		return nil, fmt.Errorf("transaction must be signed")
	}
	h.mu.Lock()
	h.transactions[tx.Hash().Hex()] = tx
	h.blockNumber++
	h.mu.Unlock()
	tx.UpdateStatus(entities.TxStatusConfirmed)
	tx.SetBlockNumber(h.blockNumber - 1)
	tx.SetConfirmations(1)
	return tx.Hash(), nil
}

func (h *EVMHarness) GetTransactionStatus(ctx context.Context, hash *valueobjects.Hash) (entities.TxStatus, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tx, exists := h.transactions[hash.Hex()]
	if !exists {
		return "", fmt.Errorf("transaction not found")
	}
	return tx.Status(), nil
}

func (h *EVMHarness) GetTransactionReceipt(ctx context.Context, hash *valueobjects.Hash) (*entities.Transaction, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tx, exists := h.transactions[hash.Hex()]
	if !exists {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, nil
}

func (h *EVMHarness) WaitForConfirmation(ctx context.Context, hash *valueobjects.Hash, confirmations uint64) error {
	return nil
}

func (h *EVMHarness) EstimateFee(ctx context.Context, tx *entities.Transaction) (*entities.Fee, error) {
	gasLimit := tx.GasLimit()
	if gasLimit == 0 {
		gasLimit = 21000
	}
	return entities.NewFee(gasLimit, h.gasPrice, "ETH")
}

func (h *EVMHarness) GetGasPrice(ctx context.Context) (*big.Int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return new(big.Int).Set(h.gasPrice), nil
}

func (h *EVMHarness) GetMaxPriorityFee(ctx context.Context) (*big.Int, error) {
	return big.NewInt(2000000000), nil
}

func (h *EVMHarness) GetNetworkInfo(ctx context.Context) (*entities.Network, error) {
	return entities.NewNetwork(h.chainID, "EVM Harness", "memory://localhost")
}

func (h *EVMHarness) GetPeers(ctx context.Context) (int, error) {
	return 5, nil
}

func (h *EVMHarness) GetLatestBlock(ctx context.Context) (uint64, error) {
	return h.GetBlockNumber(ctx)
}

func (h *EVMHarness) SetBalance(address string, balance *big.Int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.accounts[address] = new(big.Int).Set(balance)
}

func (h *EVMHarness) SetTokenBalance(address, tokenAddress string, balance *big.Int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := fmt.Sprintf("%s:%s", address, tokenAddress)
	h.accounts[key] = new(big.Int).Set(balance)
}
