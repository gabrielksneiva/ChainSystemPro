package usecases

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBigInt(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns zero", func(t *testing.T) {
		t.Parallel()
		val, ok := parseBigInt("")
		require.True(t, ok)
		require.Equal(t, big.NewInt(0), val)
	})

	t.Run("valid number", func(t *testing.T) {
		t.Parallel()
		val, ok := parseBigInt("12345")
		require.True(t, ok)
		require.Equal(t, big.NewInt(12345), val)
	})

	t.Run("number with spaces", func(t *testing.T) {
		t.Parallel()
		val, ok := parseBigInt("  999  ")
		require.True(t, ok)
		require.Equal(t, big.NewInt(999), val)
	})

	t.Run("invalid number", func(t *testing.T) {
		t.Parallel()
		_, ok := parseBigInt("notanumber")
		require.False(t, ok)
	})

	t.Run("large number", func(t *testing.T) {
		t.Parallel()
		val, ok := parseBigInt("999999999999999999999999999")
		require.True(t, ok)
		expected, _ := new(big.Int).SetString("999999999999999999999999999", 10)
		require.Equal(t, expected, val)
	})
}
