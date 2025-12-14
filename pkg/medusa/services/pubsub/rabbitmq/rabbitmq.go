// Package rabbitmq provides RabbitMQ implementation of the pubsub interfaces
package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub/observability"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub/serializers"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Config holds RabbitMQ connection configuration
type Config struct {
	URL               string
	ConnectionName    string
	ReconnectInterval time.Duration
	ReconnectAttempts int
	Heartbeat         time.Duration
	ConnectionTimeout time.Duration
	ChannelPoolSize   int
	PrefetchCount     int
	PrefetchSize      int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		URL:               "amqp://guest:guest@localhost:5672/",
		ConnectionName:    "pubsub-client",
		ReconnectInterval: 5 * time.Second,
		ReconnectAttempts: 10,
		Heartbeat:         30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		ChannelPoolSize:   10,
		PrefetchCount:     10,
		PrefetchSize:      0,
	}
}

// Broker implements MessageBroker interface for RabbitMQ
type Broker struct {
	config      *Config
	conn        *amqp.Connection
	connMu      sync.RWMutex
	channels    chan *amqp.Channel
	closeChan   chan *amqp.Error
	subscribers map[string]*subscription
	subMu       sync.RWMutex
	serializer  pubsub.Serializer
	logger      *logger.Logger
	metrics     pubsub.Metrics
	middleware  pubsub.MiddlewareChain
	topology    *pubsub.TopologyConfig
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	connected   bool
}

type subscription struct {
	topic    string
	handler  pubsub.MessageHandler
	options  *pubsub.SubscribeOptions
	cancel   context.CancelFunc
	delivery <-chan amqp.Delivery
}

// NewBroker creates a new RabbitMQ broker instance
func NewBroker(config *Config, opts ...BrokerOption) (*Broker, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := &Broker{
		config:      config,
		subscribers: make(map[string]*subscription),
		channels:    make(chan *amqp.Channel, config.ChannelPoolSize),
		serializer:  &serializers.JSONSerializer{},
		logger:      logger.NewLogger(),
		metrics:     &observability.NoopMetrics{},
		ctx:         ctx,
		cancel:      cancel,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

// BrokerOption configures the broker
type BrokerOption func(*Broker)

// WithSerializer sets the message serializer
func WithSerializer(s pubsub.Serializer) BrokerOption {
	return func(b *Broker) {
		b.serializer = s
	}
}

// WithLogger sets the logger
func WithLogger(l *logger.Logger) BrokerOption {
	return func(b *Broker) {
		b.logger = l
	}
}

// WithMetrics sets the metrics collector
func WithMetrics(m pubsub.Metrics) BrokerOption {
	return func(b *Broker) {
		b.metrics = m
	}
}

// WithMiddleware adds global middleware
func WithMiddleware(mw ...pubsub.Middleware) BrokerOption {
	return func(b *Broker) {
		b.middleware = append(b.middleware, mw...)
	}
}

// WithTopology sets topology configuration
func WithTopology(topology *pubsub.TopologyConfig) BrokerOption {
	return func(b *Broker) {
		b.topology = topology
	}
}

// Connect establishes connection to RabbitMQ
func (b *Broker) Connect(ctx context.Context) error {
	b.connMu.Lock()
	defer b.connMu.Unlock()

	if b.connected {
		return nil
	}

	amqpConfig := amqp.Config{
		Heartbeat: b.config.Heartbeat,
		Locale:    "en_US",
		Properties: amqp.Table{
			"connection_name": b.config.ConnectionName,
		},
	}

	conn, err := amqp.DialConfig(b.config.URL, amqpConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	b.conn = conn
	b.closeChan = make(chan *amqp.Error)
	b.conn.NotifyClose(b.closeChan)
	b.connected = true

	// Initialize channel pool
	for i := 0; i < b.config.ChannelPoolSize; i++ {
		ch, err := b.conn.Channel()
		if err != nil {
			b.conn.Close()
			return fmt.Errorf("failed to create channel: %w", err)
		}
		b.channels <- ch
	}

	// Setup topology if configured
	if b.topology != nil {
		if err := b.setupTopology(); err != nil {
			b.conn.Close()
			return fmt.Errorf("failed to setup topology: %w", err)
		}
	}

	// Start connection monitor
	b.wg.Add(1)
	go b.monitorConnection()

	b.logger.Sugar().Info("Connected to RabbitMQ", map[string]interface{}{
		"url": b.config.URL,
	})

	return nil
}

// Disconnect closes the connection
func (b *Broker) Disconnect() error {
	b.connMu.Lock()
	defer b.connMu.Unlock()

	if !b.connected {
		return nil
	}

	b.cancel()
	b.connected = false

	// Close all subscribers
	b.subMu.Lock()
	for _, sub := range b.subscribers {
		if sub.cancel != nil {
			sub.cancel()
		}
	}
	b.subMu.Unlock()

	// Close channel pool
	close(b.channels)
	for ch := range b.channels {
		ch.Close()
	}

	// Close connection
	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			return err
		}
	}

	b.wg.Wait()

	b.logger.Sugar().Info("Disconnected from RabbitMQ", nil)
	return nil
}

// IsConnected returns connection status
func (b *Broker) IsConnected() bool {
	b.connMu.RLock()
	defer b.connMu.RUnlock()
	return b.connected && b.conn != nil && !b.conn.IsClosed()
}

// Health checks broker health
func (b *Broker) Health(ctx context.Context) error {
	if !b.IsConnected() {
		return errors.New("not connected to RabbitMQ")
	}

	ch, err := b.getChannel()
	if err != nil {
		return err
	}
	defer b.releaseChannel(ch)

	return nil
}

// Publish sends a message to a topic
func (b *Broker) Publish(ctx context.Context, topic string, msg *pubsub.Message) error {
	start := time.Now()
	defer func() {
		b.metrics.RecordPublishLatency(topic, time.Since(start))
	}()

	ch, err := b.getChannel()
	if err != nil {
		b.metrics.IncrementFailed(topic)
		return err
	}
	defer b.releaseChannel(ch)

	publishing := amqp.Publishing{
		ContentType:   msg.ContentType,
		Body:          msg.Payload,
		MessageId:     msg.ID,
		Timestamp:     msg.Timestamp,
		CorrelationId: msg.CorrelationID,
		Priority:      uint8(msg.Priority),
		Headers:       amqp.Table{},
	}

	if msg.TTL > 0 {
		publishing.Expiration = fmt.Sprintf("%d", msg.TTL.Milliseconds())
	}

	for k, v := range msg.Headers {
		publishing.Headers[k] = v
	}

	if msg.Metadata != nil {
		metadataJSON, _ := json.Marshal(msg.Metadata)
		publishing.Headers["x-metadata"] = string(metadataJSON)
	}

	err = ch.Publish(
		topic, // exchange
		topic, // routing key
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		b.metrics.IncrementFailed(topic)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	b.metrics.IncrementPublished(topic)
	b.logger.Sugar().Debug("Published message", map[string]interface{}{
		"topic":     topic,
		"messageId": msg.ID,
	})

	return nil
}

// PublishBatch sends multiple messages atomically
func (b *Broker) PublishBatch(ctx context.Context, messages []*pubsub.Message) error {
	// RabbitMQ doesn't support true atomic batches, but we can use transactions
	ch, err := b.getChannel()
	if err != nil {
		return err
	}
	defer b.releaseChannel(ch)

	if err := ch.Tx(); err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	for _, msg := range messages {
		if err := b.Publish(ctx, msg.Topic, msg); err != nil {
			ch.TxRollback()
			return err
		}
	}

	if err := ch.TxCommit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Request sends a message and waits for reply (RPC pattern)
func (b *Broker) Request(ctx context.Context, topic string, msg *pubsub.Message, timeout time.Duration) (*pubsub.Message, error) {
	ch, err := b.getChannel()
	if err != nil {
		return nil, err
	}
	defer b.releaseChannel(ch)

	// Create temporary reply queue
	replyQueue, err := ch.QueueDeclare(
		"",    // name (auto-generated)
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	msgs, err := ch.Consume(
		replyQueue.Name,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume reply queue: %w", err)
	}

	corrID := generateUUID()
	msg.CorrelationID = corrID

	publishing := amqp.Publishing{
		ContentType:   msg.ContentType,
		Body:          msg.Payload,
		MessageId:     msg.ID,
		CorrelationId: corrID,
		ReplyTo:       replyQueue.Name,
	}

	err = ch.Publish(topic, topic, false, false, publishing)
	if err != nil {
		return nil, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, errors.New("request timeout")
		case d := <-msgs:
			if d.CorrelationId == corrID {
				reply := &pubsub.Message{
					ID:            d.MessageId,
					Payload:       d.Body,
					CorrelationID: d.CorrelationId,
					Timestamp:     d.Timestamp,
					ContentType:   d.ContentType,
					Headers:       make(map[string]string),
				}
				for k, v := range d.Headers {
					if str, ok := v.(string); ok {
						reply.Headers[k] = str
					}
				}
				return reply, nil
			}
		}
	}
}

// Subscribe registers a handler for a topic
func (b *Broker) Subscribe(ctx context.Context, topic string, handler pubsub.MessageHandler, opts ...pubsub.SubscribeOption) error {
	options := &pubsub.SubscribeOptions{
		QueueName:         topic,
		AutoAck:           false,
		Exclusive:         false,
		PrefetchCount:     b.config.PrefetchCount,
		PrefetchSize:      b.config.PrefetchSize,
		ConcurrentWorkers: 1,
	}

	for _, opt := range opts {
		opt(options)
	}

	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	if err := ch.Qos(options.PrefetchCount, options.PrefetchSize, false); err != nil {
		ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare queue
	queue, err := ch.QueueDeclare(
		options.QueueName,
		true,  // durable
		false, // auto-delete
		options.Exclusive,
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange/topic
	if err := ch.QueueBind(queue.Name, topic, topic, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	deliveries, err := ch.Consume(
		queue.Name,
		options.ConsumerTag,
		options.AutoAck,
		options.Exclusive,
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to consume: %w", err)
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &subscription{
		topic:    topic,
		handler:  handler,
		options:  options,
		cancel:   cancel,
		delivery: deliveries,
	}

	b.subMu.Lock()
	b.subscribers[topic] = sub
	b.subMu.Unlock()

	// Apply middleware
	finalHandler := options.Middleware.Apply(handler)
	finalHandler = b.middleware.Apply(finalHandler)

	// Start workers
	for i := 0; i < options.ConcurrentWorkers; i++ {
		b.wg.Add(1)
		go b.processMessages(subCtx, ch, sub, finalHandler)
	}

	b.logger.Sugar().Info("Subscribed to topic", map[string]interface{}{
		"topic": topic,
		"queue": queue.Name,
	})

	return nil
}

// Unsubscribe removes a subscription
func (b *Broker) Unsubscribe(topic string) error {
	b.subMu.Lock()
	sub, exists := b.subscribers[topic]
	if exists {
		delete(b.subscribers, topic)
	}
	b.subMu.Unlock()

	if !exists {
		return fmt.Errorf("subscription not found for topic: %s", topic)
	}

	if sub.cancel != nil {
		sub.cancel()
	}

	b.logger.Sugar().Info("Unsubscribed from topic", map[string]interface{}{
		"topic": topic,
	})

	return nil
}

// Close closes the broker
func (b *Broker) Close() error {
	return b.Disconnect()
}

// Helper methods

func (b *Broker) getChannel() (*amqp.Channel, error) {
	select {
	case ch := <-b.channels:
		return ch, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("timeout getting channel from pool")
	}
}

func (b *Broker) releaseChannel(ch *amqp.Channel) {
	select {
	case b.channels <- ch:
	default:
		ch.Close()
	}
}

func (b *Broker) setupTopology() error {
	ch, err := b.getChannel()
	if err != nil {
		return err
	}
	defer b.releaseChannel(ch)

	// Declare exchanges
	for _, exchange := range b.topology.Exchanges {
		err := ch.ExchangeDeclare(
			exchange.Name,
			string(exchange.Type),
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			amqp.Table(exchange.Args),
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}
	}

	// Declare queues
	for _, queue := range b.topology.Queues {
		_, err := ch.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			amqp.Table(queue.Args),
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue.Name, err)
		}
	}

	// Create bindings
	for _, binding := range b.topology.Bindings {
		err := ch.QueueBind(
			binding.Queue,
			binding.RoutingKey,
			binding.Exchange,
			binding.NoWait,
			amqp.Table(binding.Args),
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w",
				binding.Queue, binding.Exchange, err)
		}
	}

	return nil
}

func (b *Broker) processMessages(ctx context.Context, ch *amqp.Channel, sub *subscription, handler pubsub.MessageHandler) {
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-sub.delivery:
			if !ok {
				return
			}

			start := time.Now()
			msg := b.convertDelivery(&d)

			err := handler(ctx, msg)
			if err != nil {
				b.metrics.IncrementFailed(sub.topic)
				b.logger.Sugar().Error("Failed to process message", map[string]interface{}{
					"topic":     sub.topic,
					"messageId": msg.ID,
					"error":     err.Error(),
				})

				// Handle retry logic
				if sub.options.RetryPolicy != nil {
					b.handleRetry(ch, &d, sub, err)
				} else {
					d.Nack(false, false)
				}
			} else {
				b.metrics.IncrementConsumed(sub.topic)
				b.metrics.RecordProcessingLatency(sub.topic, time.Since(start))
				d.Ack(false)
			}
		}
	}
}

func (b *Broker) handleRetry(ch *amqp.Channel, d *amqp.Delivery, sub *subscription, err error) {
	retryCount := 0
	if val, ok := d.Headers["x-retry-count"].(int32); ok {
		retryCount = int(val)
	}

	maxRetries := 3
	if sub.options.RetryPolicy != nil {
		maxRetries = sub.options.RetryPolicy.MaxAttempts
	}

	if retryCount >= maxRetries {
		// Send to DLQ if configured
		if sub.options.DeadLetterQueue != nil && sub.options.DeadLetterQueue.Enabled {
			b.sendToDeadLetter(ch, d, sub)
		}
		d.Nack(false, false)
		return
	}

	// Requeue with delay
	retryCount++
	d.Headers["x-retry-count"] = int32(retryCount)

	delay := b.calculateRetryDelay(retryCount, sub.options.RetryPolicy)
	time.AfterFunc(delay, func() {
		ch.Publish(
			d.Exchange,
			d.RoutingKey,
			false,
			false,
			amqp.Publishing{
				Headers:       d.Headers,
				ContentType:   d.ContentType,
				Body:          d.Body,
				MessageId:     d.MessageId,
				CorrelationId: d.CorrelationId,
			},
		)
	})

	d.Ack(false)
	b.metrics.IncrementRetry(sub.topic)
}

func (b *Broker) sendToDeadLetter(ch *amqp.Channel, d *amqp.Delivery, sub *subscription) {
	dlqName := sub.options.DeadLetterQueue.QueueName
	if dlqName == "" {
		dlqName = sub.topic + ".dlq"
	}

	ch.Publish(
		"",
		dlqName,
		false,
		false,
		amqp.Publishing{
			Headers:     d.Headers,
			ContentType: d.ContentType,
			Body:        d.Body,
			MessageId:   d.MessageId,
		},
	)
}

func (b *Broker) calculateRetryDelay(attempt int, policy *pubsub.RetryPolicy) time.Duration {
	if policy == nil {
		return 5 * time.Second
	}

	delay := policy.InitialInterval
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * policy.Multiplier)
		if delay > policy.MaxInterval {
			delay = policy.MaxInterval
			break
		}
	}

	return delay
}

func (b *Broker) convertDelivery(d *amqp.Delivery) *pubsub.Message {
	msg := &pubsub.Message{
		ID:            d.MessageId,
		Payload:       d.Body,
		CorrelationID: d.CorrelationId,
		Timestamp:     d.Timestamp,
		ContentType:   d.ContentType,
		Headers:       make(map[string]string),
	}

	for k, v := range d.Headers {
		if str, ok := v.(string); ok {
			msg.Headers[k] = str
		}
	}

	return msg
}

func (b *Broker) monitorConnection() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		case err := <-b.closeChan:
			if err != nil {
				b.logger.Sugar().Error("Connection closed", map[string]interface{}{
					"error": err.Error(),
				})
				b.reconnect()
			}
		}
	}
}

func (b *Broker) reconnect() {
	b.connMu.Lock()
	b.connected = false
	b.connMu.Unlock()

	for attempt := 1; attempt <= b.config.ReconnectAttempts; attempt++ {
		b.logger.Sugar().Info("Attempting to reconnect", map[string]interface{}{
			"attempt": attempt,
		})

		time.Sleep(b.config.ReconnectInterval)

		if err := b.Connect(b.ctx); err != nil {
			b.logger.Sugar().Error("Reconnection failed", map[string]interface{}{
				"attempt": attempt,
				"error":   err.Error(),
			})
			continue
		}

		// Resubscribe to all topics
		b.subMu.RLock()
		for topic, sub := range b.subscribers {
			b.Subscribe(b.ctx, topic, sub.handler, func(o *pubsub.SubscribeOptions) {
				*o = *sub.options
			})
		}
		b.subMu.RUnlock()

		b.logger.Sugar().Info("Reconnected successfully", nil)
		return
	}

	b.logger.Sugar().Error("Failed to reconnect after max attempts", nil)
}

func generateUUID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
