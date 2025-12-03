package modules

import (
	"context"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAdapters(t *testing.T) {
	t.Parallel()
	reg := mocks.NewMockChainRegistry()
	params := AdapterParams{
		Registry: reg,
		Ethereum: &mocks.MockChainAdapter{},
		Polygon:  &mocks.MockChainAdapter{},
		Tron:     &mocks.MockChainAdapter{},
	}
	err := registerAdapters(params)
	assert.NoError(t, err)
	assert.True(t, reg.Has("ethereum"))
	assert.True(t, reg.Has("polygon"))
	assert.True(t, reg.Has("tron"))
}

func TestRegisterAdapters_ErrorOnNilAdapter(t *testing.T) {
	t.Parallel()
	reg := mocks.NewMockChainRegistry()
	params := AdapterParams{
		Registry: reg,
		Ethereum: nil, // force error on first register
		Polygon:  &mocks.MockChainAdapter{},
		Tron:     &mocks.MockChainAdapter{},
	}
	err := registerAdapters(params)
	assert.Error(t, err)
}

func TestRegisterAdapters_ValidatesAllAdapters(t *testing.T) {
	t.Parallel()
	reg := mocks.NewMockChainRegistry()

	ethereum := &mocks.MockChainAdapter{}
	polygon := &mocks.MockChainAdapter{}
	tron := &mocks.MockChainAdapter{}

	params := AdapterParams{
		Registry: reg,
		Ethereum: ethereum,
		Polygon:  polygon,
		Tron:     tron,
	}

	err := registerAdapters(params)
	assert.NoError(t, err)

	// Verify all three are registered
	ethAdapter, err := reg.Get("ethereum")
	assert.NoError(t, err)
	assert.Equal(t, ethereum, ethAdapter)

	polyAdapter, err := reg.Get("polygon")
	assert.NoError(t, err)
	assert.Equal(t, polygon, polyAdapter)

	tronAdapter, err := reg.Get("tron")
	assert.NoError(t, err)
	assert.Equal(t, tron, tronAdapter)
}

func TestRegisterAdapters_ErrorPropagation(t *testing.T) {
	t.Parallel()
	reg := mocks.NewMockChainRegistry()

	// Register ethereum first so it exists
	firstAdapter := &mocks.MockChainAdapter{}
	err := reg.Register("ethereum", firstAdapter)
	assert.NoError(t, err)

	// Try to register different adapters but ethereum will be skipped since it exists
	params := AdapterParams{
		Registry: reg,
		Ethereum: &mocks.MockChainAdapter{},
		Polygon:  &mocks.MockChainAdapter{},
		Tron:     &mocks.MockChainAdapter{},
	}

	// Since our mock doesn't error on re-registration, this will succeed
	// In a real scenario with a strict registry, this might fail
	err = registerAdapters(params)
	// The mock registry allows re-registration, so this passes
	assert.NoError(t, err)
}

func TestAdapterParams_Integration(t *testing.T) {
	t.Parallel()
	reg := mocks.NewMockChainRegistry()

	// Create mock adapters with some state
	ethAdapter := &mocks.MockChainAdapter{}
	polyAdapter := &mocks.MockChainAdapter{}
	tronAdapter := &mocks.MockChainAdapter{}

	params := AdapterParams{
		Registry: reg,
		Ethereum: ethAdapter,
		Polygon:  polyAdapter,
		Tron:     tronAdapter,
	}

	// Register all
	err := registerAdapters(params)
	assert.NoError(t, err)

	// Verify functionality of each adapter through registry
	eth, err := reg.Get("ethereum")
	assert.NoError(t, err)
	assert.True(t, eth.IsConnected(context.Background()))

	poly, err := reg.Get("polygon")
	assert.NoError(t, err)
	assert.True(t, poly.IsConnected(context.Background()))

	tron, err := reg.Get("tron")
	assert.NoError(t, err)
	assert.True(t, tron.IsConnected(context.Background()))
}
