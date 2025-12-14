package patterns

import (
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Batch Processing Pattern

// BatchProcessor processes messages in batches
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	handler   func([]*pubsub.Message) error
	buffer    []*pubsub.Message
	timer     *time.Timer
	mu        sync.Mutex
}

func NewBatchProcessor(batchSize int, timeout time.Duration, handler func([]*pubsub.Message) error) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		handler:   handler,
		buffer:    make([]*pubsub.Message, 0, batchSize),
	}
}

func (bp *BatchProcessor) Add(msg *pubsub.Message) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, msg)

	if bp.timer == nil {
		bp.timer = time.AfterFunc(bp.timeout, bp.flush)
	}

	if len(bp.buffer) >= bp.batchSize {
		return bp.flushUnlocked()
	}

	return nil
}

func (bp *BatchProcessor) flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.flushUnlocked()
}

func (bp *BatchProcessor) flushUnlocked() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	if bp.timer != nil {
		bp.timer.Stop()
		bp.timer = nil
	}

	batch := bp.buffer
	bp.buffer = make([]*pubsub.Message, 0, bp.batchSize)

	return bp.handler(batch)
}
