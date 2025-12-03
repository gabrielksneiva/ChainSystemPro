package core

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/gabrielksneiva/ChainSystemPro/internal/encoding"
	"golang.org/x/crypto/ripemd160"
)

type Network string

const (
	Mainnet Network = "mainnet"
	Testnet Network = "testnet"
	Regtest Network = "regtest"
)

type AddressType string

const (
	P2PKH  AddressType = "P2PKH"
	P2SH   AddressType = "P2SH"
	P2WPKH AddressType = "P2WPKH"
)

const (
	P2PKHVersionMainnet byte = 0x00
	P2PKHVersionTestnet byte = 0x6F
	P2SHVersionMainnet  byte = 0x05
	P2SHVersionTestnet  byte = 0xC4
)

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

func pubKeyToP2PKH(pubKey []byte, network Network) (string, error) {
	pubKeyHash := hash160(pubKey)
	var version byte
	switch network {
	case Mainnet:
		version = P2PKHVersionMainnet
	case Testnet, Regtest:
		version = P2PKHVersionTestnet
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}
	address := encoding.EncodeBase58Check(version, pubKeyHash)
	return address, nil
}

func pubKeyToP2SH(pubKey []byte, network Network) (string, error) {
	pubKeyHash := hash160(pubKey)
	witnessProgram := append([]byte{0x00, 0x14}, pubKeyHash...)
	scriptHash := hash160(witnessProgram)
	var version byte
	switch network {
	case Mainnet:
		version = P2SHVersionMainnet
	case Testnet, Regtest:
		version = P2SHVersionTestnet
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}
	address := encoding.EncodeBase58Check(version, scriptHash)
	return address, nil
}

func pubKeyToP2WPKH(pubKey []byte, network Network) (string, error) {
	pubKeyHash := hash160(pubKey)
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
	converted, err := bech32.ConvertBits(pubKeyHash, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert bits: %w", err)
	}
	data := append([]byte{0}, converted...)
	address, err := bech32.Encode(hrp, data)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32: %w", err)
	}
	return address, nil
}

func hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	return ripemd.Sum(nil)
}
