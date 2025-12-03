package bitcoin

import (
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ripemd160"
)

func TestSignP2PKH(t *testing.T) {
	// Create a simple tx with one input and one output
	tx := wire.NewMsgTx(2)
	// prevTxHash not used in this minimal test; rely on index only
	txin := wire.NewTxIn(&wire.OutPoint{Hash: [32]byte{}, Index: 0}, nil, nil)
	tx.AddTxIn(txin)
	// Generate a random private key (for test only)
	priv, _ := btcec.NewPrivateKey()

	// Build a P2PKH pkScript manually using hash160(pubkey)
	pubkey := priv.PubKey().SerializeCompressed()
	sha := sha256.Sum256(pubkey)
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	pubKeyHash := ripemd.Sum(nil)
	pkScript, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	txout := wire.NewTxOut(50000000, pkScript)
	tx.AddTxOut(txout)

	// Sign input
	sigScript, err := SignP2PKH(tx, 0, pkScript, priv)
	require.NoError(t, err)
	assert.NotEmpty(t, sigScript)

	// Attach scriptSig
	AttachScriptSig(tx, 0, sigScript)
	assert.NotEmpty(t, tx.TxIn[0].SignatureScript)
}
