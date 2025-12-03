package eventbus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEvent struct {
	eventType string
	data      string
}

func (e testEvent) Type() string {
	return e.eventType
}

func TestNewInMemoryEventBus(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)

	require.NotNil(t, bus)
	assert.NotNil(t, bus.subscribers)
}

func TestInMemoryEventBus_PublishSubscribe(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	t.Run("subscribe and publish event", func(t *testing.T) {
		received := make(chan interface{}, 1)
		handler := func(ctx context.Context, event interface{}) error {
			received <- event
			return nil
		}

		err := bus.Subscribe(ctx, "test.event", handler)
		require.NoError(t, err)

		event := testEvent{eventType: "test.event", data: "test data"}
		err = bus.Publish(ctx, event)
		require.NoError(t, err)

		select {
		case e := <-received:
			assert.Equal(t, event, e)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for event")
		}
	})

	t.Run("publish without subscribers", func(t *testing.T) {
		event := testEvent{eventType: "no.subscribers", data: "test"}
		err := bus.Publish(ctx, event)
		require.NoError(t, err)
	})

	t.Run("publish nil event", func(t *testing.T) {
		err := bus.Publish(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event cannot be nil")
	})

	t.Run("subscribe with empty event type", func(t *testing.T) {
		handler := func(ctx context.Context, event interface{}) error {
			return nil
		}
		err := bus.Subscribe(ctx, "", handler)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event type cannot be empty")
	})

	t.Run("subscribe with nil handler", func(t *testing.T) {
		err := bus.Subscribe(ctx, "test.event", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})
}

func TestInMemoryEventBus_MultipleSubscribers(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	received1 := make(chan interface{}, 1)
	received2 := make(chan interface{}, 1)

	handler1 := func(ctx context.Context, event interface{}) error {
		received1 <- event
		return nil
	}

	handler2 := func(ctx context.Context, event interface{}) error {
		received2 <- event
		return nil
	}

	bus.Subscribe(ctx, "test.event", handler1)
	bus.Subscribe(ctx, "test.event", handler2)

	event := testEvent{eventType: "test.event", data: "test"}
	err := bus.Publish(ctx, event)
	require.NoError(t, err)

	select {
	case <-received1:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler1")
	}

	select {
	case <-received2:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler2")
	}
}

func TestInMemoryEventBus_HandlerError(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	handler := func(ctx context.Context, event interface{}) error {
		return fmt.Errorf("handler error")
	}

	bus.Subscribe(ctx, "test.event", handler)

	event := testEvent{eventType: "test.event", data: "test"}
	err := bus.Publish(ctx, event)
	require.Error(t, err)
}

func TestInMemoryEventBus_Unsubscribe(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	received := make(chan interface{}, 1)
	handler := func(ctx context.Context, event interface{}) error {
		received <- event
		return nil
	}

	bus.Subscribe(ctx, "test.event", handler)
	bus.Unsubscribe(ctx, "test.event")

	event := testEvent{eventType: "test.event", data: "test"}
	err := bus.Publish(ctx, event)
	require.NoError(t, err)

	select {
	case <-received:
		t.Fatal("should not receive event after unsubscribe")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestInMemoryEventBus_PublishBatch(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	received := make(chan interface{}, 3)
	handler := func(ctx context.Context, event interface{}) error {
		received <- event
		return nil
	}

	bus.Subscribe(ctx, "test.event", handler)

	events := []interface{}{
		testEvent{eventType: "test.event", data: "1"},
		testEvent{eventType: "test.event", data: "2"},
		testEvent{eventType: "test.event", data: "3"},
	}

	err := bus.PublishBatch(ctx, events)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		select {
		case <-received:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for events")
		}
	}
}

func TestInMemoryEventBus_StartStop(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	err := bus.Start(ctx)
	require.NoError(t, err)

	err = bus.Stop(ctx)
	require.NoError(t, err)

	assert.Empty(t, bus.subscribers)
}

func TestInMemoryEventBus_PublishBatchError(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	handler := func(ctx context.Context, event interface{}) error {
		if e, ok := event.(testEvent); ok && e.data == "fail" {
			return fmt.Errorf("handler failed")
		}
		return nil
	}

	bus.Subscribe(ctx, "test.event", handler)

	events := []interface{}{
		testEvent{eventType: "test.event", data: "ok"},
		testEvent{eventType: "test.event", data: "fail"},
	}

	err := bus.PublishBatch(ctx, events)
	require.Error(t, err)
}

func TestInMemoryEventBus_PublishBatchEmpty(t *testing.T) {
	logger := mocks.NewMockLogger()
	bus := NewInMemoryEventBus(logger)
	ctx := context.Background()

	err := bus.PublishBatch(ctx, []interface{}{})
	require.NoError(t, err)
}
