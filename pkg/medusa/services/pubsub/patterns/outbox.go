package patterns

import (
	"context"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Outbox Pattern

// OutboxMessage represents a message in the outbox
type OutboxMessage struct {
	ID            string
	Topic         string
	Payload       []byte
	Headers       map[string]string
	Status        OutboxStatus
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	Attempts      int
	LastAttemptAt *time.Time
	Error         string
}

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusProcessed OutboxStatus = "processed"
	OutboxStatusFailed    OutboxStatus = "failed"
)

// OutboxStore stores outbox messages
type OutboxStore interface {
	Add(ctx context.Context, msg *OutboxMessage) error
	GetPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, err error) error
	IncrementAttempts(ctx context.Context, id string) error
}

// OutboxProcessor processes outbox messages
type OutboxProcessor struct {
	store     OutboxStore
	publisher pubsub.Publisher
	interval  time.Duration
	batchSize int
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewOutboxProcessor(store OutboxStore, publisher pubsub.Publisher, interval time.Duration, batchSize int) *OutboxProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &OutboxProcessor{
		store:     store,
		publisher: publisher,
		interval:  interval,
		batchSize: batchSize,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (op *OutboxProcessor) Start() {
	op.wg.Add(1)
	go op.process()
}

func (op *OutboxProcessor) Stop() {
	op.cancel()
	op.wg.Wait()
}

func (op *OutboxProcessor) process() {
	defer op.wg.Done()

	ticker := time.NewTicker(op.interval)
	defer ticker.Stop()

	for {
		select {
		case <-op.ctx.Done():
			return
		case <-ticker.C:
			op.processBatch()
		}
	}
}

func (op *OutboxProcessor) processBatch() {
	messages, err := op.store.GetPending(op.ctx, op.batchSize)
	if err != nil {
		return
	}

	for _, outboxMsg := range messages {
		op.store.IncrementAttempts(op.ctx, outboxMsg.ID)

		msg := &pubsub.Message{
			ID:        outboxMsg.ID,
			Topic:     outboxMsg.Topic,
			Payload:   outboxMsg.Payload,
			Headers:   outboxMsg.Headers,
			Timestamp: time.Now(),
		}

		err := op.publisher.Publish(op.ctx, outboxMsg.Topic, msg)
		if err != nil {
			op.store.MarkFailed(op.ctx, outboxMsg.ID, err)
		} else {
			op.store.MarkProcessed(op.ctx, outboxMsg.ID)
		}
	}
}
