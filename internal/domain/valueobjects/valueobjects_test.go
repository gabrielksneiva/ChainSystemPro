package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddress(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		chain   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid address",
			value:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			chain:   "ethereum",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			chain:   "ethereum",
			wantErr: true,
			errMsg:  "address cannot be empty",
		},
		{
			name:    "empty chain",
			value:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			chain:   "",
			wantErr: true,
			errMsg:  "chain cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			chain:   "ethereum",
			wantErr: true,
			errMsg:  "address cannot be empty after normalization",
		},
		{
			name:    "address with spaces",
			value:   "  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb  ",
			chain:   "ethereum",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewAddress(tt.value, tt.chain)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, addr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, addr)
				assert.NotEmpty(t, addr.Value())
				assert.Equal(t, tt.chain, addr.Chain())
			}
		})
	}
}

func TestAddress_Methods(t *testing.T) {
	addr, err := NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "ethereum")
	require.NoError(t, err)

	assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", addr.Value())
	assert.Equal(t, "ethereum", addr.Chain())
	assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", addr.String())
}

func TestAddress_Equals(t *testing.T) {
	addr1, _ := NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "ethereum")
	addr2, _ := NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "ethereum")
	addr3, _ := NewAddress("0xOTHER", "ethereum")
	addr4, _ := NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "polygon")

	assert.True(t, addr1.Equals(addr2))
	assert.False(t, addr1.Equals(addr3))
	assert.False(t, addr1.Equals(addr4))
	assert.False(t, addr1.Equals(nil))
}

func TestNewHash(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid hash with 0x prefix",
			hexStr:  "0xabcdef1234567890",
			wantErr: false,
		},
		{
			name:    "valid hash without 0x prefix",
			hexStr:  "abcdef1234567890",
			wantErr: false,
		},
		{
			name:    "empty hash",
			hexStr:  "",
			wantErr: true,
			errMsg:  "hash cannot be empty",
		},
		{
			name:    "invalid hex",
			hexStr:  "0xGGGGGG",
			wantErr: true,
			errMsg:  "invalid hash format",
		},
		{
			name:    "hash with spaces gets trimmed",
			hexStr:  "  abcdef  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewHash(tt.hexStr)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, hash)
			} else {
				require.NoError(t, err)
				require.NotNil(t, hash)
				assert.NotEmpty(t, hash.Bytes())
			}
		})
	}
}

func TestNewHashFromBytes(t *testing.T) {
	data := []byte{0xab, 0xcd, 0xef}
	hash, err := NewHashFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, hash)
	assert.Equal(t, data, hash.Bytes())

	// Test empty bytes
	_, err = NewHashFromBytes([]byte{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash bytes cannot be empty")
}

func TestHash_Methods(t *testing.T) {
	hash, err := NewHash("0xabcdef")
	require.NoError(t, err)

	assert.Equal(t, "0xabcdef", hash.Hex())
	assert.Equal(t, "abcdef", hash.HexWithoutPrefix())
	assert.Equal(t, "0xabcdef", hash.String())
	assert.NotNil(t, hash.Bytes())
}

func TestHash_Equals(t *testing.T) {
	hash1, _ := NewHash("0xabcdef")
	hash2, _ := NewHash("0xabcdef")
	hash3, _ := NewHash("0x123456")

	assert.True(t, hash1.Equals(hash2))
	assert.False(t, hash1.Equals(hash3))
	assert.False(t, hash1.Equals(nil))
}

func TestNewSignature(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid signature with 0x prefix",
			hexStr:  "0xabcdef1234567890",
			wantErr: false,
		},
		{
			name:    "valid signature without 0x prefix",
			hexStr:  "abcdef1234567890",
			wantErr: false,
		},
		{
			name:    "empty signature",
			hexStr:  "",
			wantErr: true,
			errMsg:  "signature cannot be empty",
		},
		{
			name:    "invalid hex",
			hexStr:  "0xINVALID",
			wantErr: true,
			errMsg:  "invalid signature format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := NewSignature(tt.hexStr)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, sig)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sig)
				assert.NotEmpty(t, sig.Bytes())
			}
		})
	}
}

func TestNewSignatureFromBytes(t *testing.T) {
	data := []byte{0xab, 0xcd, 0xef}
	sig, err := NewSignatureFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, sig)
	assert.Equal(t, data, sig.Bytes())

	// Test empty bytes
	_, err = NewSignatureFromBytes([]byte{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature bytes cannot be empty")
}

func TestSignature_Methods(t *testing.T) {
	sig, err := NewSignature("0xabcdef")
	require.NoError(t, err)

	assert.Equal(t, "0xabcdef", sig.Hex())
	assert.Equal(t, "abcdef", sig.HexWithoutPrefix())
	assert.Equal(t, "0xabcdef", sig.String())
	assert.NotNil(t, sig.Bytes())
}

func TestNewNonce(t *testing.T) {
	nonce := NewNonce(42)
	require.NotNil(t, nonce)
	assert.Equal(t, uint64(42), nonce.Value())
}

func TestNonce_Increment(t *testing.T) {
	nonce := NewNonce(42)
	incremented := nonce.Increment()

	assert.Equal(t, uint64(42), nonce.Value())
	assert.Equal(t, uint64(43), incremented.Value())
}

func TestNonce_String(t *testing.T) {
	nonce := NewNonce(42)
	assert.Equal(t, "42", nonce.String())
}

func TestNonce_Equals(t *testing.T) {
	nonce1 := NewNonce(42)
	nonce2 := NewNonce(42)
	nonce3 := NewNonce(43)

	assert.True(t, nonce1.Equals(nonce2))
	assert.False(t, nonce1.Equals(nonce3))
	assert.False(t, nonce1.Equals(nil))
}
