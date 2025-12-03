package entities

import (
	"math/big"
	"testing"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChain(t *testing.T) {
	tests := []struct {
		name        string
		chainName   string
		chainType   ChainType
		networkID   string
		isTestnet   bool
		nativeToken string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid chain",
			chainName:   "Ethereum",
			chainType:   ChainTypeEVM,
			networkID:   "1",
			isTestnet:   false,
			nativeToken: "ETH",
			wantErr:     false,
		},
		{
			name:        "empty name",
			chainName:   "",
			chainType:   ChainTypeEVM,
			networkID:   "1",
			isTestnet:   false,
			nativeToken: "ETH",
			wantErr:     true,
			errMsg:      "chain name cannot be empty",
		},
		{
			name:        "empty chain type",
			chainName:   "Ethereum",
			chainType:   "",
			networkID:   "1",
			isTestnet:   false,
			nativeToken: "ETH",
			wantErr:     true,
			errMsg:      "chain type cannot be empty",
		},
		{
			name:        "empty network ID",
			chainName:   "Ethereum",
			chainType:   ChainTypeEVM,
			networkID:   "",
			isTestnet:   false,
			nativeToken: "ETH",
			wantErr:     true,
			errMsg:      "network ID cannot be empty",
		},
		{
			name:        "empty native token",
			chainName:   "Ethereum",
			chainType:   ChainTypeEVM,
			networkID:   "1",
			isTestnet:   false,
			nativeToken: "",
			wantErr:     true,
			errMsg:      "native token cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := NewChain(tt.chainName, tt.chainType, tt.networkID, tt.isTestnet, tt.nativeToken)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, chain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, chain)
				assert.NotEmpty(t, chain.ID())
				assert.Equal(t, tt.chainName, chain.Name())
				assert.Equal(t, tt.chainType, chain.ChainType())
				assert.Equal(t, tt.networkID, chain.NetworkID())
				assert.Equal(t, tt.isTestnet, chain.IsTestnet())
				assert.Equal(t, tt.nativeToken, chain.NativeToken())
				assert.False(t, chain.CreatedAt().IsZero())
				assert.False(t, chain.UpdatedAt().IsZero())
			}
		})
	}
}

func TestNewTransaction(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	nonce := valueobjects.NewNonce(1)

	tests := []struct {
		name    string
		params  TransactionParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid transaction",
			params: TransactionParams{
				ChainID:  "ethereum",
				From:     from,
				To:       to,
				Value:    big.NewInt(1000),
				Nonce:    nonce,
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
			},
			wantErr: false,
		},
		{
			name: "empty chain ID",
			params: TransactionParams{
				ChainID:  "",
				From:     from,
				To:       to,
				Value:    big.NewInt(1000),
				Nonce:    nonce,
				GasLimit: 21000,
			},
			wantErr: true,
			errMsg:  "chain ID cannot be empty",
		},
		{
			name: "nil from address",
			params: TransactionParams{
				ChainID:  "ethereum",
				From:     nil,
				To:       to,
				Value:    big.NewInt(1000),
				Nonce:    nonce,
				GasLimit: 21000,
			},
			wantErr: true,
			errMsg:  "from address cannot be nil",
		},
		{
			name: "nil to address",
			params: TransactionParams{
				ChainID:  "ethereum",
				From:     from,
				To:       nil,
				Value:    big.NewInt(1000),
				Nonce:    nonce,
				GasLimit: 21000,
			},
			wantErr: true,
			errMsg:  "to address cannot be nil",
		},
		{
			name: "negative value",
			params: TransactionParams{
				ChainID:  "ethereum",
				From:     from,
				To:       to,
				Value:    big.NewInt(-1000),
				Nonce:    nonce,
				GasLimit: 21000,
			},
			wantErr: true,
			errMsg:  "value cannot be negative",
		},
		{
			name: "nil value defaults to zero",
			params: TransactionParams{
				ChainID:  "ethereum",
				From:     from,
				To:       to,
				Value:    nil,
				Nonce:    nonce,
				GasLimit: 21000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := NewTransaction(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, tx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tx)
				assert.NotEmpty(t, tx.ID())
				assert.Equal(t, tt.params.ChainID, tx.ChainID())
				assert.Equal(t, TxStatusPending, tx.Status())
				assert.NotNil(t, tx.Metadata())
			}
		})
	}
}

func TestTransaction_SetHash(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	hash, _ := valueobjects.NewHash("0xabcdef")
	err := tx.SetHash(hash)
	require.NoError(t, err)
	assert.Equal(t, hash, tx.Hash())

	err = tx.SetHash(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash cannot be nil")
}

func TestTransaction_SetSignature(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	sig, _ := valueobjects.NewSignature("0xabcdef")
	err := tx.SetSignature(sig)
	require.NoError(t, err)
	assert.Equal(t, sig, tx.Signature())

	err = tx.SetSignature(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature cannot be nil")
}

func TestTransaction_UpdateStatus(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	assert.Equal(t, TxStatusPending, tx.Status())

	oldUpdatedAt := tx.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	tx.UpdateStatus(TxStatusConfirmed)
	assert.Equal(t, TxStatusConfirmed, tx.Status())
	assert.True(t, tx.UpdatedAt().After(oldUpdatedAt))
}

func TestTransaction_SetBlockNumber(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	tx.SetBlockNumber(12345)
	assert.Equal(t, uint64(12345), tx.BlockNumber())
}

func TestTransaction_SetConfirmations(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	tx.SetConfirmations(10)
	assert.Equal(t, uint64(10), tx.Confirmations())
}

func TestTransaction_SetMetadata(t *testing.T) {
	from, _ := valueobjects.NewAddress("0xfrom", "ethereum")
	to, _ := valueobjects.NewAddress("0xto", "ethereum")
	tx, _ := NewTransaction(TransactionParams{
		ChainID: "ethereum",
		From:    from,
		To:      to,
		Value:   big.NewInt(1000),
	})

	tx.SetMetadata("key", "value")
	assert.Equal(t, "value", tx.Metadata()["key"])
}

func TestNewWallet(t *testing.T) {
	addr, _ := valueobjects.NewAddress("0xwallet", "ethereum")

	tests := []struct {
		name    string
		address *valueobjects.Address
		chainID string
		label   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid wallet",
			address: addr,
			chainID: "ethereum",
			label:   "My Wallet",
			wantErr: false,
		},
		{
			name:    "nil address",
			address: nil,
			chainID: "ethereum",
			label:   "My Wallet",
			wantErr: true,
			errMsg:  "address cannot be nil",
		},
		{
			name:    "empty chain ID",
			address: addr,
			chainID: "",
			label:   "My Wallet",
			wantErr: true,
			errMsg:  "chain ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := NewWallet(tt.address, tt.chainID, tt.label)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, wallet)
			} else {
				require.NoError(t, err)
				require.NotNil(t, wallet)
				assert.NotEmpty(t, wallet.ID())
				assert.Equal(t, tt.address, wallet.Address())
				assert.Equal(t, tt.chainID, wallet.ChainID())
				assert.Equal(t, tt.label, wallet.Label())
			}
		})
	}
}

func TestWallet_UpdateLabel(t *testing.T) {
	addr, _ := valueobjects.NewAddress("0xwallet", "ethereum")
	wallet, _ := NewWallet(addr, "ethereum", "Old Label")

	wallet.UpdateLabel("New Label")
	assert.Equal(t, "New Label", wallet.Label())
}

func TestWallet_SetMetadata(t *testing.T) {
	addr, _ := valueobjects.NewAddress("0xwallet", "ethereum")
	wallet, _ := NewWallet(addr, "ethereum", "My Wallet")

	wallet.SetMetadata("key", "value")
	assert.Equal(t, "value", wallet.Metadata()["key"])
}

func TestNewFee(t *testing.T) {
	tests := []struct {
		name     string
		gasLimit uint64
		gasPrice *big.Int
		currency string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid fee",
			gasLimit: 21000,
			gasPrice: big.NewInt(20000000000),
			currency: "ETH",
			wantErr:  false,
		},
		{
			name:     "zero gas limit",
			gasLimit: 0,
			gasPrice: big.NewInt(20000000000),
			currency: "ETH",
			wantErr:  true,
			errMsg:   "gas limit cannot be zero",
		},
		{
			name:     "nil gas price",
			gasLimit: 21000,
			gasPrice: nil,
			currency: "ETH",
			wantErr:  true,
			errMsg:   "gas price must be positive",
		},
		{
			name:     "zero gas price",
			gasLimit: 21000,
			gasPrice: big.NewInt(0),
			currency: "ETH",
			wantErr:  true,
			errMsg:   "gas price must be positive",
		},
		{
			name:     "negative gas price",
			gasLimit: 21000,
			gasPrice: big.NewInt(-1),
			currency: "ETH",
			wantErr:  true,
			errMsg:   "gas price must be positive",
		},
		{
			name:     "empty currency",
			gasLimit: 21000,
			gasPrice: big.NewInt(20000000000),
			currency: "",
			wantErr:  true,
			errMsg:   "currency cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, err := NewFee(tt.gasLimit, tt.gasPrice, tt.currency)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, fee)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fee)
				assert.Equal(t, tt.gasLimit, fee.GasLimit())
				assert.Equal(t, tt.currency, fee.Currency())
				assert.NotNil(t, fee.Total())
			}
		})
	}
}

func TestNewEIP1559Fee(t *testing.T) {
	tests := []struct {
		name           string
		gasLimit       uint64
		maxFeePerGas   *big.Int
		maxPriorityFee *big.Int
		currency       string
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid EIP-1559 fee",
			gasLimit:       21000,
			maxFeePerGas:   big.NewInt(30000000000),
			maxPriorityFee: big.NewInt(2000000000),
			currency:       "ETH",
			wantErr:        false,
		},
		{
			name:           "zero gas limit",
			gasLimit:       0,
			maxFeePerGas:   big.NewInt(30000000000),
			maxPriorityFee: big.NewInt(2000000000),
			currency:       "ETH",
			wantErr:        true,
			errMsg:         "gas limit cannot be zero",
		},
		{
			name:           "nil max fee per gas",
			gasLimit:       21000,
			maxFeePerGas:   nil,
			maxPriorityFee: big.NewInt(2000000000),
			currency:       "ETH",
			wantErr:        true,
			errMsg:         "max fee per gas must be positive",
		},
		{
			name:           "nil max priority fee",
			gasLimit:       21000,
			maxFeePerGas:   big.NewInt(30000000000),
			maxPriorityFee: nil,
			currency:       "ETH",
			wantErr:        true,
			errMsg:         "max priority fee cannot be negative",
		},
		{
			name:           "negative max priority fee",
			gasLimit:       21000,
			maxFeePerGas:   big.NewInt(30000000000),
			maxPriorityFee: big.NewInt(-1),
			currency:       "ETH",
			wantErr:        true,
			errMsg:         "max priority fee cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, err := NewEIP1559Fee(tt.gasLimit, tt.maxFeePerGas, tt.maxPriorityFee, tt.currency)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, fee)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fee)
				assert.Equal(t, tt.gasLimit, fee.GasLimit())
				assert.Equal(t, tt.currency, fee.Currency())
				assert.NotNil(t, fee.Total())
			}
		})
	}
}

func TestNewNetwork(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		netName string
		rpcURL  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid network",
			chainID: "ethereum",
			netName: "Ethereum Mainnet",
			rpcURL:  "https://eth.llamarpc.com",
			wantErr: false,
		},
		{
			name:    "empty chain ID",
			chainID: "",
			netName: "Ethereum Mainnet",
			rpcURL:  "https://eth.llamarpc.com",
			wantErr: true,
			errMsg:  "chain ID cannot be empty",
		},
		{
			name:    "empty name",
			chainID: "ethereum",
			netName: "",
			rpcURL:  "https://eth.llamarpc.com",
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "empty RPC URL",
			chainID: "ethereum",
			netName: "Ethereum Mainnet",
			rpcURL:  "",
			wantErr: true,
			errMsg:  "RPC URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network, err := NewNetwork(tt.chainID, tt.netName, tt.rpcURL)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, network)
			} else {
				require.NoError(t, err)
				require.NotNil(t, network)
				assert.NotEmpty(t, network.ID())
				assert.Equal(t, tt.chainID, network.ChainID())
				assert.Equal(t, tt.netName, network.Name())
				assert.Equal(t, tt.rpcURL, network.RPCURL())
				assert.True(t, network.IsActive())
			}
		})
	}
}

func TestNetwork_SetExplorerURL(t *testing.T) {
	network, _ := NewNetwork("ethereum", "Ethereum Mainnet", "https://eth.llamarpc.com")

	network.SetExplorerURL("https://etherscan.io")
	assert.Equal(t, "https://etherscan.io", network.ExplorerURL())
}

func TestNetwork_ActivateDeactivate(t *testing.T) {
	network, _ := NewNetwork("ethereum", "Ethereum Mainnet", "https://eth.llamarpc.com")

	assert.True(t, network.IsActive())

	network.Deactivate()
	assert.False(t, network.IsActive())

	network.Activate()
	assert.True(t, network.IsActive())
}
