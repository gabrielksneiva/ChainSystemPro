package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/gabrielksneiva/ChainSystemPro/internal/encoding"
)

type TxInput struct {
	PrevTxID  string
	PrevVout  uint32
	ScriptSig []byte
	Sequence  uint32
}

type TxOutput struct {
	Amount       uint64
	ScriptPubKey []byte
}

type Transaction struct {
	Version  uint32
	Inputs   []TxInput
	Outputs  []TxOutput
	Locktime uint32
}

type TransactionBuilder struct {
	version  uint32
	inputs   []TxInput
	outputs  []TxOutput
	locktime uint32
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{version: 2, inputs: []TxInput{}, outputs: []TxOutput{}, locktime: 0}
}

func (tb *TransactionBuilder) AddInput(prevTxID string, prevVout uint32, scriptSig []byte, sequence uint32) *TransactionBuilder {
	input := TxInput{PrevTxID: prevTxID, PrevVout: prevVout, ScriptSig: scriptSig, Sequence: sequence}
	tb.inputs = append(tb.inputs, input)
	return tb
}

func (tb *TransactionBuilder) AddOutput(address string, amount uint64) error {
	scriptPubKey, err := addressToScriptPubKey(address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	output := TxOutput{Amount: amount, ScriptPubKey: scriptPubKey}
	tb.outputs = append(tb.outputs, output)
	return nil
}

func (tb *TransactionBuilder) SetLocktime(locktime uint32) *TransactionBuilder {
	tb.locktime = locktime
	return tb
}

func (tb *TransactionBuilder) Build() (*Transaction, error) {
	if len(tb.inputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one input")
	}
	if len(tb.outputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one output")
	}
	return &Transaction{Version: tb.version, Inputs: tb.inputs, Outputs: tb.outputs, Locktime: tb.locktime}, nil
}

func addressToScriptPubKey(address string) ([]byte, error) {
	version, decoded, err := encoding.DecodeBase58Check(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}
	if version == 0x00 || version == 0x6f {
		script := make([]byte, 0, 25)
		script = append(script, 0x76, 0xa9, 0x14)
		script = append(script, decoded...)
		script = append(script, 0x88, 0xac)
		return script, nil
	}
	if version == 0x05 || version == 0xc4 {
		script := make([]byte, 0, 23)
		script = append(script, 0xa9, 0x14)
		script = append(script, decoded...)
		script = append(script, 0x87)
		return script, nil
	}
	return nil, fmt.Errorf("unsupported address type")
}

func (tx *Transaction) Serialize() (string, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Version)
	writeVarInt(buf, uint64(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		txidBytes, err := hex.DecodeString(input.PrevTxID)
		if err != nil {
			return "", fmt.Errorf("invalid txid: %w", err)
		}
		for i := len(txidBytes) - 1; i >= 0; i-- {
			buf.WriteByte(txidBytes[i])
		}
		binary.Write(buf, binary.LittleEndian, input.PrevVout)
		writeVarInt(buf, uint64(len(input.ScriptSig)))
		buf.Write(input.ScriptSig)
		binary.Write(buf, binary.LittleEndian, input.Sequence)
	}
	writeVarInt(buf, uint64(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		binary.Write(buf, binary.LittleEndian, output.Amount)
		writeVarInt(buf, uint64(len(output.ScriptPubKey)))
		buf.Write(output.ScriptPubKey)
	}
	binary.Write(buf, binary.LittleEndian, tx.Locktime)
	return hex.EncodeToString(buf.Bytes()), nil
}

func (tx *Transaction) Hash() string {
	serialized, err := tx.Serialize()
	if err != nil {
		return ""
	}
	data, _ := hex.DecodeString(serialized)
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])
	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = hash2[31-i]
	}
	return hex.EncodeToString(reversed)
}

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

func EstimateTransactionSize(inputCount, outputCount int) int {
	base := 10
	inputSize := 148 * inputCount
	outputSize := 34 * outputCount
	return base + inputSize + outputSize
}

func CalculateFee(inputCount, outputCount int, feeRateSatPerByte uint64) uint64 {
	size := EstimateTransactionSize(inputCount, outputCount)
	return uint64(size) * feeRateSatPerByte
}
