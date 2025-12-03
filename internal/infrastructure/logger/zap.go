package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a zap-based logger implementation
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zap logger
func NewZapLogger(level string) (*ZapLogger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{logger: logger}, nil
}

// NewDevelopmentLogger creates a development logger
func NewDevelopmentLogger() (*ZapLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create development logger: %w", err)
	}

	return &ZapLogger{logger: logger}, nil
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields map[string]interface{}) {
	l.logger.Debug(msg, zapFields(fields)...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields map[string]interface{}) {
	l.logger.Info(msg, zapFields(fields)...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields map[string]interface{}) {
	l.logger.Warn(msg, zapFields(fields)...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, err error, fields map[string]interface{}) {
	if err != nil {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		fields["error"] = err.Error()
	}
	l.logger.Error(msg, zapFields(fields)...)
}

// Fatal logs a fatal message and exits
func (l *ZapLogger) Fatal(msg string, err error, fields map[string]interface{}) {
	if err != nil {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		fields["error"] = err.Error()
	}
	l.logger.Fatal(msg, zapFields(fields)...)
	os.Exit(1)
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// zapFields converts map to zap fields
func zapFields(fields map[string]interface{}) []zap.Field {
	if fields == nil {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return zapFields
}
