package encoding

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
)

// Base58Check encoding/decoding for Bitcoin addresses

// EncodeBase58Check encodes data with Base58Check encoding
func EncodeBase58Check(version byte, payload []byte) string {
	// Prepend version byte
	versionedPayload := append([]byte{version}, payload...)

	// Calculate checksum (double SHA256)
	checksum := doubleSHA256(versionedPayload)[:4]

	// Append checksum
	fullPayload := append(versionedPayload, checksum...)

	// Encode to base58
	return base58.Encode(fullPayload)
}

// DecodeBase58Check decodes a Base58Check encoded string
func DecodeBase58Check(encoded string) (version byte, payload []byte, err error) {
	decoded := base58.Decode(encoded)

	if len(decoded) < 5 {
		return 0, nil, errors.New("invalid base58check encoding: too short")
	}

	// Extract components
	version = decoded[0]
	payload = decoded[1 : len(decoded)-4]
	checksumReceived := decoded[len(decoded)-4:]

	// Verify checksum
	checksumCalculated := doubleSHA256(decoded[:len(decoded)-4])[:4]

	for i := 0; i < 4; i++ {
		if checksumReceived[i] != checksumCalculated[i] {
			return 0, nil, fmt.Errorf("invalid checksum")
		}
	}

	return version, payload, nil
}

// doubleSHA256 performs double SHA256 hashing
func doubleSHA256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}
