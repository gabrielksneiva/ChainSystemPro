package registry

import (
	"fmt"
	"sync"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// ChainRegistry manages chain adapters
type ChainRegistry struct {
	adapters map[string]ports.ChainAdapter
	mu       sync.RWMutex
	logger   ports.Logger
}

// NewChainRegistry creates a new chain registry
func NewChainRegistry(logger ports.Logger) *ChainRegistry {
	return &ChainRegistry{
		adapters: make(map[string]ports.ChainAdapter),
		logger:   logger,
	}
}

// Register registers a chain adapter
func (r *ChainRegistry) Register(chainID string, adapter ports.ChainAdapter) error {
	if chainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[chainID]; exists {
		return fmt.Errorf("chain adapter already registered: %s", chainID)
	}

	r.adapters[chainID] = adapter

	r.logger.Info("chain adapter registered", map[string]interface{}{
		"chain_id": chainID,
	})

	return nil
}

// Unregister unregisters a chain adapter
func (r *ChainRegistry) Unregister(chainID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[chainID]; !exists {
		return fmt.Errorf("chain adapter not found: %s", chainID)
	}

	delete(r.adapters, chainID)

	r.logger.Info("chain adapter unregistered", map[string]interface{}{
		"chain_id": chainID,
	})

	return nil
}

// Get returns a chain adapter by ID
func (r *ChainRegistry) Get(chainID string) (ports.ChainAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[chainID]
	if !exists {
		return nil, fmt.Errorf("chain adapter not found: %s", chainID)
	}

	return adapter, nil
}

// List returns all registered chain IDs
func (r *ChainRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chains := make([]string, 0, len(r.adapters))
	for chainID := range r.adapters {
		chains = append(chains, chainID)
	}

	return chains
}

// Has checks if a chain adapter is registered
func (r *ChainRegistry) Has(chainID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.adapters[chainID]
	return exists
}
