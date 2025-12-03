package core

import (
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubKeyToAddress_P2PKH(t *testing.T) {
	// Test with known public key and expected address
	// This is the first address from the standard test mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	account, err := wallet.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	require.NoError(t, err)

	addr, err := account.DeriveAddress(0, 0)
	require.NoError(t, err)

	pubKey := addr.PublicKey()

	tests := []struct {
		name    string
		network Network
		wantErr bool
	}{
		{
			name:    "mainnet P2PKH",
			network: Mainnet,
			wantErr: false,
		},
		{
			name:    "testnet P2PKH",
			network: Testnet,
			wantErr: false,
		},
		{
			name:    "regtest P2PKH",
			network: Regtest,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := PubKeyToAddress(pubKey, P2PKH, tt.network)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, address)

				// Mainnet addresses start with '1'
				if tt.network == Mainnet {
					assert.Equal(t, byte('1'), address[0])
				}
				// Testnet addresses start with 'm' or 'n'
				if tt.network == Testnet || tt.network == Regtest {
					first := address[0]
					assert.True(t, first == 'm' || first == 'n')
				}
			}
		})
	}
}

func TestPubKeyToAddress_P2SH(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	account, err := wallet.DeriveAccount(0, 0, crypto.AddressTypeP2SH)
	require.NoError(t, err)

	addr, err := account.DeriveAddress(0, 0)
	require.NoError(t, err)

	pubKey := addr.PublicKey()

	tests := []struct {
		name    string
		network Network
	}{
		{
			name:    "mainnet P2SH",
			network: Mainnet,
		},
		{
			name:    "testnet P2SH",
			network: Testnet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := PubKeyToAddress(pubKey, P2SH, tt.network)

			require.NoError(t, err)
			assert.NotEmpty(t, address)

			// Mainnet P2SH addresses start with '3'
			if tt.network == Mainnet {
				assert.Equal(t, byte('3'), address[0])
			}
			// Testnet P2SH addresses start with '2'
			if tt.network == Testnet || tt.network == Regtest {
				assert.Equal(t, byte('2'), address[0])
			}
		})
	}
}

func TestPubKeyToAddress_P2WPKH(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	account, err := wallet.DeriveAccount(0, 0, crypto.AddressTypeP2WPKH)
	require.NoError(t, err)

	addr, err := account.DeriveAddress(0, 0)
	require.NoError(t, err)

	pubKey := addr.PublicKey()

	tests := []struct {
		name       string
		network    Network
		wantPrefix string
	}{
		{
			name:       "mainnet P2WPKH",
			network:    Mainnet,
			wantPrefix: "bc1",
		},
		{
			name:       "testnet P2WPKH",
			network:    Testnet,
			wantPrefix: "tb1",
		},
		{
			name:       "regtest P2WPKH",
			network:    Regtest,
			wantPrefix: "bcrt1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := PubKeyToAddress(pubKey, P2WPKH, tt.network)

			require.NoError(t, err)
			assert.NotEmpty(t, address)
			assert.True(t, len(address) >= len(tt.wantPrefix))
			assert.Equal(t, tt.wantPrefix, address[:len(tt.wantPrefix)])
		})
	}
}

func TestPubKeyToAddress_InvalidInputs(t *testing.T) {
	pubKey := make([]byte, 33) // Valid compressed public key length

	tests := []struct {
		name     string
		pubKey   []byte
		addrType AddressType
		network  Network
		wantErr  bool
	}{
		{
			name:     "unsupported address type",
			pubKey:   pubKey,
			addrType: "INVALID",
			network:  Mainnet,
			wantErr:  true,
		},
		{
			name:     "unsupported network P2PKH",
			pubKey:   pubKey,
			addrType: P2PKH,
			network:  "invalid",
			wantErr:  true,
		},
		{
			name:     "unsupported network P2SH",
			pubKey:   pubKey,
			addrType: P2SH,
			network:  "invalid",
			wantErr:  true,
		},
		{
			name:     "unsupported network P2WPKH",
			pubKey:   pubKey,
			addrType: P2WPKH,
			network:  "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PubKeyToAddress(tt.pubKey, tt.addrType, tt.network)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeterministicAddresses(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	// Generate address twice with same params
	wallet1, _ := crypto.NewHDWallet(mnemonic, "")
	account1, _ := wallet1.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	addr1, _ := account1.DeriveAddress(0, 0)
	address1, _ := PubKeyToAddress(addr1.PublicKey(), P2PKH, Mainnet)

	wallet2, _ := crypto.NewHDWallet(mnemonic, "")
	account2, _ := wallet2.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	addr2, _ := account2.DeriveAddress(0, 0)
	address2, _ := PubKeyToAddress(addr2.PublicKey(), P2PKH, Mainnet)

	assert.Equal(t, address1, address2)
}

func TestAllAddressTypes(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	// Generate addresses for all types
	addressTypes := []struct {
		cryptoType  crypto.AddressType
		bitcoinType AddressType
	}{
		{crypto.AddressTypeP2PKH, P2PKH},
		{crypto.AddressTypeP2SH, P2SH},
		{crypto.AddressTypeP2WPKH, P2WPKH},
	}

	for _, at := range addressTypes {
		t.Run(string(at.bitcoinType), func(t *testing.T) {
			account, err := wallet.DeriveAccount(0, 0, at.cryptoType)
			require.NoError(t, err)

			addr, err := account.DeriveAddress(0, 0)
			require.NoError(t, err)

			address, err := PubKeyToAddress(addr.PublicKey(), at.bitcoinType, Mainnet)
			require.NoError(t, err)
			assert.NotEmpty(t, address)

			t.Logf("%s address: %s", at.bitcoinType, address)
		})
	}
}
