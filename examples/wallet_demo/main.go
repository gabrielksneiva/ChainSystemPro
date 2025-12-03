package main

import (
	"fmt"
	"log"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/bitcoin"
	"github.com/gabrielksneiva/ChainSystemPro/pkg/crypto"
)

func main() {
	fmt.Println("=== ChainSystemPro - Bitcoin HD Wallet Demo ===")

	mnemonic, err := crypto.GenerateMnemonic(256)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generated Mnemonic (24 words):")
	fmt.Println(mnemonic)
	fmt.Println()

	wallet, err := crypto.NewHDWallet(mnemonic, "")
	if err != nil {
		log.Fatal(err)
	}

	addressTypes := []struct {
		name        string
		cryptoType  crypto.AddressType
		bitcoinType bitcoin.AddressType
	}{
		{"P2PKH (Legacy)", crypto.AddressTypeP2PKH, bitcoin.P2PKH},
		{"P2SH (Nested SegWit)", crypto.AddressTypeP2SH, bitcoin.P2SH},
		{"P2WPKH (Native SegWit)", crypto.AddressTypeP2WPKH, bitcoin.P2WPKH},
	}

	for _, at := range addressTypes {
		fmt.Printf("=== %s ===\n", at.name)

		account, err := wallet.DeriveAccount(0, 0, at.cryptoType)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Account Path: %s\n\n", account.Path())

		fmt.Println("Receiving Addresses:")
		for i := uint32(0); i < 3; i++ {
			addr, err := account.DeriveAddress(0, i)
			if err != nil {
				log.Fatal(err)
			}

			btcAddr, err := bitcoin.PubKeyToAddress(addr.PublicKey(), at.bitcoinType, bitcoin.Mainnet)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("  %s: %s\n", addr.Path(), btcAddr)
		}
		fmt.Println()
	}
}
