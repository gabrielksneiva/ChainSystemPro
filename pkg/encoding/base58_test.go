package encoding_test

import (
	"bytes"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeBase58Check(t *testing.T) {
	tests := []struct {
		name    string
		version byte
		payload []byte
	}{
		{
			name:    "P2PKH mainnet address",
			version: 0x00,
			payload: bytes.Repeat([]byte{0x01}, 20),
		},
		{
			name:    "P2PKH testnet address",
			version: 0x6F,
			payload: bytes.Repeat([]byte{0x02}, 20),
		},
		{
			name:    "P2SH mainnet address",
			version: 0x05,
			payload: bytes.Repeat([]byte{0x03}, 20),
		},
		{
			name:    "empty payload",
			version: 0x00,
			payload: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded := encoding.EncodeBase58Check(tt.version, tt.payload)
			assert.NotEmpty(t, encoded)

			// Decode
			version, payload, err := encoding.DecodeBase58Check(encoded)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.version, version)
			assert.Equal(t, tt.payload, payload)
		})
	}
}

func TestDecodeBase58Check_InvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{
			name:    "too short",
			encoded: "1",
			wantErr: true,
		},
		{
			name:    "invalid checksum - corrupted",
			encoded: "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvX", // changed last char
			wantErr: true,
		},
		{
			name:    "empty string",
			encoded: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := encoding.DecodeBase58Check(tt.encoded)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBase58CheckRoundTrip(t *testing.T) {
	// Test with known Bitcoin address
	payload := []byte{
		0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4,
		0x54, 0x94, 0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23,
		0xf1, 0x43, 0x3b, 0xd6,
	}
	version := byte(0x00)

	encoded := encoding.EncodeBase58Check(version, payload)
	decodedVersion, decodedPayload, err := encoding.DecodeBase58Check(encoded)

	require.NoError(t, err)
	assert.Equal(t, version, decodedVersion)
	assert.Equal(t, payload, decodedPayload)
}
