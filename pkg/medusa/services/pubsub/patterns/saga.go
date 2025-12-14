package patterns

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// Saga Pattern

// SagaStep represents a step in a saga
type SagaStep struct {
	Name         string
	Action       func(ctx context.Context, data interface{}) error
	Compensation func(ctx context.Context, data interface{}) error
}

// Saga orchestrates a series of steps with compensation
type Saga struct {
	ID             string
	Steps          []*SagaStep
	CurrentStep    int
	State          SagaState
	Data           interface{}
	CompletedSteps []string
	FailedStep     string
	Error          error
	eventPublisher pubsub.Publisher
	mu             sync.Mutex
}

type SagaState int

const (
	SagaStatePending SagaState = iota
	SagaStateRunning
	SagaStateCompleted
	SagaStateFailed
	SagaStateCompensating
	SagaStateCompensated
)

func NewSaga(id string, publisher pubsub.Publisher) *Saga {
	return &Saga{
		ID:             id,
		Steps:          make([]*SagaStep, 0),
		State:          SagaStatePending,
		CompletedSteps: make([]string, 0),
		eventPublisher: publisher,
	}
}

func (s *Saga) AddStep(step *SagaStep) *Saga {
	s.Steps = append(s.Steps, step)
	return s
}

func (s *Saga) Execute(ctx context.Context, data interface{}) error {
	s.mu.Lock()
	s.State = SagaStateRunning
	s.Data = data
	s.mu.Unlock()

	for i, step := range s.Steps {
		s.CurrentStep = i

		if err := step.Action(ctx, data); err != nil {
			s.mu.Lock()
			s.State = SagaStateFailed
			s.FailedStep = step.Name
			s.Error = err
			s.mu.Unlock()

			// Compensate completed steps
			s.compensate(ctx, data)
			return fmt.Errorf("saga step %s failed: %w", step.Name, err)
		}

		s.CompletedSteps = append(s.CompletedSteps, step.Name)

		// Publish step completed event
		s.publishEvent(ctx, "saga.step.completed", map[string]interface{}{
			"saga_id": s.ID,
			"step":    step.Name,
		})
	}

	s.mu.Lock()
	s.State = SagaStateCompleted
	s.mu.Unlock()

	// Publish saga completed event
	s.publishEvent(ctx, "saga.completed", map[string]interface{}{
		"saga_id": s.ID,
	})

	return nil
}

func (s *Saga) compensate(ctx context.Context, data interface{}) error {
	s.mu.Lock()
	s.State = SagaStateCompensating
	s.mu.Unlock()

	// Compensate in reverse order
	for i := len(s.CompletedSteps) - 1; i >= 0; i-- {
		stepName := s.CompletedSteps[i]
		var step *SagaStep
		for _, st := range s.Steps {
			if st.Name == stepName {
				step = st
				break
			}
		}

		if step != nil && step.Compensation != nil {
			if err := step.Compensation(ctx, data); err != nil {
				// Log compensation error but continue
				s.publishEvent(ctx, "saga.compensation.failed", map[string]interface{}{
					"saga_id": s.ID,
					"step":    step.Name,
					"error":   err.Error(),
				})
			}
		}
	}

	s.mu.Lock()
	s.State = SagaStateCompensated
	s.mu.Unlock()

	s.publishEvent(ctx, "saga.compensated", map[string]interface{}{
		"saga_id": s.ID,
	})

	return nil
}

func (s *Saga) publishEvent(ctx context.Context, eventType string, data map[string]interface{}) {
	if s.eventPublisher == nil {
		return
	}

	payload, _ := json.Marshal(data)
	msg := &pubsub.Message{
		ID:        fmt.Sprintf("%s-%s-%d", s.ID, eventType, time.Now().UnixNano()),
		Topic:     eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	s.eventPublisher.Publish(ctx, eventType, msg)
}
