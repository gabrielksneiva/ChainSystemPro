package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/tyler-smith/go-bip32"
)

// AddressType represents the type of Bitcoin address
type AddressType string

const (
	AddressTypeP2PKH  AddressType = "P2PKH"  // Legacy addresses (1...)
	AddressTypeP2SH   AddressType = "P2SH"   // Nested SegWit (3...)
	AddressTypeP2WPKH AddressType = "P2WPKH" // Native SegWit (bc1...)
)

// Purpose returns the BIP44 purpose field for the address type
func (at AddressType) Purpose() uint32 {
	switch at {
	case AddressTypeP2PKH:
		return 44
	case AddressTypeP2SH:
		return 49
	case AddressTypeP2WPKH:
		return 84
	default:
		return 44
	}
}

// HDWallet represents a hierarchical deterministic wallet
type HDWallet struct {
	masterKey *bip32.Key
	mnemonic  string
}

// HDAccount represents an account in the HD wallet (BIP44 account level)
type HDAccount struct {
	key         *bip32.Key
	path        string
	coinType    uint32
	account     uint32
	addressType AddressType
}

// HDAddress represents a derived address from an account
type HDAddress struct {
	key        *bip32.Key
	publicKey  *btcec.PublicKey
	privateKey *btcec.PrivateKey
	path       string
}

// NewHDWallet creates a new HD wallet from a mnemonic phrase
func NewHDWallet(mnemonic, passphrase string) (*HDWallet, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, err
	}

	seed, err := MnemonicToSeed(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	return &HDWallet{
		masterKey: masterKey,
		mnemonic:  mnemonic,
	}, nil
}

// DerivePath derives a key from a BIP32 path
// Path format: m/44'/0'/0'/0/0
func (w *HDWallet) DerivePath(path string) (*bip32.Key, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("path must start with 'm/'")
	}

	// Remove the 'm/' prefix
	path = strings.TrimPrefix(path, "m/")
	if path == "" {
		return w.masterKey, nil
	}

	segments := strings.Split(path, "/")
	key := w.masterKey

	for i, segment := range segments {
		if segment == "" {
			return nil, fmt.Errorf("invalid path: empty segment at position %d", i)
		}

		hardened := strings.HasSuffix(segment, "'")
		indexStr := strings.TrimSuffix(segment, "'")

		index, err := strconv.ParseUint(indexStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid index at position %d: %w", i, err)
		}

		var childIndex uint32
		if hardened {
			childIndex = uint32(index) + bip32.FirstHardenedChild
		} else {
			childIndex = uint32(index)
		}

		key, err = key.NewChildKey(childIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to derive child at position %d: %w", i, err)
		}
	}

	return key, nil
}

// DeriveAccount derives a BIP44 account
// m / purpose' / coin_type' / account'
func (w *HDWallet) DeriveAccount(coinType, account uint32, addressType AddressType) (*HDAccount, error) {
	purpose := addressType.Purpose()
	path := fmt.Sprintf("m/%d'/%d'/%d'", purpose, coinType, account)

	key, err := w.DerivePath(path)
	if err != nil {
		return nil, err
	}

	return &HDAccount{
		key:         key,
		path:        path,
		coinType:    coinType,
		account:     account,
		addressType: addressType,
	}, nil
}

// Path returns the derivation path of the account
func (a *HDAccount) Path() string {
	return a.path
}

// DeriveAddress derives an address from the account
// change: 0 for receiving addresses, 1 for change addresses
// index: address index
func (a *HDAccount) DeriveAddress(change, index uint32) (*HDAddress, error) {
	// Derive change level: m / purpose' / coin_type' / account' / change
	changeKey, err := a.key.NewChildKey(change)
	if err != nil {
		return nil, fmt.Errorf("failed to derive change key: %w", err)
	}

	// Derive address level: m / purpose' / coin_type' / account' / change / index
	addressKey, err := changeKey.NewChildKey(index)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address key: %w", err)
	}

	// Convert to ECDSA keys
	privKey, pubKey := btcec.PrivKeyFromBytes(addressKey.Key)

	path := fmt.Sprintf("%s/%d/%d", a.path, change, index)

	return &HDAddress{
		key:        addressKey,
		publicKey:  pubKey,
		privateKey: privKey,
		path:       path,
	}, nil
}

// Path returns the derivation path of the address
func (a *HDAddress) Path() string {
	return a.path
}

// PublicKey returns the public key as bytes
func (a *HDAddress) PublicKey() []byte {
	return a.publicKey.SerializeCompressed()
}

// PrivateKey returns the private key as bytes
func (a *HDAddress) PrivateKey() []byte {
	return a.privateKey.Serialize()
}

// ECDSAPrivateKey returns the ECDSA private key
func (a *HDAddress) ECDSAPrivateKey() *ecdsa.PrivateKey {
	return a.privateKey.ToECDSA()
}

// ECDSAPublicKey returns the ECDSA public key
func (a *HDAddress) ECDSAPublicKey() *ecdsa.PublicKey {
	ecdsaPubKey := a.privateKey.PubKey().ToECDSA()
	return ecdsaPubKey
}

// ValidatePath validates a BIP32 derivation path
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if !strings.HasPrefix(path, "m/") {
		return fmt.Errorf("path must start with 'm/'")
	}

	// Regular expression for valid BIP32 path
	// m/44'/0'/0'/0/0 or m/44h/0h/0h/0/0
	pathRegex := regexp.MustCompile(`^m(/\d+['h]?)*$`)
	if !pathRegex.MatchString(path) {
		return fmt.Errorf("invalid path format")
	}

	return nil
}
