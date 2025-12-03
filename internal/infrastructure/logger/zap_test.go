package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDevelopmentLogger(t *testing.T) {
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewZapLogger(t *testing.T) {
	logger, err := NewZapLogger("info")
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestZapLogger_Info(t *testing.T) {
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Info("test message", nil)
	})

	assert.NotPanics(t, func() {
		logger.Info("test message with fields", map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		})
	})
}

func TestZapLogger_Error(t *testing.T) {
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Error("test error", assert.AnError, nil)
	})

	assert.NotPanics(t, func() {
		logger.Error("test error with fields", assert.AnError, map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		})
	})
}

func TestZapLogger_Debug(t *testing.T) {
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Debug("test debug", nil)
	})

	assert.NotPanics(t, func() {
		logger.Debug("test debug with fields", map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		})
	})
}

func TestZapLogger_Sync(t *testing.T) {
	logger, err := NewDevelopmentLogger()
	require.NoError(t, err)

	// Should not panic or error
	assert.NotPanics(t, func() {
		logger.Sync()
	})
}

func TestNewZapLoggerLevels(t *testing.T) {
	tests := []struct {
		level string
	}{
		{"debug"},
		{"info"},
		{"warn"},
		{"error"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger, err := NewZapLogger(tt.level)
			require.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestZapFieldsNil(t *testing.T) {
	fields := zapFields(nil)
	assert.Nil(t, fields)
}

func TestZapFieldsEmpty(t *testing.T) {
	fields := zapFields(map[string]interface{}{})
	assert.NotNil(t, fields)
	assert.Len(t, fields, 0)
}

func TestZapFieldsMultiple(t *testing.T) {
	fields := zapFields(map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	})
	assert.NotNil(t, fields)
	assert.Len(t, fields, 3)
}
