package patterns

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Request-Reply Pattern

// RequestReplyBus handles request-reply messaging
type RequestReplyBus struct {
	broker         pubsub.MessageBroker
	pendingReplies map[string]chan *pubsub.Message
	mu             sync.RWMutex
}

func NewRequestReplyBus(broker pubsub.MessageBroker) *RequestReplyBus {
	return &RequestReplyBus{
		broker:         broker,
		pendingReplies: make(map[string]chan *pubsub.Message),
	}
}

func (rrb *RequestReplyBus) Request(ctx context.Context, topic string, request *pubsub.Message, timeout time.Duration) (*pubsub.Message, error) {
	replyChan := make(chan *pubsub.Message, 1)

	rrb.mu.Lock()
	rrb.pendingReplies[request.CorrelationID] = replyChan
	rrb.mu.Unlock()

	defer func() {
		rrb.mu.Lock()
		delete(rrb.pendingReplies, request.CorrelationID)
		rrb.mu.Unlock()
		close(replyChan)
	}()

	if err := rrb.broker.Publish(ctx, topic, request); err != nil {
		return nil, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, errors.New("request timeout")
	case reply := <-replyChan:
		return reply, nil
	}
}

func (rrb *RequestReplyBus) Reply(ctx context.Context, replyTo string, reply *pubsub.Message) error {
	rrb.mu.RLock()
	replyChan, exists := rrb.pendingReplies[reply.CorrelationID]
	rrb.mu.RUnlock()

	if exists {
		select {
		case replyChan <- reply:
			return nil
		default:
			return errors.New("reply channel full")
		}
	}

	return errors.New("no pending request found")
}
