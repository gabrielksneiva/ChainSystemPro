package usecases

import (
	"math/big"
	"strings"
)

// parseBigInt parses a string into a big.Int
func parseBigInt(s string) (*big.Int, bool) {
	if s == "" {
		return big.NewInt(0), true
	}

	s = strings.TrimSpace(s)
	value := new(big.Int)
	_, ok := value.SetString(s, 10)
	return value, ok
}
