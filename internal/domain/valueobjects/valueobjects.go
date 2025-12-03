package valueobjects

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Address represents a blockchain address
type Address struct {
	value string
	chain string
}

// NewAddress creates a new Address value object
func NewAddress(value, chain string) (*Address, error) {
	if value == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}

	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, fmt.Errorf("address cannot be empty after normalization")
	}

	return &Address{
		value: normalized,
		chain: chain,
	}, nil
}

// Value returns the address value
func (a *Address) Value() string {
	return a.value
}

// Chain returns the chain identifier
func (a *Address) Chain() string {
	return a.chain
}

// String returns string representation
func (a *Address) String() string {
	return a.value
}

// Equals checks if two addresses are equal
func (a *Address) Equals(other *Address) bool {
	if other == nil {
		return false
	}
	return a.value == other.value && a.chain == other.chain
}

// Hash represents a transaction or block hash
type Hash struct {
	value []byte
}

// NewHash creates a new Hash from hex string
func NewHash(hexStr string) (*Hash, error) {
	if hexStr == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	cleaned := strings.TrimPrefix(hexStr, "0x")
	cleaned = strings.TrimSpace(cleaned)

	bytes, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %w", err)
	}

	if len(bytes) == 0 {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	return &Hash{value: bytes}, nil
}

// NewHashFromBytes creates a new Hash from bytes
func NewHashFromBytes(data []byte) (*Hash, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("hash bytes cannot be empty")
	}

	copied := make([]byte, len(data))
	copy(copied, data)

	return &Hash{value: copied}, nil
}

// Hex returns hex string representation with 0x prefix
func (h *Hash) Hex() string {
	return "0x" + hex.EncodeToString(h.value)
}

// HexWithoutPrefix returns hex string without 0x prefix
func (h *Hash) HexWithoutPrefix() string {
	return hex.EncodeToString(h.value)
}

// Bytes returns the underlying bytes (copy)
func (h *Hash) Bytes() []byte {
	result := make([]byte, len(h.value))
	copy(result, h.value)
	return result
}

// String returns string representation
func (h *Hash) String() string {
	return h.Hex()
}

// Equals checks if two hashes are equal
func (h *Hash) Equals(other *Hash) bool {
	if other == nil {
		return false
	}
	if len(h.value) != len(other.value) {
		return false
	}
	for i := range h.value {
		if h.value[i] != other.value[i] {
			return false
		}
	}
	return true
}

// Signature represents a cryptographic signature
type Signature struct {
	value []byte
}

// NewSignature creates a new Signature from hex string
func NewSignature(hexStr string) (*Signature, error) {
	if hexStr == "" {
		return nil, fmt.Errorf("signature cannot be empty")
	}

	cleaned := strings.TrimPrefix(hexStr, "0x")
	cleaned = strings.TrimSpace(cleaned)

	bytes, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("invalid signature format: %w", err)
	}

	if len(bytes) == 0 {
		return nil, fmt.Errorf("signature cannot be empty")
	}

	return &Signature{value: bytes}, nil
}

// NewSignatureFromBytes creates a new Signature from bytes
func NewSignatureFromBytes(data []byte) (*Signature, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("signature bytes cannot be empty")
	}

	copied := make([]byte, len(data))
	copy(copied, data)

	return &Signature{value: copied}, nil
}

// Hex returns hex string representation with 0x prefix
func (s *Signature) Hex() string {
	return "0x" + hex.EncodeToString(s.value)
}

// HexWithoutPrefix returns hex string without 0x prefix
func (s *Signature) HexWithoutPrefix() string {
	return hex.EncodeToString(s.value)
}

// Bytes returns the underlying bytes (copy)
func (s *Signature) Bytes() []byte {
	result := make([]byte, len(s.value))
	copy(result, s.value)
	return result
}

// String returns string representation
func (s *Signature) String() string {
	return s.Hex()
}

// Nonce represents a transaction nonce
type Nonce struct {
	value uint64
}

// NewNonce creates a new Nonce
func NewNonce(value uint64) *Nonce {
	return &Nonce{value: value}
}

// Value returns the nonce value
func (n *Nonce) Value() uint64 {
	return n.value
}

// Increment returns a new Nonce with incremented value
func (n *Nonce) Increment() *Nonce {
	return &Nonce{value: n.value + 1}
}

// String returns string representation
func (n *Nonce) String() string {
	return fmt.Sprintf("%d", n.value)
}

// Equals checks if two nonces are equal
func (n *Nonce) Equals(other *Nonce) bool {
	if other == nil {
		return false
	}
	return n.value == other.value
}
