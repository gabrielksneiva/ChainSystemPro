package modules

import (
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
