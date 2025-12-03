package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisBackend(t *testing.T) (*RedisStreamBackend, func()) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	cfg := RedisConfig{
		Addr:     endpoint,
		Password: "",
		DB:       0,
	}

	backend, err := NewRedisStreamBackend(cfg)
	require.NoError(t, err)

	cleanup := func() {
		backend.Close()
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return backend, cleanup
}

func TestRedisStreamBackend_PublishAndSubscribe(t *testing.T) {
	backend, cleanup := setupRedisBackend(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream := "test-stream"
	consumerGroup := "test-group"

	event := Event{
		ID:        uuid.New().String(),
		Type:      "test.event",
		Payload:   map[string]interface{}{"message": "hello"},
		Metadata:  map[string]interface{}{"source": "test"},
		CreatedAt: time.Now(),
	}

	// Publish event directly to the test stream
	err := backend.PublishToStream(ctx, stream, event)
	require.NoError(t, err)

	// Subscribe using private method for testing
	received := make(chan Event, 1)
	handler := func(ctx context.Context, e Event) error {
		received <- e
		cancel() // Stop subscription after receiving
		return nil
	}

	go func() {
		err := backend.subscribeToStream(ctx, stream, consumerGroup, handler)
		if err != nil && err != context.Canceled {
			t.Logf("subscribe error: %v", err)
		}
	}()

	select {
	case receivedEvent := <-received:
		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.Equal(t, event.Type, receivedEvent.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestRedisStreamBackend_Replay(t *testing.T) {
	backend, cleanup := setupRedisBackend(t)
	defer cleanup()

	ctx := context.Background()
	stream := "replay-stream"

	// Publish multiple events
	events := make([]Event, 3)
	for i := 0; i < 3; i++ {
		events[i] = Event{
			ID:        uuid.New().String(),
			Type:      "replay.event",
			Payload:   map[string]interface{}{"index": i},
			Metadata:  map[string]interface{}{},
			CreatedAt: time.Now(),
		}
		err := backend.PublishToStream(ctx, stream, events[i])
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Replay from beginning using private method
	var replayed []Event
	handler := func(ctx context.Context, e Event) error {
		replayed = append(replayed, e)
		return nil
	}

	err := backend.replayFromTime(ctx, stream, time.Now().Add(-1*time.Hour), handler)
	require.NoError(t, err)
	assert.Len(t, replayed, 3)
}

func TestRedisStreamBackend_IdempotentDelivery(t *testing.T) {
	backend, cleanup := setupRedisBackend(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream := "idempotent-stream"
	consumerGroup := "idempotent-group"

	event := Event{
		ID:        uuid.New().String(),
		Type:      "idempotent.event",
		Payload:   map[string]interface{}{"data": "test"},
		Metadata:  map[string]interface{}{},
		CreatedAt: time.Now(),
	}

	err := backend.PublishToStream(ctx, stream, event)
	require.NoError(t, err)

	processedIDs := make(map[string]int)
	handler := func(ctx context.Context, e Event) error {
		processedIDs[e.ID]++
		if processedIDs[e.ID] >= 1 {
			cancel()
		}
		return nil
	}

	go func() {
		err := backend.subscribeToStream(ctx, stream, consumerGroup, handler)
		if err != nil && err != context.Canceled {
			t.Logf("subscribe error: %v", err)
		}
	}()

	<-ctx.Done()

	// Event should be processed exactly once due to acknowledgment
	assert.Equal(t, 1, processedIDs[event.ID])
}

func TestRedisStreamBackend_PublicMethods(t *testing.T) {
	backend, cleanup := setupRedisBackend(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test Publish method
	t.Run("Publish", func(t *testing.T) {
		payload := map[string]interface{}{"message": "test publish"}
		err := backend.Publish(ctx, payload)
		require.NoError(t, err)
	})

	// Test PublishBatch method
	t.Run("PublishBatch", func(t *testing.T) {
		events := []interface{}{
			map[string]interface{}{"batch": 1},
			map[string]interface{}{"batch": 2},
			map[string]interface{}{"batch": 3},
		}
		err := backend.PublishBatch(ctx, events)
		require.NoError(t, err)
	})

	// Test Subscribe method
	t.Run("Subscribe", func(t *testing.T) {
		receivedEvents := make(chan interface{}, 1)

		// Create a typed event
		type TestEvent struct {
			Data string
		}
		eventType := "map[string]interface {}" // This is how getEventType formats map types

		handler := func(ctx context.Context, event interface{}) error {
			receivedEvents <- event
			return nil
		}

		// Start subscriber in background
		subscribeCtx, subscribeCancel := context.WithCancel(ctx)
		defer subscribeCancel()

		go func() {
			_ = backend.Subscribe(subscribeCtx, eventType, handler)
		}()

		// Give subscriber time to set up and create consumer group
		time.Sleep(500 * time.Millisecond)

		// Publish an event that should be received
		payload := map[string]interface{}{"subscribe_test": "data"}
		err := backend.Publish(ctx, payload)
		require.NoError(t, err)

		// Wait for event (with timeout)
		select {
		case received := <-receivedEvents:
			require.NotNil(t, received)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for subscribed event")
		}

		subscribeCancel()
	})
}
