package mocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockChainRegistry_AllMethods(t *testing.T) {
	t.Parallel()
	r := NewMockChainRegistry()
	adapter := &MockChainAdapter{}

	// Test Register with nil should error
	err := r.Register("nil-test", nil)
	assert.Error(t, err)

	// Register valid adapter
	err = r.Register("eth", adapter)
	assert.NoError(t, err)
	assert.True(t, r.Has("eth"))

	// Get
	a, err := r.Get("eth")
	assert.NoError(t, err)
	assert.Equal(t, adapter, a)

	// Get non-existent
	_, err = r.Get("unknown")
	assert.Error(t, err)

	// List
	chains := r.List()
	assert.Contains(t, chains, "eth")

	// Unregister
	err = r.Unregister("eth")
	assert.NoError(t, err)
	assert.False(t, r.Has("eth"))
	_, err = r.Get("eth")
	assert.Error(t, err)
}

func TestMockEventPublisher_AllMethods(t *testing.T) {
	t.Parallel()
	p := NewMockEventPublisher()
	event := struct{ Name string }{"evt"}
	err := p.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.Len(t, p.PublishedEvents, 1)
	events := []interface{}{event, event}
	err = p.PublishBatch(context.Background(), events)
	assert.NoError(t, err)
	assert.Len(t, p.PublishedEvents, 3)
}

func TestMockLogger_AllMethods(t *testing.T) {
	t.Parallel()
	l := NewMockLogger()
	l.Debug("debug", nil)
	l.Info("info", nil)
	l.Warn("warn", nil)
	l.Error("err", nil, nil)
	l.Fatal("fatal", nil, nil)
	assert.Len(t, l.DebugCalls, 1)
	assert.Len(t, l.InfoCalls, 1)
	assert.Len(t, l.WarnCalls, 1)
	assert.Len(t, l.ErrorCalls, 1)
	assert.Len(t, l.FatalCalls, 1)
}
