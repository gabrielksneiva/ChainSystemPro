package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Renamed to avoid duplication with existing sync test
func TestZapLogger_Sync_Dupe(t *testing.T) {
	t.Parallel()
	logger, err := NewZapLogger("debug")
	require.NoError(t, err)
	_ = logger.Sync()
}
