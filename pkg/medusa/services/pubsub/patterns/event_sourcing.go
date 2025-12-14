package patterns

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Event Sourcing Pattern

// Event represents a domain event
type Event struct {
	ID            string
	AggregateID   string
	AggregateType string
	EventType     string
	Version       int
	Data          json.RawMessage
	Metadata      map[string]interface{}
	Timestamp     time.Time
}

// EventStore stores and retrieves events
type EventStore interface {
	AppendEvent(ctx context.Context, event *Event) error
	GetEvents(ctx context.Context, aggregateID string) ([]*Event, error)
	GetEventsByType(ctx context.Context, eventType string) ([]*Event, error)
	Subscribe(ctx context.Context, eventType string, handler func(*Event) error) error
}

// EventSourcedAggregate is an aggregate that uses event sourcing
type EventSourcedAggregate interface {
	AggregateID() string
	Version() int
	ApplyEvent(event *Event) error
	GetUncommittedEvents() []*Event
	MarkEventsAsCommitted()
}

// EventBus publishes and subscribes to events
type EventBus struct {
	broker    pubsub.MessageBroker
	store     EventStore
	publisher pubsub.Publisher
}

func NewEventBus(broker pubsub.MessageBroker, store EventStore) *EventBus {
	return &EventBus{
		broker:    broker,
		store:     store,
		publisher: broker,
	}
}

func (eb *EventBus) PublishEvent(ctx context.Context, event *Event) error {
	// Store event
	if err := eb.store.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Publish to message bus
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &pubsub.Message{
		ID:            event.ID,
		Topic:         event.EventType,
		Payload:       data,
		CorrelationID: event.AggregateID,
		Timestamp:     event.Timestamp,
		ContentType:   "application/json",
		Headers: map[string]string{
			"x-aggregate-id":   event.AggregateID,
			"x-aggregate-type": event.AggregateType,
			"x-event-version":  fmt.Sprintf("%d", event.Version),
		},
	}

	return eb.publisher.Publish(ctx, event.EventType, msg)
}

func (eb *EventBus) SubscribeToEvent(ctx context.Context, eventType string, handler func(*Event) error) error {
	return eb.broker.Subscribe(ctx, eventType, func(ctx context.Context, msg *pubsub.Message) error {
		var event Event
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			return err
		}
		return handler(&event)
	})
}
