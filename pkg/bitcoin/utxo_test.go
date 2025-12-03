package bitcoin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectUTXOs_FIFO(t *testing.T) {
	utxos := []UTXO{
		{TxID: "tx1", Vout: 0, Amount: 100000, Confirmations: 10},
		{TxID: "tx2", Vout: 0, Amount: 50000, Confirmations: 5},
		{TxID: "tx3", Vout: 0, Amount: 25000, Confirmations: 15},
	}

	tests := []struct {
		name          string
		targetAmount  uint64
		expectedCount int
		expectedTotal uint64
		expectErr     bool
	}{
		{
			name:          "exact match first UTXO",
			targetAmount:  100000,
			expectedCount: 1,
			expectedTotal: 100000,
			expectErr:     false,
		},
		{
			name:          "need multiple UTXOs",
			targetAmount:  120000,
			expectedCount: 2,
			expectedTotal: 150000,
			expectErr:     false,
		},
		{
			name:          "need all UTXOs",
			targetAmount:  170000,
			expectedCount: 3,
			expectedTotal: 175000,
			expectErr:     false,
		},
		{
			name:          "insufficient funds",
			targetAmount:  200000,
			expectedCount: 0,
			expectedTotal: 0,
			expectErr:     true,
		},
		{
			name:          "zero amount",
			targetAmount:  0,
			expectedCount: 0,
			expectedTotal: 0,
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected, total, err := SelectUTXOs(utxos, tt.targetAmount, AlgorithmFIFO)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(selected))
				assert.Equal(t, tt.expectedTotal, total)

				// Verify total is correct
				var sum uint64
				for _, utxo := range selected {
					sum += utxo.Amount
				}
				assert.Equal(t, total, sum)
			}
		})
	}
}

func TestSelectUTXOs_LargestFirst(t *testing.T) {
	utxos := []UTXO{
		{TxID: "tx1", Vout: 0, Amount: 25000, Confirmations: 10},
		{TxID: "tx2", Vout: 0, Amount: 100000, Confirmations: 5},
		{TxID: "tx3", Vout: 0, Amount: 50000, Confirmations: 15},
	}

	tests := []struct {
		name          string
		targetAmount  uint64
		expectedCount int
		expectedTotal uint64
		expectErr     bool
	}{
		{
			name:          "exact match largest UTXO",
			targetAmount:  100000,
			expectedCount: 1,
			expectedTotal: 100000,
			expectErr:     false,
		},
		{
			name:          "need two largest",
			targetAmount:  120000,
			expectedCount: 2,
			expectedTotal: 150000,
			expectErr:     false,
		},
		{
			name:          "small amount uses smallest",
			targetAmount:  30000,
			expectedCount: 1,
			expectedTotal: 100000,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected, total, err := SelectUTXOs(utxos, tt.targetAmount, AlgorithmLargestFirst)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(selected))
				assert.Equal(t, tt.expectedTotal, total)
			}
		})
	}
}

func TestSelectUTXOs_EmptyUTXOs(t *testing.T) {
	utxos := []UTXO{}
	_, _, err := SelectUTXOs(utxos, 100000, AlgorithmFIFO)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no UTXOs available")
}

func TestSelectUTXOs_InvalidAlgorithm(t *testing.T) {
	utxos := []UTXO{
		{TxID: "tx1", Vout: 0, Amount: 100000, Confirmations: 10},
	}
	_, _, err := SelectUTXOs(utxos, 50000, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown algorithm")
}

func TestFilterConfirmed(t *testing.T) {
	utxos := []UTXO{
		{TxID: "tx1", Vout: 0, Amount: 100000, Confirmations: 0},
		{TxID: "tx2", Vout: 0, Amount: 50000, Confirmations: 1},
		{TxID: "tx3", Vout: 0, Amount: 25000, Confirmations: 6},
		{TxID: "tx4", Vout: 0, Amount: 10000, Confirmations: 10},
	}

	tests := []struct {
		name          string
		minConf       uint32
		expectedCount int
	}{
		{
			name:          "all UTXOs (minConf=0)",
			minConf:       0,
			expectedCount: 4,
		},
		{
			name:          "confirmed UTXOs (minConf=1)",
			minConf:       1,
			expectedCount: 3,
		},
		{
			name:          "standard confirmations (minConf=6)",
			minConf:       6,
			expectedCount: 2,
		},
		{
			name:          "high confirmations (minConf=10)",
			minConf:       10,
			expectedCount: 1,
		},
		{
			name:          "no UTXOs meet requirement",
			minConf:       100,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterConfirmed(utxos, tt.minConf)
			assert.Equal(t, tt.expectedCount, len(filtered))

			// Verify all filtered UTXOs meet requirement
			for _, utxo := range filtered {
				assert.GreaterOrEqual(t, utxo.Confirmations, tt.minConf)
			}
		})
	}
}

func TestFilterByMinAmount(t *testing.T) {
	utxos := []UTXO{
		{TxID: "tx1", Vout: 0, Amount: 1000, Confirmations: 6},
		{TxID: "tx2", Vout: 0, Amount: 5000, Confirmations: 6},
		{TxID: "tx3", Vout: 0, Amount: 10000, Confirmations: 6},
		{TxID: "tx4", Vout: 0, Amount: 50000, Confirmations: 6},
	}

	tests := []struct {
		name          string
		minAmount     uint64
		expectedCount int
	}{
		{
			name:          "all UTXOs",
			minAmount:     0,
			expectedCount: 4,
		},
		{
			name:          "filter dust (minAmount=5460)",
			minAmount:     5460,
			expectedCount: 2,
		},
		{
			name:          "only large UTXOs",
			minAmount:     25000,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByMinAmount(utxos, tt.minAmount)
			assert.Equal(t, tt.expectedCount, len(filtered))

			// Verify all filtered UTXOs meet requirement
			for _, utxo := range filtered {
				assert.GreaterOrEqual(t, utxo.Amount, tt.minAmount)
			}
		})
	}
}

func TestTotalAmount(t *testing.T) {
	tests := []struct {
		name     string
		utxos    []UTXO
		expected uint64
	}{
		{
			name:     "empty slice",
			utxos:    []UTXO{},
			expected: 0,
		},
		{
			name: "single UTXO",
			utxos: []UTXO{
				{TxID: "tx1", Vout: 0, Amount: 100000, Confirmations: 6},
			},
			expected: 100000,
		},
		{
			name: "multiple UTXOs",
			utxos: []UTXO{
				{TxID: "tx1", Vout: 0, Amount: 100000, Confirmations: 6},
				{TxID: "tx2", Vout: 0, Amount: 50000, Confirmations: 6},
				{TxID: "tx3", Vout: 0, Amount: 25000, Confirmations: 6},
			},
			expected: 175000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := TotalAmount(tt.utxos)
			assert.Equal(t, tt.expected, total)
		})
	}
}

func TestUTXO_String(t *testing.T) {
	utxo := UTXO{
		TxID:          "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		Vout:          0,
		Amount:        50000000,
		Confirmations: 100,
	}

	str := utxo.String()
	assert.Contains(t, str, "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
	assert.Contains(t, str, "0")
}
