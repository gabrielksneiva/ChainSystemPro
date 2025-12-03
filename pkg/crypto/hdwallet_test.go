package crypto_test

import (
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHDWallet(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	tests := []struct {
		name       string
		mnemonic   string
		passphrase string
		wantErr    bool
	}{
		{
			name:       "valid mnemonic without passphrase",
			mnemonic:   mnemonic,
			passphrase: "",
			wantErr:    false,
		},
		{
			name:       "valid mnemonic with passphrase",
			mnemonic:   mnemonic,
			passphrase: "test",
			wantErr:    false,
		},
		{
			name:       "invalid mnemonic",
			mnemonic:   "invalid mnemonic",
			passphrase: "",
			wantErr:    true,
		},
		{
			name:       "empty mnemonic",
			mnemonic:   "",
			passphrase: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := crypto.NewHDWallet(tt.mnemonic, tt.passphrase)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, wallet)
			}
		})
	}
}

func TestHDWallet_DerivePath(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid BIP44 Bitcoin path",
			path:    "m/44'/0'/0'/0/0",
			wantErr: false,
		},
		{
			name:    "valid BIP49 Bitcoin path",
			path:    "m/49'/0'/0'/0/0",
			wantErr: false,
		},
		{
			name:    "valid BIP84 Bitcoin path",
			path:    "m/84'/0'/0'/0/0",
			wantErr: false,
		},
		{
			name:    "master key only",
			path:    "m/",
			wantErr: false,
		},
		{
			name:    "invalid path missing m",
			path:    "44'/0'/0'/0/0",
			wantErr: true,
		},
		{
			name:    "invalid path malformed",
			path:    "m/invalid",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with empty segment",
			path:    "m/44'//0'/0/0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := wallet.DerivePath(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, key)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestHDWallet_DeriveAccount(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	tests := []struct {
		name        string
		coinType    uint32
		account     uint32
		addressType crypto.AddressType
		wantPath    string
	}{
		{
			name:        "Bitcoin mainnet BIP44 account 0",
			coinType:    0,
			account:     0,
			addressType: crypto.AddressTypeP2PKH,
			wantPath:    "m/44'/0'/0'",
		},
		{
			name:        "Bitcoin testnet BIP44 account 1",
			coinType:    1,
			account:     1,
			addressType: crypto.AddressTypeP2PKH,
			wantPath:    "m/44'/1'/1'",
		},
		{
			name:        "Bitcoin mainnet BIP49 account 0",
			coinType:    0,
			account:     0,
			addressType: crypto.AddressTypeP2SH,
			wantPath:    "m/49'/0'/0'",
		},
		{
			name:        "Bitcoin mainnet BIP84 account 0",
			coinType:    0,
			account:     0,
			addressType: crypto.AddressTypeP2WPKH,
			wantPath:    "m/84'/0'/0'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := wallet.DeriveAccount(tt.coinType, tt.account, tt.addressType)

			require.NoError(t, err)
			assert.NotNil(t, account)
			assert.Equal(t, tt.wantPath, account.Path())
		})
	}
}

func TestHDAccount_DeriveAddress(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	account, err := wallet.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	require.NoError(t, err)

	tests := []struct {
		name     string
		change   uint32
		index    uint32
		wantPath string
	}{
		{
			name:     "first receiving address",
			change:   0,
			index:    0,
			wantPath: "m/44'/0'/0'/0/0",
		},
		{
			name:     "second receiving address",
			change:   0,
			index:    1,
			wantPath: "m/44'/0'/0'/0/1",
		},
		{
			name:     "first change address",
			change:   1,
			index:    0,
			wantPath: "m/44'/0'/0'/1/0",
		},
		{
			name:     "100th receiving address",
			change:   0,
			index:    99,
			wantPath: "m/44'/0'/0'/0/99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := account.DeriveAddress(tt.change, tt.index)

			require.NoError(t, err)
			assert.NotNil(t, addr)
			assert.Equal(t, tt.wantPath, addr.Path())
			assert.NotEmpty(t, addr.PublicKey())
			assert.NotEmpty(t, addr.PrivateKey())
		})
	}
}

func TestDeterministicAddressGeneration(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	wallet1, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)
	account1, err := wallet1.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	require.NoError(t, err)
	addr1, err := account1.DeriveAddress(0, 0)
	require.NoError(t, err)

	wallet2, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)
	account2, err := wallet2.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	require.NoError(t, err)
	addr2, err := account2.DeriveAddress(0, 0)
	require.NoError(t, err)

	assert.Equal(t, addr1.PublicKey(), addr2.PublicKey())
	assert.Equal(t, addr1.PrivateKey(), addr2.PrivateKey())
	assert.Equal(t, addr1.Path(), addr2.Path())
}

func TestHDAddress_Keys(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	require.NoError(t, err)

	account, err := wallet.DeriveAccount(0, 0, crypto.AddressTypeP2PKH)
	require.NoError(t, err)

	addr, err := account.DeriveAddress(0, 0)
	require.NoError(t, err)

	t.Run("public key not empty", func(t *testing.T) {
		assert.NotEmpty(t, addr.PublicKey())
	})

	t.Run("private key not empty", func(t *testing.T) {
		assert.NotEmpty(t, addr.PrivateKey())
	})

	t.Run("ECDSA private key valid", func(t *testing.T) {
		ecdsaPriv := addr.ECDSAPrivateKey()
		assert.NotNil(t, ecdsaPriv)
		assert.NotNil(t, ecdsaPriv.PublicKey)
	})

	t.Run("ECDSA public key valid", func(t *testing.T) {
		ecdsaPub := addr.ECDSAPublicKey()
		assert.NotNil(t, ecdsaPub)
		assert.NotNil(t, ecdsaPub.X)
		assert.NotNil(t, ecdsaPub.Y)
	})
}

func TestAddressType_Purpose(t *testing.T) {
	tests := []struct {
		addrType crypto.AddressType
		want     uint32
	}{
		{crypto.AddressTypeP2PKH, 44},
		{crypto.AddressTypeP2SH, 49},
		{crypto.AddressTypeP2WPKH, 84},
		{"unknown", 44}, // default case
	}

	for _, tt := range tests {
		t.Run(string(tt.addrType), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.addrType.Purpose())
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid BIP44 path",
			path:    "m/44'/0'/0'/0/0",
			wantErr: false,
		},
		{
			name:    "valid path with h notation",
			path:    "m/44h/0h/0h/0/0",
			wantErr: false,
		},
		{
			name:    "valid single level",
			path:    "m/0",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "missing m prefix",
			path:    "44'/0'/0'/0/0",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			path:    "m/44'/invalid/0'/0/0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crypto.ValidatePath(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
