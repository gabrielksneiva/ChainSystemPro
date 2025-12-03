package examples

import (
	"fmt"
	"log"

	core "github.com/gabrielksneiva/ChainSystemPro/internal/adapters/bitcoin/core"
	"github.com/gabrielksneiva/ChainSystemPro/pkg/crypto"
)

func main() {
	fmt.Println("=== ChainSystemPro - Bitcoin HD Wallet Demo ===")

	// Generate a new mnemonic
	mnemonic, err := crypto.GenerateMnemonic(256)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated Mnemonic (24 words):")
	fmt.Println(mnemonic)
	fmt.Println()

	// Create HD Wallet from mnemonic
	wallet, err := crypto.NewHDWallet(mnemonic, "")
	if err != nil {
		log.Fatal(err)
	}

	// Demonstrate different address types
	addressTypes := []struct {
		name        string
		cryptoType  crypto.AddressType
		bitcoinType core.AddressType
		network     core.Network
	}{
		{"P2PKH (Legacy)", crypto.AddressTypeP2PKH, core.P2PKH, core.Mainnet},
		{"P2SH (Nested SegWit)", crypto.AddressTypeP2SH, core.P2SH, core.Mainnet},
		{"P2WPKH (Native SegWit)", crypto.AddressTypeP2WPKH, core.P2WPKH, core.Mainnet},
	}

	for _, at := range addressTypes {
		fmt.Printf("=== %s ===\n", at.name)

		// Derive account (m/purpose'/0'/0')
		account, err := wallet.DeriveAccount(0, 0, at.cryptoType)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Account Path: %s\n", account.Path())

		// Generate first 3 receiving addresses
		fmt.Println("\nReceiving Addresses:")
		for i := uint32(0); i < 3; i++ {
			addr, err := account.DeriveAddress(0, i)
			if err != nil {
				log.Fatal(err)
			}
			btcAddr, err := core.PubKeyToAddress(addr.PublicKey(), at.bitcoinType, at.network)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %s: %s\n", addr.Path(), btcAddr)
		}

		// Generate first change address
		fmt.Println("\nChange Address:")
		changeAddr, err := account.DeriveAddress(1, 0)
		if err != nil {
			log.Fatal(err)
		}
		btcChangeAddr, err := core.PubKeyToAddress(changeAddr.PublicKey(), at.bitcoinType, at.network)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %s: %s\n", changeAddr.Path(), btcChangeAddr)
		fmt.Println()
	}

	// Test with testnet
	fmt.Println("=== Testnet Addresses ===")
	account, err := wallet.DeriveAccount(1, 0, crypto.AddressTypeP2PKH) // coinType 1 = testnet
	if err != nil {
		log.Fatal(err)
	}
	addr, err := account.DeriveAddress(0, 0)
	if err != nil {
		log.Fatal(err)
	}
	testnetAddr, err := core.PubKeyToAddress(addr.PublicKey(), core.P2PKH, core.Testnet)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Testnet P2PKH: %s\n", testnetAddr)
}
