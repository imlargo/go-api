package patterns

import (
	"sync"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Priority Queue Pattern

// PriorityMessage wraps a message with priority
type PriorityMessage struct {
	Message  *pubsub.Message
	Priority int
}

// PriorityQueue implements a priority queue for messages
type PriorityQueue struct {
	items []*PriorityMessage
	mu    sync.Mutex
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*PriorityMessage, 0),
	}
}

func (pq *PriorityQueue) Enqueue(msg *PriorityMessage) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.items = append(pq.items, msg)
	pq.bubbleUp(len(pq.items) - 1)
}

func (pq *PriorityQueue) Dequeue() *PriorityMessage {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	item := pq.items[0]
	lastIdx := len(pq.items) - 1
	pq.items[0] = pq.items[lastIdx]
	pq.items = pq.items[:lastIdx]

	if len(pq.items) > 0 {
		pq.bubbleDown(0)
	}

	return item
}

func (pq *PriorityQueue) bubbleUp(idx int) {
	for idx > 0 {
		parentIdx := (idx - 1) / 2
		if pq.items[idx].Priority <= pq.items[parentIdx].Priority {
			break
		}
		pq.items[idx], pq.items[parentIdx] = pq.items[parentIdx], pq.items[idx]
		idx = parentIdx
	}
}

func (pq *PriorityQueue) bubbleDown(idx int) {
	for {
		leftChild := 2*idx + 1
		rightChild := 2*idx + 2
		largest := idx

		if leftChild < len(pq.items) && pq.items[leftChild].Priority > pq.items[largest].Priority {
			largest = leftChild
		}
		if rightChild < len(pq.items) && pq.items[rightChild].Priority > pq.items[largest].Priority {
			largest = rightChild
		}
		if largest == idx {
			break
		}

		pq.items[idx], pq.items[largest] = pq.items[largest], pq.items[idx]
		idx = largest
	}
}
