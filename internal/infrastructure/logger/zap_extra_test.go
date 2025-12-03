package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZapLogger_ErrorCases(t *testing.T) {
	t.Parallel()
	logger, err := NewZapLogger("debug")
	require.NoError(t, err)
	logger.Error("error with nil", nil, nil)
	logger.Error("error with err", errors.New("err"), map[string]interface{}{"foo": "bar"})
}

func TestZapLoggerWarnFatal(t *testing.T) {
	t.Parallel()
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)
	require.NotNil(t, logger)
	logger.Warn("warning message", map[string]interface{}{"key": "value"})
}

func TestNewZapLogger_InvalidLevel(t *testing.T) {
	t.Parallel()
	_, err := NewZapLogger("invalid")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestZapLogger_Sync_Extra(t *testing.T) {
	t.Parallel()
	logger, err := NewZapLogger("debug")
	require.NoError(t, err)
	_ = logger.Sync()
}
