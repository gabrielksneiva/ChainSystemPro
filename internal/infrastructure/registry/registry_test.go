package registry

import (
	"fmt"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChainRegistry(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	require.NotNil(t, registry)
	assert.NotNil(t, registry.adapters)
	assert.Equal(t, 0, len(registry.List()))
}

func TestChainRegistry_Register(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	t.Run("success", func(t *testing.T) {
		adapter := &mocks.MockChainAdapter{}
		err := registry.Register("ethereum", adapter)

		require.NoError(t, err)
		assert.True(t, registry.Has("ethereum"))
		assert.Contains(t, registry.List(), "ethereum")
	})

	t.Run("error - empty chain ID", func(t *testing.T) {
		adapter := &mocks.MockChainAdapter{}
		err := registry.Register("", adapter)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "chain ID cannot be empty")
	})

	t.Run("error - nil adapter", func(t *testing.T) {
		err := registry.Register("ethereum", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "adapter cannot be nil")
	})

	t.Run("error - already registered", func(t *testing.T) {
		registry := NewChainRegistry(logger)
		adapter := &mocks.MockChainAdapter{}

		err := registry.Register("ethereum", adapter)
		require.NoError(t, err)

		err = registry.Register("ethereum", adapter)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestChainRegistry_Get(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	t.Run("success", func(t *testing.T) {
		adapter := &mocks.MockChainAdapter{}
		_ = registry.Register("ethereum", adapter)

		retrieved, err := registry.Get("ethereum")

		require.NoError(t, err)
		assert.Equal(t, adapter, retrieved)
	})

	t.Run("error - not found", func(t *testing.T) {
		_, err := registry.Get("unknown")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestChainRegistry_Unregister(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	t.Run("success", func(t *testing.T) {
		adapter := &mocks.MockChainAdapter{}
		_ = registry.Register("ethereum", adapter)

		err := registry.Unregister("ethereum")

		require.NoError(t, err)
		assert.False(t, registry.Has("ethereum"))
	})

	t.Run("error - not found", func(t *testing.T) {
		err := registry.Unregister("unknown")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestChainRegistry_List(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	adapter1 := &mocks.MockChainAdapter{}
	adapter2 := &mocks.MockChainAdapter{}

	_ = registry.Register("ethereum", adapter1)
	_ = registry.Register("polygon", adapter2)

	chains := registry.List()

	assert.Len(t, chains, 2)
	assert.Contains(t, chains, "ethereum")
	assert.Contains(t, chains, "polygon")
}

func TestChainRegistry_Has(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	adapter := &mocks.MockChainAdapter{}
	_ = registry.Register("ethereum", adapter)

	assert.True(t, registry.Has("ethereum"))
	assert.False(t, registry.Has("unknown"))
}

func TestChainRegistry_Concurrent(t *testing.T) {
	logger := mocks.NewMockLogger()
	registry := NewChainRegistry(logger)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			adapter := &mocks.MockChainAdapter{}
			chainID := fmt.Sprintf("chain-%d", id)
			_ = registry.Register(chainID, adapter)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Len(t, registry.List(), 10)
}
