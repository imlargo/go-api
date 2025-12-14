package pubsub

import (
	"context"
	"encoding/json"
	"time"
)

// Message represents a message in the pub/sub system
type Message struct {
	ID          string                 `json:"id"`
	Topic       string                 `json:"topic"`
	Payload     []byte                 `json:"payload"`
	Headers     map[string]string      `json:"headers"`
	Timestamp   time.Time              `json:"timestamp"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata"`

	// Internal fields for message handling
	ackFunc  func() error
	nackFunc func(requeue bool) error
}

// Ack acknowledges successful processing of the message
func (m *Message) Ack() error {
	if m.ackFunc != nil {
		return m.ackFunc()
	}
	return nil
}

// Nack negatively acknowledges the message (reject)
func (m *Message) Nack(requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(requeue)
	}
	return nil
}

// DecodeJSON unmarshals the payload into the provided interface
func (m *Message) DecodeJSON(v interface{}) error {
	return json.Unmarshal(m.Payload, v)
}

// EncodeJSON marshals the provided interface into the payload
func (m *Message) EncodeJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Payload = data
	m.ContentType = "application/json"
	return nil
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, msg *Message) error

// NewMessage creates a new message with the given payload
func NewMessage(payload []byte) *Message {
	return &Message{
		ID:        generateID(),
		Payload:   payload,
		Timestamp: time.Now(),
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
	}
}

// NewJSONMessage creates a message from a JSON-encodable value
func NewJSONMessage(v interface{}) (*Message, error) {
	msg := NewMessage(nil)
	if err := msg.EncodeJSON(v); err != nil {
		return nil, err
	}
	return msg, nil
}

// SetAckFunc and SetNackFunc are helper methods for Message
func (m *Message) SetAckFunc(fn func() error) {
	m.ackFunc = fn
}

func (m *Message) SetNackFunc(fn func(bool) error) {
	m.nackFunc = fn
}
