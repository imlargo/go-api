package patterns

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// CQRS Pattern

// Command represents a command in CQRS
type Command struct {
	ID          string
	Type        string
	AggregateID string
	Data        json.RawMessage
	Metadata    map[string]interface{}
	Timestamp   time.Time
}

// CommandHandler handles commands
type CommandHandler interface {
	Handle(ctx context.Context, command *Command) error
	CommandType() string
}

// CommandBus routes commands to handlers
type CommandBus struct {
	handlers map[string]CommandHandler
	broker   pubsub.MessageBroker
	mu       sync.RWMutex
}

func NewCommandBus(broker pubsub.MessageBroker) *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandler),
		broker:   broker,
	}
}

func (cb *CommandBus) RegisterHandler(handler CommandHandler) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.handlers[handler.CommandType()] = handler
}

func (cb *CommandBus) Send(ctx context.Context, command *Command) error {
	cb.mu.RLock()
	handler, exists := cb.handlers[command.Type]
	cb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for command type: %s", command.Type)
	}

	return handler.Handle(ctx, command)
}

func (cb *CommandBus) SendAsync(ctx context.Context, command *Command) error {
	data, err := json.Marshal(command)
	if err != nil {
		return err
	}

	msg := &pubsub.Message{
		ID:            command.ID,
		Topic:         "commands." + command.Type,
		Payload:       data,
		CorrelationID: command.AggregateID,
		Timestamp:     command.Timestamp,
		ContentType:   "application/json",
	}

	return cb.broker.Publish(ctx, msg.Topic, msg)
}

// Query represents a query in CQRS
type Query struct {
	ID        string
	Type      string
	Params    map[string]interface{}
	Timestamp time.Time
}

// QueryHandler handles queries
type QueryHandler interface {
	Handle(ctx context.Context, query *Query) (interface{}, error)
	QueryType() string
}

// QueryBus routes queries to handlers
type QueryBus struct {
	handlers map[string]QueryHandler
	mu       sync.RWMutex
}

func NewQueryBus() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

func (qb *QueryBus) RegisterHandler(handler QueryHandler) {
	qb.mu.Lock()
	defer qb.mu.Unlock()
	qb.handlers[handler.QueryType()] = handler
}

func (qb *QueryBus) Execute(ctx context.Context, query *Query) (interface{}, error) {
	qb.mu.RLock()
	handler, exists := qb.handlers[query.Type]
	qb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for query type: %s", query.Type)
	}

	return handler.Handle(ctx, query)
}
