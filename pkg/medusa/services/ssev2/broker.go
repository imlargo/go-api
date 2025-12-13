package ssev2

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Broker manages SSE connections and message routing
type Broker struct {
	clients       map[string]*Client
	subscriptions map[string]map[string]bool // topic -> clientID -> exists
	mu            sync.RWMutex
	eventStore    EventStore
	pubsub        PubSub

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

// SetPubSub sets the pubsub system for distributed messaging
func (b *Broker) SetPubSub(ps PubSub) error {
	b.pubsub = ps

	// Subscribe to all topics this broker manages
	return ps.Subscribe([]string{"*"}, func(topic string, event Event) {
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
				case <-time.After(time.Second):
					// Client is slow, skip
				}
			}
		}
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
	if b.pubsub != nil {
		if err := b.pubsub.Publish(topic, event); err != nil {
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

	if b.pubsub != nil {
		return b.pubsub.Close()
	}

	return nil
}

// ClientCount returns the number of connected clients
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
