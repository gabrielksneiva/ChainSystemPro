package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// MockChainRegistry is a mock implementation of ChainRegistry
type MockChainRegistry struct {
	adapters map[string]ports.ChainAdapter
}

// NewMockChainRegistry creates a new mock chain registry
func NewMockChainRegistry() *MockChainRegistry {
	return &MockChainRegistry{
		adapters: make(map[string]ports.ChainAdapter),
	}
}

func (r *MockChainRegistry) Register(chainID string, adapter ports.ChainAdapter) error {
	if chainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil")
	}
	r.adapters[chainID] = adapter
	return nil
}

func (r *MockChainRegistry) Unregister(chainID string) error {
	delete(r.adapters, chainID)
	return nil
}

func (r *MockChainRegistry) Get(chainID string) (ports.ChainAdapter, error) {
	adapter, exists := r.adapters[chainID]
	if !exists {
		return nil, fmt.Errorf("chain adapter not found: %s", chainID)
	}
	return adapter, nil
}

func (r *MockChainRegistry) List() []string {
	chains := make([]string, 0, len(r.adapters))
	for chainID := range r.adapters {
		chains = append(chains, chainID)
	}
	return chains
}

func (r *MockChainRegistry) Has(chainID string) bool {
	_, exists := r.adapters[chainID]
	return exists
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	PublishedEvents []interface{}
}

// NewMockEventPublisher creates a new mock event publisher
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		PublishedEvents: make([]interface{}, 0),
	}
}

func (p *MockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	p.PublishedEvents = append(p.PublishedEvents, event)
	return nil
}

func (p *MockEventPublisher) PublishBatch(ctx context.Context, events []interface{}) error {
	p.PublishedEvents = append(p.PublishedEvents, events...)
	return nil
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mu         sync.Mutex
	DebugCalls []LogCall
	InfoCalls  []LogCall
	WarnCalls  []LogCall
	ErrorCalls []LogCall
	FatalCalls []LogCall
}

// LogCall represents a log method call
type LogCall struct {
	Message string
	Fields  map[string]interface{}
	Error   error
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugCalls: make([]LogCall, 0),
		InfoCalls:  make([]LogCall, 0),
		WarnCalls:  make([]LogCall, 0),
		ErrorCalls: make([]LogCall, 0),
		FatalCalls: make([]LogCall, 0),
	}
}

func (l *MockLogger) Debug(msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.DebugCalls = append(l.DebugCalls, LogCall{Message: msg, Fields: fields})
}

func (l *MockLogger) Info(msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.InfoCalls = append(l.InfoCalls, LogCall{Message: msg, Fields: fields})
}

func (l *MockLogger) Warn(msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.WarnCalls = append(l.WarnCalls, LogCall{Message: msg, Fields: fields})
}

func (l *MockLogger) Error(msg string, err error, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ErrorCalls = append(l.ErrorCalls, LogCall{Message: msg, Fields: fields, Error: err})
}

func (l *MockLogger) Fatal(msg string, err error, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.FatalCalls = append(l.FatalCalls, LogCall{Message: msg, Fields: fields, Error: err})
}
