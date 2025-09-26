package utils

import (
	"compress/gzip"
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitBreakerOpen indicates the circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// Pool for commonly used objects to reduce GC pressure
var (
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 10)
		},
	}
	
	mapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}
)

// GetStringSlice gets a string slice from the pool
func GetStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(s []string) {
	s = s[:0] // Reset length but keep capacity
	stringSlicePool.Put(s)
}

// GetMap gets a map from the pool
func GetMap() map[string]interface{} {
	return mapPool.Get().(map[string]interface{})
}

// PutMap returns a map to the pool
func PutMap(m map[string]interface{}) {
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

// Debouncer helps prevent rapid successive calls
type Debouncer struct {
	timer   *time.Timer
	mutex   sync.Mutex
	delay   time.Duration
	running bool
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{delay: delay}
}

// Debounce executes the function after the delay, cancelling any previous calls
func (d *Debouncer) Debounce(f func()) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	if d.timer != nil {
		d.timer.Stop()
	}
	
	d.timer = time.AfterFunc(d.delay, f)
}

// BatchProcessor helps batch operations for better performance
type BatchProcessor struct {
	batchSize   int
	flushDelay  time.Duration
	processor   func([]interface{}) error
	items       []interface{}
	mutex       sync.Mutex
	timer       *time.Timer
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, flushDelay time.Duration, processor func([]interface{}) error) *BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BatchProcessor{
		batchSize:  batchSize,
		flushDelay: flushDelay,
		processor:  processor,
		items:      make([]interface{}, 0, batchSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Add adds an item to the batch
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	bp.items = append(bp.items, item)
	
	// Reset timer
	if bp.timer != nil {
		bp.timer.Stop()
	}
	bp.timer = time.AfterFunc(bp.flushDelay, bp.flush)
	
	// Check if batch is full
	if len(bp.items) >= bp.batchSize {
		bp.processImmediately()
	}
	
	return nil
}

// flush processes the current batch
func (bp *BatchProcessor) flush() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	bp.processImmediately()
}

// processImmediately processes items without delay
func (bp *BatchProcessor) processImmediately() {
	if len(bp.items) == 0 {
		return
	}
	
	// Copy items to process
	itemsToProcess := make([]interface{}, len(bp.items))
	copy(itemsToProcess, bp.items)
	bp.items = bp.items[:0] // Reset slice
	
	// Process asynchronously
	bp.wg.Add(1)
	go func() {
		defer bp.wg.Done()
		if err := bp.processor(itemsToProcess); err != nil {
			// Log error or handle as appropriate
			// Could add error callback here
		}
	}()
}

// Flush processes any remaining items
func (bp *BatchProcessor) Flush() {
	bp.mutex.Lock()
	if bp.timer != nil {
		bp.timer.Stop()
	}
	bp.processImmediately()
	bp.mutex.Unlock()
	
	bp.wg.Wait()
}

// Close shuts down the batch processor
func (bp *BatchProcessor) Close() {
	bp.cancel()
	bp.Flush()
}

// GzipCompressionLevel represents gzip compression levels
type GzipCompressionLevel int

const (
	GzipBestSpeed          GzipCompressionLevel = gzip.BestSpeed
	GzipBestCompression    GzipCompressionLevel = gzip.BestCompression
	GzipDefaultCompression GzipCompressionLevel = gzip.DefaultCompression
)

// CircuitBreaker implements the circuit breaker pattern for resilience
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	onStateChange func(string, string)
	
	mutex        sync.RWMutex
	failures     int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// Call executes the function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.RLock()
	state := cb.state
	lastFailTime := cb.lastFailTime
	cb.mutex.RUnlock()
	
	// Check if we should transition from open to half-open
	if state == "open" && time.Since(lastFailTime) > cb.resetTimeout {
		cb.mutex.Lock()
		cb.state = "half-open"
		cb.mutex.Unlock()
		state = "half-open"
	}
	
	// If open, reject immediately
	if state == "open" {
		return ErrCircuitBreakerOpen
	}
	
	// Execute function
	err := fn()
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		if cb.failures >= cb.maxFailures {
			cb.state = "open"
			if cb.onStateChange != nil {
				cb.onStateChange("closed", "open")
			}
		}
	} else {
		// Success - reset failures and close circuit
		if cb.state == "half-open" {
			cb.state = "closed"
			if cb.onStateChange != nil {
				cb.onStateChange("half-open", "closed")
			}
		}
		cb.failures = 0
	}
	
	return err
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}