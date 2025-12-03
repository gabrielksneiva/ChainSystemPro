package crypto

import (
	"fmt"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a new BIP39 mnemonic phrase
// Supported bit sizes: 128 (12 words), 256 (24 words)
func GenerateMnemonic(bitSize int) (string, error) {
	if bitSize != 128 && bitSize != 256 {
		return "", fmt.Errorf("invalid bit size: %d (must be 128 or 256)", bitSize)
	}

	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic checks if a mnemonic phrase is valid according to BIP39
func ValidateMnemonic(mnemonic string) error {
	if mnemonic == "" {
		return fmt.Errorf("mnemonic cannot be empty")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic")
	}

	return nil
}

// MnemonicToSeed converts a mnemonic phrase to a seed
// The passphrase can be empty string for no passphrase
func MnemonicToSeed(mnemonic, passphrase string) ([]byte, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, err
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	return seed, nil
}

// SplitMnemonic splits a mnemonic phrase into individual words
func SplitMnemonic(mnemonic string) []string {
	return strings.Fields(mnemonic)
}
