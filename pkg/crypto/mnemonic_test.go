package crypto_test

import (
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMnemonic(t *testing.T) {
	tests := []struct {
		name      string
		bitSize   int
		wantErr   bool
		wantWords int
	}{
		{
			name:      "128 bits generates 12 words",
			bitSize:   128,
			wantErr:   false,
			wantWords: 12,
		},
		{
			name:      "256 bits generates 24 words",
			bitSize:   256,
			wantErr:   false,
			wantWords: 24,
		},
		{
			name:    "invalid bit size",
			bitSize: 100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mnemonic, err := crypto.GenerateMnemonic(tt.bitSize)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, mnemonic)

			words := crypto.SplitMnemonic(mnemonic)
			assert.Equal(t, tt.wantWords, len(words))
		})
	}
}

func TestValidateMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		wantErr  bool
	}{
		{
			name:     "valid 12-word mnemonic",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			wantErr:  false,
		},
		{
			name:     "invalid mnemonic - wrong checksum",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
			wantErr:  true,
		},
		{
			name:     "invalid mnemonic - empty",
			mnemonic: "",
			wantErr:  true,
		},
		{
			name:     "invalid mnemonic - invalid word",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crypto.ValidateMnemonic(tt.mnemonic)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMnemonicToSeed(t *testing.T) {
	tests := []struct {
		name       string
		mnemonic   string
		passphrase string
		wantLen    int
	}{
		{
			name:       "mnemonic without passphrase",
			mnemonic:   "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			passphrase: "",
			wantLen:    64,
		},
		{
			name:       "mnemonic with passphrase",
			mnemonic:   "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			passphrase: "test",
			wantLen:    64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, err := crypto.MnemonicToSeed(tt.mnemonic, tt.passphrase)

			require.NoError(t, err)
			assert.Len(t, seed, tt.wantLen)
		})
	}
}

func TestDeterministicSeed(t *testing.T) {
	// Same mnemonic and passphrase should always generate the same seed
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	passphrase := "test"

	seed1, err := crypto.MnemonicToSeed(mnemonic, passphrase)
	require.NoError(t, err)

	seed2, err := crypto.MnemonicToSeed(mnemonic, passphrase)
	require.NoError(t, err)

	assert.Equal(t, seed1, seed2)
}

func TestDifferentPassphraseDifferentSeed(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	seed1, err := crypto.MnemonicToSeed(mnemonic, "")
	require.NoError(t, err)

	seed2, err := crypto.MnemonicToSeed(mnemonic, "password")
	require.NoError(t, err)

	assert.NotEqual(t, seed1, seed2)
}
