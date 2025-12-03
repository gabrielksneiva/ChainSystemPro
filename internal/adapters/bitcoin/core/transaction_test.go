package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionBuilder(t *testing.T) {
	builder := NewTransactionBuilder()
	assert.NotNil(t, builder)
	assert.Equal(t, uint32(2), builder.version)
	assert.Equal(t, uint32(0), builder.locktime)
}

func TestTransactionBuilder_AddInput(t *testing.T) {
	builder := NewTransactionBuilder()

	txid := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	vout := uint32(0)

	builder.AddInput(txid, vout, nil, 0xffffffff)

	assert.Equal(t, 1, len(builder.inputs))
	assert.Equal(t, txid, builder.inputs[0].PrevTxID)
	assert.Equal(t, vout, builder.inputs[0].PrevVout)
	assert.Equal(t, uint32(0xffffffff), builder.inputs[0].Sequence)
}

func TestTransactionBuilder_AddOutput(t *testing.T) {
	builder := NewTransactionBuilder()

	address := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	amount := uint64(100000)

	err := builder.AddOutput(address, amount)
	require.NoError(t, err)

	assert.Equal(t, 1, len(builder.outputs))
	assert.Equal(t, amount, builder.outputs[0].Amount)
}

func TestTransactionBuilder_AddOutput_InvalidAddress(t *testing.T) {
	builder := NewTransactionBuilder()

	err := builder.AddOutput("invalid", 100000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestTransactionBuilder_SetLocktime(t *testing.T) {
	builder := NewTransactionBuilder()
	builder.SetLocktime(500000)

	assert.Equal(t, uint32(500000), builder.locktime)
}

func TestTransactionBuilder_Build(t *testing.T) {
	builder := NewTransactionBuilder()

	// Add input
	builder.AddInput("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", 0, nil, 0xffffffff)

	// Add output
	err := builder.AddOutput("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 50000000)
	require.NoError(t, err)

	tx, err := builder.Build()
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, 1, len(tx.Inputs))
	assert.Equal(t, 1, len(tx.Outputs))
	assert.Equal(t, uint32(2), tx.Version)
}

func TestTransactionBuilder_Build_NoInputs(t *testing.T) {
	builder := NewTransactionBuilder()
	builder.AddOutput("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 50000000)

	_, err := builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one input")
}

func TestTransactionBuilder_Build_NoOutputs(t *testing.T) {
	builder := NewTransactionBuilder()
	builder.AddInput("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", 0, nil, 0xffffffff)

	_, err := builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one output")
}

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name         string
		inputCount   int
		outputCount  int
		feeRate      uint64
		expectedSize int
		expectedFee  uint64
	}{
		{
			name:         "single input, single output",
			inputCount:   1,
			outputCount:  1,
			feeRate:      1,
			expectedSize: 192, // approx: 10 + 148*1 + 34*1
			expectedFee:  192,
		},
		{
			name:         "two inputs, two outputs",
			inputCount:   2,
			outputCount:  2,
			feeRate:      10,
			expectedSize: 374, // approx: 10 + 148*2 + 34*2
			expectedFee:  3740,
		},
		{
			name:         "high fee rate",
			inputCount:   1,
			outputCount:  1,
			feeRate:      100,
			expectedSize: 192,
			expectedFee:  19200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := EstimateTransactionSize(tt.inputCount, tt.outputCount)
			assert.Equal(t, tt.expectedSize, size)

			fee := CalculateFee(tt.inputCount, tt.outputCount, tt.feeRate)
			assert.Equal(t, tt.expectedFee, fee)
		})
	}
}

func TestTransaction_Hash(t *testing.T) {
	builder := NewTransactionBuilder()
	builder.AddInput("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", 0, nil, 0xffffffff)
	builder.AddOutput("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 50000000)

	tx, err := builder.Build()
	require.NoError(t, err)

	hash := tx.Hash()
	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash)) // hex string of 32 bytes
}

func TestTransaction_Serialize(t *testing.T) {
	builder := NewTransactionBuilder()
	builder.AddInput("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", 0, nil, 0xffffffff)
	builder.AddOutput("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 50000000)

	tx, err := builder.Build()
	require.NoError(t, err)

	serialized, err := tx.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, serialized)

	// Verify it's hex
	assert.Regexp(t, "^[0-9a-f]+$", serialized)
}

func TestTxInput_Value(t *testing.T) {
	input := TxInput{
		PrevTxID:  "abc123",
		PrevVout:  0,
		Sequence:  0xffffffff,
	}

	assert.Equal(t, "abc123", input.PrevTxID)
	assert.Equal(t, uint32(0), input.PrevVout)
	assert.Equal(t, uint32(0xffffffff), input.Sequence)
}

func TestTxOutput_Value(t *testing.T) {
	output := TxOutput{
		Amount:       50000000,
		ScriptPubKey: []byte{0x76, 0xa9},
	}

	assert.Equal(t, uint64(50000000), output.Amount)
	assert.NotEmpty(t, output.ScriptPubKey)
}

func TestAddressToScriptPubKey_P2PKH(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "mainnet P2PKH",
			address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			wantErr: false,
		},
		{
			name:    "testnet P2PKH",
			address: "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := addressToScriptPubKey(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, script)
				// P2PKH should be 25 bytes
				assert.Equal(t, 25, len(script))
				assert.Equal(t, byte(0x76), script[0]) // OP_DUP
				assert.Equal(t, byte(0xa9), script[1]) // OP_HASH160
			}
		})
	}
}

func TestAddressToScriptPubKey_P2SH(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "mainnet P2SH",
			address: "3DemCeo7BSdkHz1beo1DaGqUddmV1qyD6P",
			wantErr: false,
		},
		{
			name:    "testnet P2SH",
			address: "2N4Q5FhU2497BryFfUgbqkAJE87aKHUhXMp",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := addressToScriptPubKey(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, script)
				// P2SH should be 23 bytes
				assert.Equal(t, 23, len(script))
				assert.Equal(t, byte(0xa9), script[0])  // OP_HASH160
				assert.Equal(t, byte(0x87), script[22]) // OP_EQUAL
			}
		})
	}
}

func TestAddressToScriptPubKey_InvalidAddress(t *testing.T) {
	_, err := addressToScriptPubKey("invalid")
	assert.Error(t, err)
}

func TestAddressToScriptPubKey_UnsupportedType(t *testing.T) {
	// Bech32 address (not supported yet in addressToScriptPubKey)
	_, err := addressToScriptPubKey("bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode address")
}
