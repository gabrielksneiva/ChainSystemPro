package eventbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
)

// InMemoryEventBus is an in-memory implementation of EventBus
type InMemoryEventBus struct {
	subscribers map[string][]ports.EventHandler
	mu          sync.RWMutex
	logger      ports.Logger
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus(logger ports.Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[string][]ports.EventHandler),
		logger:      logger,
	}
}

// Publish publishes an event to all subscribers
func (bus *InMemoryEventBus) Publish(ctx context.Context, event interface{}) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventType := getEventType(event)

	bus.mu.RLock()
	handlers, exists := bus.subscribers[eventType]
	bus.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		bus.logger.Debug("no subscribers for event type", map[string]interface{}{
			"event_type": eventType,
		})
		return nil
	}

	bus.logger.Debug("publishing event", map[string]interface{}{
		"event_type":       eventType,
		"subscriber_count": len(handlers),
	})

	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h ports.EventHandler) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				errChan <- err
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		bus.logger.Error("errors occurred while publishing event", nil, map[string]interface{}{
			"event_type":  eventType,
			"error_count": len(errors),
		})
		return fmt.Errorf("failed to publish event to %d handler(s)", len(errors))
	}

	return nil
}

// PublishBatch publishes multiple events
func (bus *InMemoryEventBus) PublishBatch(ctx context.Context, events []interface{}) error {
	if len(events) == 0 {
		return nil
	}

	bus.logger.Debug("publishing batch of events", map[string]interface{}{
		"count": len(events),
	})

	for _, event := range events {
		if err := bus.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event in batch: %w", err)
		}
	}

	return nil
}

// Subscribe subscribes to events of a specific type
func (bus *InMemoryEventBus) Subscribe(ctx context.Context, eventType string, handler ports.EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)

	bus.logger.Info("subscribed to event type", map[string]interface{}{
		"event_type": eventType,
	})

	return nil
}

// Unsubscribe unsubscribes from events of a specific type
func (bus *InMemoryEventBus) Unsubscribe(ctx context.Context, eventType string) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	delete(bus.subscribers, eventType)

	bus.logger.Info("unsubscribed from event type", map[string]interface{}{
		"event_type": eventType,
	})

	return nil
}

// Start starts the event bus
func (bus *InMemoryEventBus) Start(ctx context.Context) error {
	bus.logger.Info("event bus started", map[string]interface{}{})
	return nil
}

// Stop stops the event bus
func (bus *InMemoryEventBus) Stop(ctx context.Context) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.subscribers = make(map[string][]ports.EventHandler)

	bus.logger.Info("event bus stopped", map[string]interface{}{})
	return nil
}

// getEventType extracts the event type from an event
func getEventType(event interface{}) string {
	type eventTyper interface {
		Type() string
	}

	if e, ok := event.(eventTyper); ok {
		return e.Type()
	}

	return fmt.Sprintf("%T", event)
}
