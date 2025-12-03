package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const eventsStream = "events"

// Event represents an event structure for Redis
type Event struct {
	ID        string
	Type      string
	Payload   map[string]interface{}
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

// RedisStreamBackend implements EventBus using Redis Streams
type RedisStreamBackend struct {
	client *redis.Client
	config RedisConfig
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisStreamBackend creates a new Redis Streams backend
func NewRedisStreamBackend(cfg RedisConfig) (*RedisStreamBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStreamBackend{
		client: client,
		config: cfg,
	}, nil
}

// Publish publishes an event to all subscribers (implements ports.EventBus)
func (r *RedisStreamBackend) Publish(ctx context.Context, event interface{}) error {
	stream := eventsStream

	// Convert event to Event structure
	evt := Event{
		ID:        uuid.New().String(),
		Type:      getEventType(event),
		Payload:   map[string]interface{}{"data": event},
		Metadata:  map[string]interface{}{},
		CreatedAt: time.Now(),
	}

	return r.PublishToStream(ctx, stream, evt)
}

// PublishBatch publishes multiple events (implements ports.EventBus)
func (r *RedisStreamBackend) PublishBatch(ctx context.Context, events []interface{}) error {
	for _, event := range events {
		if err := r.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes to events of a specific type (implements ports.EventBus)
func (r *RedisStreamBackend) Subscribe(ctx context.Context, eventType string, handler ports.EventHandler) error {
	stream := "events"
	consumerGroup := fmt.Sprintf("group-%s", eventType)

	return r.subscribeToStream(ctx, stream, consumerGroup, func(ctx context.Context, evt Event) error {
		if evt.Type == eventType {
			return handler(ctx, evt.Payload["data"])
		}
		return nil
	})
}

// publishToStream publishes an event to a Redis stream
// Exported as PublishToStream for testing purposes
func (r *RedisStreamBackend) PublishToStream(ctx context.Context, stream string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	values := map[string]interface{}{
		"id":         event.ID,
		"type":       event.Type,
		"payload":    data,
		"metadata":   mustMarshalJSON(event.Metadata),
		"created_at": event.CreatedAt.Format(time.RFC3339Nano),
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish event to Redis: %w", err)
	}

	return nil
}

type eventHandlerFunc func(context.Context, Event) error

// subscribeToStream subscribes to a stream and processes events
func (r *RedisStreamBackend) subscribeToStream(ctx context.Context, stream, consumerGroup string, handler eventHandlerFunc) error {
	// Create consumer group if it doesn't exist
	err := r.client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	consumerName := fmt.Sprintf("consumer-%s", uuid.New().String())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read from stream
			streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    consumerGroup,
				Consumer: consumerName,
				Streams:  []string{stream, ">"},
				Count:    10,
				Block:    time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue
				}
				return fmt.Errorf("failed to read from stream: %w", err)
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					event, err := r.parseMessage(message)
					if err != nil {
						// Log error but continue processing
						continue
					}

					if err := handler(ctx, event); err != nil {
						// Nack and continue
						continue
					}

					// Acknowledge message
					r.client.XAck(ctx, stream.Stream, consumerGroup, message.ID)
				}
			}
		}
	}
}

// replayFromTime replays events from a specific point in time (internal helper)
func (r *RedisStreamBackend) replayFromTime(ctx context.Context, stream string, fromTime time.Time, handler eventHandlerFunc) error {
	// Convert time to stream ID (milliseconds-seqno)
	startID := fmt.Sprintf("%d-0", fromTime.UnixMilli())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			streams, err := r.client.XRead(ctx, &redis.XReadArgs{
				Streams: []string{stream, startID},
				Count:   100,
				Block:   time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					return nil // No more messages
				}
				return fmt.Errorf("failed to read from stream: %w", err)
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					event, err := r.parseMessage(message)
					if err != nil {
						continue
					}

					if err := handler(ctx, event); err != nil {
						return fmt.Errorf("handler error during replay: %w", err)
					}

					startID = message.ID
				}
			}

			if len(streams) == 0 || len(streams[0].Messages) < 100 {
				return nil // Finished replay
			}
		}
	}
}

// Close closes the Redis connection
func (r *RedisStreamBackend) Close() error {
	return r.client.Close()
}

func (r *RedisStreamBackend) parseMessage(msg redis.XMessage) (Event, error) {
	var event Event

	idStr, ok := msg.Values["id"].(string)
	if !ok {
		return event, fmt.Errorf("invalid event ID")
	}
	event.ID = idStr

	eventType, ok := msg.Values["type"].(string)
	if !ok {
		return event, fmt.Errorf("invalid event type")
	}
	event.Type = eventType

	payloadBytes, ok := msg.Values["payload"].(string)
	if !ok {
		return event, fmt.Errorf("invalid payload")
	}

	if err := json.Unmarshal([]byte(payloadBytes), &event); err != nil {
		return event, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	metadataStr, _ := msg.Values["metadata"].(string)
	if metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			event.Metadata = metadata
		}
	}

	createdAtStr, ok := msg.Values["created_at"].(string)
	if ok {
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err == nil {
			event.CreatedAt = createdAt
		}
	}

	return event, nil
}

func mustMarshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
