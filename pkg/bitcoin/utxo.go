package bitcoin

import (
	"fmt"
	"sort"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID          string
	Vout          uint32
	Amount        uint64 // satoshis
	ScriptPubKey  string
	Confirmations uint32
}

// String returns a string representation of the UTXO
func (u UTXO) String() string {
	return fmt.Sprintf("%s:%d", u.TxID, u.Vout)
}

// CoinSelectionAlgorithm defines the type for coin selection algorithms
type CoinSelectionAlgorithm string

const (
	// AlgorithmFIFO selects UTXOs in the order they appear
	AlgorithmFIFO CoinSelectionAlgorithm = "fifo"
	// AlgorithmLargestFirst selects largest UTXOs first
	AlgorithmLargestFirst CoinSelectionAlgorithm = "largest-first"
)

// SelectUTXOs selects UTXOs to meet the target amount using the specified algorithm
func SelectUTXOs(utxos []UTXO, targetAmount uint64, algorithm CoinSelectionAlgorithm) ([]UTXO, uint64, error) {
	if len(utxos) == 0 {
		return nil, 0, fmt.Errorf("no UTXOs available")
	}

	if targetAmount == 0 {
		return nil, 0, fmt.Errorf("target amount must be greater than 0")
	}

	var selected []UTXO
	var total uint64

	switch algorithm {
	case AlgorithmFIFO:
		selected, total = selectFIFO(utxos, targetAmount)
	case AlgorithmLargestFirst:
		selected, total = selectLargestFirst(utxos, targetAmount)
	default:
		return nil, 0, fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	if total < targetAmount {
		return nil, 0, fmt.Errorf("insufficient funds: have %d, need %d", total, targetAmount)
	}

	return selected, total, nil
}

// selectFIFO selects UTXOs in order until target is met
func selectFIFO(utxos []UTXO, targetAmount uint64) ([]UTXO, uint64) {
	var selected []UTXO
	var total uint64

	for _, utxo := range utxos {
		selected = append(selected, utxo)
		total += utxo.Amount
		if total >= targetAmount {
			break
		}
	}

	return selected, total
}

// selectLargestFirst selects largest UTXOs first
func selectLargestFirst(utxos []UTXO, targetAmount uint64) ([]UTXO, uint64) {
	// Create a copy to avoid modifying original slice
	sortedUTXOs := make([]UTXO, len(utxos))
	copy(sortedUTXOs, utxos)

	// Sort by amount descending
	sort.Slice(sortedUTXOs, func(i, j int) bool {
		return sortedUTXOs[i].Amount > sortedUTXOs[j].Amount
	})

	return selectFIFO(sortedUTXOs, targetAmount)
}

// FilterConfirmed returns UTXOs with at least minConf confirmations
func FilterConfirmed(utxos []UTXO, minConf uint32) []UTXO {
	filtered := make([]UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		if utxo.Confirmations >= minConf {
			filtered = append(filtered, utxo)
		}
	}
	return filtered
}

// FilterByMinAmount returns UTXOs with at least minAmount satoshis
func FilterByMinAmount(utxos []UTXO, minAmount uint64) []UTXO {
	filtered := make([]UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		if utxo.Amount >= minAmount {
			filtered = append(filtered, utxo)
		}
	}
	return filtered
}

// TotalAmount calculates the total amount of all UTXOs
func TotalAmount(utxos []UTXO) uint64 {
	var total uint64
	for _, utxo := range utxos {
		total += utxo.Amount
	}
	return total
}
