package bitcoin

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/gabrielksneiva/ChainSystemPro/pkg/encoding"
	"golang.org/x/crypto/ripemd160"
)

// Network represents Bitcoin network type
type Network string

const (
	Mainnet Network = "mainnet"
	Testnet Network = "testnet"
	Regtest Network = "regtest"
)

// AddressType represents the type of Bitcoin address
type AddressType string

const (
	P2PKH  AddressType = "P2PKH"  // Pay to Public Key Hash (Legacy)
	P2SH   AddressType = "P2SH"   // Pay to Script Hash (Nested SegWit)
	P2WPKH AddressType = "P2WPKH" // Pay to Witness Public Key Hash (Native SegWit)
)

// Address version bytes
const (
	P2PKHVersionMainnet byte = 0x00
	P2PKHVersionTestnet byte = 0x6F
	P2SHVersionMainnet  byte = 0x05
	P2SHVersionTestnet  byte = 0xC4
)

// PubKeyToAddress converts a public key to a Bitcoin address
func PubKeyToAddress(pubKey []byte, addrType AddressType, network Network) (string, error) {
	switch addrType {
	case P2PKH:
		return pubKeyToP2PKH(pubKey, network)
	case P2SH:
		return pubKeyToP2SH(pubKey, network)
	case P2WPKH:
		return pubKeyToP2WPKH(pubKey, network)
	default:
		return "", fmt.Errorf("unsupported address type: %s", addrType)
	}
}

// pubKeyToP2PKH creates a P2PKH (Legacy) address
func pubKeyToP2PKH(pubKey []byte, network Network) (string, error) {
	// Hash public key (SHA256 then RIPEMD160)
	pubKeyHash := hash160(pubKey)

	// Get version byte based on network
	var version byte
	switch network {
	case Mainnet:
		version = P2PKHVersionMainnet
	case Testnet, Regtest:
		version = P2PKHVersionTestnet
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	// Encode with Base58Check
	address := encoding.EncodeBase58Check(version, pubKeyHash)
	return address, nil
}

// pubKeyToP2SH creates a P2SH-P2WPKH (Nested SegWit) address
func pubKeyToP2SH(pubKey []byte, network Network) (string, error) {
	// Create witness program (OP_0 <20-byte-pubkey-hash>)
	pubKeyHash := hash160(pubKey)
	witnessProgram := append([]byte{0x00, 0x14}, pubKeyHash...)

	// Hash the witness program
	scriptHash := hash160(witnessProgram)

	// Get version byte based on network
	var version byte
	switch network {
	case Mainnet:
		version = P2SHVersionMainnet
	case Testnet, Regtest:
		version = P2SHVersionTestnet
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	// Encode with Base58Check
	address := encoding.EncodeBase58Check(version, scriptHash)
	return address, nil
}

// pubKeyToP2WPKH creates a P2WPKH (Native SegWit/Bech32) address
func pubKeyToP2WPKH(pubKey []byte, network Network) (string, error) {
	// Hash public key
	pubKeyHash := hash160(pubKey)

	// Get HRP (Human Readable Part) based on network
	var hrp string
	switch network {
	case Mainnet:
		hrp = "bc"
	case Testnet:
		hrp = "tb"
	case Regtest:
		hrp = "bcrt"
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	// Convert to 5-bit groups for bech32
	converted, err := bech32.ConvertBits(pubKeyHash, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert bits: %w", err)
	}

	// Prepend witness version (0)
	data := append([]byte{0}, converted...)

	// Encode with bech32
	address, err := bech32.Encode(hrp, data)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32: %w", err)
	}

	return address, nil
}

// hash160 performs SHA256 followed by RIPEMD160
func hash160(data []byte) []byte {
	// SHA256
	sha := sha256.Sum256(data)

	// RIPEMD160
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	return ripemd.Sum(nil)
}
