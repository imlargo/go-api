package ssev2

import "sync"

// EventStore defines the interface for event persistence
type EventStore interface {
	Save(topic string, event Event) error
	GetSince(topic string, lastEventID string) ([]Event, error)
	Clear(topic string) error
}

// MemoryEventStore implements EventStore with in-memory storage
type MemoryEventStore struct {
	events  map[string][]Event
	mu      sync.RWMutex
	maxSize int
}

// NewMemoryEventStore creates a new in-memory event store
func NewMemoryEventStore(maxSize int) *MemoryEventStore {
	return &MemoryEventStore{
		events:  make(map[string][]Event),
		maxSize: maxSize,
	}
}

func (m *MemoryEventStore) Save(topic string, event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.events[topic] == nil {
		m.events[topic] = make([]Event, 0)
	}

	m.events[topic] = append(m.events[topic], event)

	// Trim if exceeds max size
	if len(m.events[topic]) > m.maxSize {
		m.events[topic] = m.events[topic][len(m.events[topic])-m.maxSize:]
	}

	return nil
}

func (m *MemoryEventStore) GetSince(topic string, lastEventID string) ([]Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := m.events[topic]
	if events == nil {
		return []Event{}, nil
	}

	// Find the index of lastEventID
	startIdx := 0
	for i, event := range events {
		if event.ID == lastEventID {
			startIdx = i + 1
			break
		}
	}

	if startIdx >= len(events) {
		return []Event{}, nil
	}

	return events[startIdx:], nil
}

func (m *MemoryEventStore) Clear(topic string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.events, topic)
	return nil
}
