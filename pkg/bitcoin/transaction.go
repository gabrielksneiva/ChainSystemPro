package bitcoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/pkg/encoding"
)

// TxInput represents a transaction input
type TxInput struct {
	PrevTxID  string
	PrevVout  uint32
	ScriptSig []byte
	Sequence  uint32
}

// TxOutput represents a transaction output
type TxOutput struct {
	Amount       uint64
	ScriptPubKey []byte
}

// Transaction represents a Bitcoin transaction
type Transaction struct {
	Version  uint32
	Inputs   []TxInput
	Outputs  []TxOutput
	Locktime uint32
}

// TransactionBuilder helps construct Bitcoin transactions
type TransactionBuilder struct {
	version  uint32
	inputs   []TxInput
	outputs  []TxOutput
	locktime uint32
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		version:  2,
		inputs:   []TxInput{},
		outputs:  []TxOutput{},
		locktime: 0,
	}
}

// AddInput adds an input to the transaction
func (tb *TransactionBuilder) AddInput(prevTxID string, prevVout uint32, scriptSig []byte, sequence uint32) *TransactionBuilder {
	input := TxInput{
		PrevTxID:  prevTxID,
		PrevVout:  prevVout,
		ScriptSig: scriptSig,
		Sequence:  sequence,
	}
	tb.inputs = append(tb.inputs, input)
	return tb
}

// AddOutput adds an output to the transaction
func (tb *TransactionBuilder) AddOutput(address string, amount uint64) error {
	// Decode address to get scriptPubKey
	scriptPubKey, err := addressToScriptPubKey(address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	output := TxOutput{
		Amount:       amount,
		ScriptPubKey: scriptPubKey,
	}
	tb.outputs = append(tb.outputs, output)
	return nil
}

// SetLocktime sets the locktime for the transaction
func (tb *TransactionBuilder) SetLocktime(locktime uint32) *TransactionBuilder {
	tb.locktime = locktime
	return tb
}

// Build builds the transaction
func (tb *TransactionBuilder) Build() (*Transaction, error) {
	if len(tb.inputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one input")
	}
	if len(tb.outputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one output")
	}

	return &Transaction{
		Version:  tb.version,
		Inputs:   tb.inputs,
		Outputs:  tb.outputs,
		Locktime: tb.locktime,
	}, nil
}

// addressToScriptPubKey converts a Bitcoin address to a scriptPubKey
func addressToScriptPubKey(address string) ([]byte, error) {
	// Decode Base58Check address
	version, decoded, err := encoding.DecodeBase58Check(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}

	// P2PKH: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	if version == 0x00 || version == 0x6f { // mainnet or testnet P2PKH
		script := make([]byte, 0, 25)
		script = append(script, 0x76) // OP_DUP
		script = append(script, 0xa9) // OP_HASH160
		script = append(script, 0x14) // Push 20 bytes
		script = append(script, decoded...)
		script = append(script, 0x88) // OP_EQUALVERIFY
		script = append(script, 0xac) // OP_CHECKSIG
		return script, nil
	}

	// P2SH: OP_HASH160 <scriptHash> OP_EQUAL
	if version == 0x05 || version == 0xc4 { // mainnet or testnet P2SH
		script := make([]byte, 0, 23)
		script = append(script, 0xa9) // OP_HASH160
		script = append(script, 0x14) // Push 20 bytes
		script = append(script, decoded...)
		script = append(script, 0x87) // OP_EQUAL
		return script, nil
	}

	return nil, fmt.Errorf("unsupported address type")
}

// Serialize serializes the transaction to hex
func (tx *Transaction) Serialize() (string, error) {
	buf := new(bytes.Buffer)

	// Version
	binary.Write(buf, binary.LittleEndian, tx.Version)

	// Input count
	writeVarInt(buf, uint64(len(tx.Inputs)))

	// Inputs
	for _, input := range tx.Inputs {
		// Previous txid (reversed)
		txidBytes, err := hex.DecodeString(input.PrevTxID)
		if err != nil {
			return "", fmt.Errorf("invalid txid: %w", err)
		}
		// Reverse for little-endian
		for i := len(txidBytes) - 1; i >= 0; i-- {
			buf.WriteByte(txidBytes[i])
		}

		// Previous vout
		binary.Write(buf, binary.LittleEndian, input.PrevVout)

		// ScriptSig length and data
		writeVarInt(buf, uint64(len(input.ScriptSig)))
		buf.Write(input.ScriptSig)

		// Sequence
		binary.Write(buf, binary.LittleEndian, input.Sequence)
	}

	// Output count
	writeVarInt(buf, uint64(len(tx.Outputs)))

	// Outputs
	for _, output := range tx.Outputs {
		// Amount
		binary.Write(buf, binary.LittleEndian, output.Amount)

		// ScriptPubKey length and data
		writeVarInt(buf, uint64(len(output.ScriptPubKey)))
		buf.Write(output.ScriptPubKey)
	}

	// Locktime
	binary.Write(buf, binary.LittleEndian, tx.Locktime)

	return hex.EncodeToString(buf.Bytes()), nil
}

// Hash calculates the transaction hash (TXID)
func (tx *Transaction) Hash() string {
	serialized, err := tx.Serialize()
	if err != nil {
		return ""
	}

	data, _ := hex.DecodeString(serialized)
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])

	// Reverse for display
	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = hash2[31-i]
	}

	return hex.EncodeToString(reversed)
}

// writeVarInt writes a variable-length integer
func writeVarInt(buf *bytes.Buffer, n uint64) {
	if n < 0xfd {
		buf.WriteByte(byte(n))
	} else if n <= 0xffff {
		buf.WriteByte(0xfd)
		binary.Write(buf, binary.LittleEndian, uint16(n))
	} else if n <= 0xffffffff {
		buf.WriteByte(0xfe)
		binary.Write(buf, binary.LittleEndian, uint32(n))
	} else {
		buf.WriteByte(0xff)
		binary.Write(buf, binary.LittleEndian, n)
	}
}

// EstimateTransactionSize estimates the size of a transaction in bytes
func EstimateTransactionSize(inputCount, outputCount int) int {
	// Base: version (4) + locktime (4) + input count (1) + output count (1)
	base := 10

	// Each input: txid (32) + vout (4) + script length (1) + script (~107 avg for P2PKH) + sequence (4)
	inputSize := 148 * inputCount

	// Each output: amount (8) + script length (1) + script (~25 for P2PKH)
	outputSize := 34 * outputCount

	return base + inputSize + outputSize
}

// CalculateFee calculates the fee for a transaction given the input/output counts and fee rate
func CalculateFee(inputCount, outputCount int, feeRateSatPerByte uint64) uint64 {
	size := EstimateTransactionSize(inputCount, outputCount)
	return uint64(size) * feeRateSatPerByte
}
