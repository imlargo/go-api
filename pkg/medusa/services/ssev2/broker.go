package ssev2

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Broker manages SSE connections and message routing
type Broker struct {
	clients       map[string]*Client
	subscriptions map[string]map[string]bool // topic -> clientID -> exists
	mu            sync.RWMutex
	eventStore    EventStore
	pubsubBroker  pubsub.MessageBroker

	// Configuration
	config BrokerConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// BrokerConfig holds broker configuration
type BrokerConfig struct {
	BufferSize        int
	ClientTimeout     time.Duration
	HeartbeatInterval time.Duration
	EnableReplay      bool
}

// DefaultConfig returns default broker configuration
func DefaultConfig() BrokerConfig {
	return BrokerConfig{
		BufferSize:        100,
		ClientTimeout:     time.Minute * 5,
		HeartbeatInterval: time.Second * 30,
		EnableReplay:      true,
	}
}

// NewBroker creates a new SSE broker
func NewBroker(config BrokerConfig) *Broker {
	ctx, cancel := context.WithCancel(context.Background())

	b := &Broker{
		clients:       make(map[string]*Client),
		subscriptions: make(map[string]map[string]bool),
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
	}

	b.wg.Add(1)
	go b.heartbeat()

	return b
}

// SetEventStore sets the event store for persistence
func (b *Broker) SetEventStore(store EventStore) {
	b.eventStore = store
}

// eventToMessage converts an SSE Event to a pubsub Message
func eventToMessage(topic string, event Event) (*pubsub.Message, error) {
	// Marshal event data to JSON
	payload, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	msg := &pubsub.Message{
		ID:          event.ID,
		Topic:       topic,
		Payload:     payload,
		Timestamp:   time.Now(),
		ContentType: "application/json",
		Headers:     make(map[string]string),
	}

	// Store event type in headers
	if event.Event != "" {
		msg.Headers["event-type"] = event.Event
	}
	if event.Retry > 0 {
		msg.Headers["retry"] = strconv.Itoa(event.Retry)
	}

	return msg, nil
}

// messageToEvent converts a pubsub Message to an SSE Event
func messageToEvent(msg *pubsub.Message) (Event, error) {
	event := Event{
		ID: msg.ID,
	}

	// Extract event type from headers
	if eventType, ok := msg.Headers["event-type"]; ok {
		event.Event = eventType
	}
	if retryStr, ok := msg.Headers["retry"]; ok {
		if retry, err := strconv.Atoi(retryStr); err == nil {
			event.Retry = retry
		}
		// Silently ignore parse errors - retry is optional and we gracefully degrade
	}

	// Unmarshal payload to data
	var data interface{}
	if err := json.Unmarshal(msg.Payload, &data); err != nil {
		// If unmarshal fails, use raw bytes as string
		event.Data = string(msg.Payload)
	} else {
		event.Data = data
	}

	return event, nil
}

// SetPubSub sets the pubsub system for distributed messaging
func (b *Broker) SetPubSub(ps pubsub.MessageBroker) error {
	b.pubsubBroker = ps

	// Subscribe to all topics this broker manages using wildcard pattern
	// Note: The wildcard pattern "*" may not be supported by all message broker implementations.
	// For RabbitMQ, this requires proper exchange configuration (e.g., topic exchange with "#" pattern).
	// If wildcard subscription fails, topics must be subscribed to individually via SubscribeToTopic.
	err := ps.Subscribe(b.ctx, "*", func(ctx context.Context, msg *pubsub.Message) error {
		event, err := messageToEvent(msg)
		if err != nil {
			return fmt.Errorf("failed to convert message to event: %w", err)
		}

		b.mu.RLock()
		clients := b.subscriptions[msg.Topic]
		b.mu.RUnlock()

		for clientID := range clients {
			b.mu.RLock()
			client, exists := b.clients[clientID]
			b.mu.RUnlock()

			if exists {
				select {
				case client.Channel <- event:
				case <-time.After(time.Second):
					// Client is slow, skip
				}
			}
		}
		return nil
	})
	
	// If wildcard subscription fails, return error but keep broker reference
	// Caller should subscribe to specific topics using SubscribeToTopic
	if err != nil {
		return fmt.Errorf("wildcard subscription failed (broker-specific topics must be used): %w", err)
	}
	
	return nil
}

// SubscribeToTopic subscribes to a specific topic when wildcard is not supported
func (b *Broker) SubscribeToTopic(topic string) error {
	if b.pubsubBroker == nil {
		return fmt.Errorf("pubsub broker not configured")
	}
	
	return b.pubsubBroker.Subscribe(b.ctx, topic, func(ctx context.Context, msg *pubsub.Message) error {
		event, err := messageToEvent(msg)
		if err != nil {
			return fmt.Errorf("failed to convert message to event: %w", err)
		}

		b.mu.RLock()
		clients := b.subscriptions[msg.Topic]
		b.mu.RUnlock()

		for clientID := range clients {
			b.mu.RLock()
			client, exists := b.clients[clientID]
			b.mu.RUnlock()

			if exists {
				select {
				case client.Channel <- event:
				case <-time.After(time.Second):
					// Client is slow, skip
				}
			}
		}
		return nil
	})
}

// Subscribe creates a new client subscription
func (b *Broker) Subscribe(clientID string, topics []string, lastEventID string) (*Client, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if client already exists
	if client, exists := b.clients[clientID]; exists {
		return client, nil
	}

	ctx, cancel := context.WithCancel(b.ctx)
	client := &Client{
		ID:       clientID,
		Channel:  make(chan Event, b.config.BufferSize),
		ctx:      ctx,
		cancel:   cancel,
		lastSeen: time.Now(),
	}

	b.clients[clientID] = client

	// Subscribe to topics
	for _, topic := range topics {
		if b.subscriptions[topic] == nil {
			b.subscriptions[topic] = make(map[string]bool)
		}
		b.subscriptions[topic][clientID] = true
	}

	// Replay missed events if enabled
	if b.config.EnableReplay && lastEventID != "" && b.eventStore != nil {
		go b.replayEvents(client, topics, lastEventID)
	}

	return client, nil
}

// Unsubscribe removes a client subscription
func (b *Broker) Unsubscribe(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	client, exists := b.clients[clientID]
	if !exists {
		return
	}

	client.cancel()
	close(client.Channel)
	delete(b.clients, clientID)

	// Remove from all topic subscriptions
	for topic := range b.subscriptions {
		delete(b.subscriptions[topic], clientID)
		if len(b.subscriptions[topic]) == 0 {
			delete(b.subscriptions, topic)
		}
	}
}

// Publish sends an event to all subscribers of a topic
func (b *Broker) Publish(topic string, event Event) error {
	// Save to event store if available
	if b.eventStore != nil {
		if err := b.eventStore.Save(topic, event); err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}

	// Publish to distributed system if available
	if b.pubsubBroker != nil {
		msg, err := eventToMessage(topic, event)
		if err != nil {
			return fmt.Errorf("failed to convert event to message: %w", err)
		}

		if err := b.pubsubBroker.Publish(b.ctx, topic, msg); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
		return nil
	}

	// Local publish
	b.mu.RLock()
	clients := b.subscriptions[topic]
	b.mu.RUnlock()

	for clientID := range clients {
		b.mu.RLock()
		client, exists := b.clients[clientID]
		b.mu.RUnlock()

		if exists {
			select {
			case client.Channel <- event:
				client.lastSeen = time.Now()
			case <-time.After(time.Second):
				// Client is slow, consider removing
			}
		}
	}

	return nil
}

// replayEvents replays missed events to a client
func (b *Broker) replayEvents(client *Client, topics []string, lastEventID string) {
	for _, topic := range topics {
		events, err := b.eventStore.GetSince(topic, lastEventID)
		if err != nil {
			continue
		}

		for _, event := range events {
			select {
			case client.Channel <- event:
			case <-client.ctx.Done():
				return
			}
		}
	}
}

// heartbeat sends periodic heartbeat messages
func (b *Broker) heartbeat() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.sendHeartbeat()
		case <-b.ctx.Done():
			return
		}
	}
}

// sendHeartbeat sends a heartbeat to all connected clients
func (b *Broker) sendHeartbeat() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	heartbeat := Event{
		Event: "heartbeat",
		Data:  map[string]interface{}{"timestamp": time.Now().Unix()},
	}

	for _, client := range b.clients {
		// Check for timeout
		if time.Since(client.lastSeen) > b.config.ClientTimeout {
			go b.Unsubscribe(client.ID)
			continue
		}

		select {
		case client.Channel <- heartbeat:
		default:
			// Channel full, skip
		}
	}
}

// Close shuts down the broker
func (b *Broker) Close() error {
	b.cancel()
	b.wg.Wait()

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, client := range b.clients {
		client.cancel()
		close(client.Channel)
	}

	if b.pubsubBroker != nil {
		return b.pubsubBroker.Close()
	}

	return nil
}

// ClientCount returns the number of connected clients
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
