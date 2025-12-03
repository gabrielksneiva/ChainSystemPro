package bitcoin

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// SignP2PKH signs a legacy P2PKH input given the previous output's pubKeyHash and returns scriptSig
func SignP2PKH(tx *wire.MsgTx, idx int, prevPkScript []byte, priv *btcec.PrivateKey) ([]byte, error) {
	// Calculate signature hash (double-SHA256 per SigHashAll)
	hash, err := txscript.CalcSignatureHash(prevPkScript, txscript.SigHashAll, tx, idx)
	if err != nil {
		return nil, fmt.Errorf("calc sig hash failed: %w", err)
	}

	// Sign the 32-byte digest using ECDSA
	if len(hash) != 32 {
		return nil, fmt.Errorf("unexpected sighash length: %d", len(hash))
	}
	sig := ecdsa.Sign(priv, hash)
	der := sig.Serialize()
	// Append sighash type
	derWithHashType := append(der, byte(txscript.SigHashAll))

	// Build scriptSig: <sig> <pubkey>
	pubkey := priv.PubKey().SerializeCompressed()
	scriptSig, err := txscript.NewScriptBuilder().AddData(derWithHashType).AddData(pubkey).Script()
	if err != nil {
		return nil, fmt.Errorf("build scriptSig failed: %w", err)
	}
	return scriptSig, nil
}

// AttachScriptSig sets the scriptSig for a tx input
func AttachScriptSig(tx *wire.MsgTx, idx int, scriptSig []byte) {
	// wire.TxIn has SignatureScript field
	tx.TxIn[idx].SignatureScript = scriptSig
}

// HexToBytes converts hex string to bytes
func HexToBytes(h string) ([]byte, error) {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return b, nil
}
