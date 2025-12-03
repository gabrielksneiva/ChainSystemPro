package core

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SignP2PKH(tx *wire.MsgTx, idx int, prevPkScript []byte, priv *btcec.PrivateKey) ([]byte, error) {
	hash, err := txscript.CalcSignatureHash(prevPkScript, txscript.SigHashAll, tx, idx)
	if err != nil {
		return nil, fmt.Errorf("calc sig hash failed: %w", err)
	}
	if len(hash) != 32 {
		return nil, fmt.Errorf("unexpected sighash length: %d", len(hash))
	}
	sig := ecdsa.Sign(priv, hash)
	der := sig.Serialize()
	derWithHashType := append(der, byte(txscript.SigHashAll))
	pubkey := priv.PubKey().SerializeCompressed()
	scriptSig, err := txscript.NewScriptBuilder().AddData(derWithHashType).AddData(pubkey).Script()
	if err != nil {
		return nil, fmt.Errorf("build scriptSig failed: %w", err)
	}
	return scriptSig, nil
}

func AttachScriptSig(tx *wire.MsgTx, idx int, scriptSig []byte) {
	tx.TxIn[idx].SignatureScript = scriptSig
}

func HexToBytes(h string) ([]byte, error) {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return b, nil
}
