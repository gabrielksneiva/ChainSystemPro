package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/modules"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestMain_AppInitialization(t *testing.T) {
	t.Parallel()

	// Create app with a timeout to prevent hanging
	app := fx.New(
		modules.LoggerModule,
		modules.EventBusModule,
		modules.RegistryModule,
		modules.AdaptersModule,
		modules.UseCasesModule,
		modules.APIModule,
		fx.NopLogger, // Suppress fx logs during tests
	)

	// Start the app in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startErr := app.Start(ctx)
	require.NoError(t, startErr, "app should start without errors")

	// Stop the app
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	stopErr := app.Stop(stopCtx)
	require.NoError(t, stopErr, "app should stop without errors")
}

func TestMain_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that main() can be called without panicking
	// We'll send a signal to stop it after a short time
	done := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
			done <- true
		}()
		main()
	}()

	// Wait a bit for the server to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal to stop the server gracefully
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	err = proc.Signal(os.Interrupt)
	require.NoError(t, err)

	// Wait for main to finish or timeout
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("main() did not stop after interrupt signal")
	}
}
