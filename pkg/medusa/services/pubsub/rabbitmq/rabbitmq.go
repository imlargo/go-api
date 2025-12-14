// Package rabbitmq provides RabbitMQ implementation of the pubsub interfaces
// File: pubsub/rabbitmq/rabbitmq.go

package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ============================================================================
// RABBITMQ CLIENT
// ============================================================================

// Client implements the PubSub interface for RabbitMQ
type Client struct {
	config      *pubsub.Config
	conn        *amqp.Connection
	publishChan *amqp.Channel
	mu          sync.RWMutex

	subscriptions map[string]*subscription
	subMu         sync.RWMutex

	reconnecting  bool
	reconnectChan chan struct{}
	closed        bool
	closeChan     chan struct{}
	wg            sync.WaitGroup

	logger  *logger.Logger
	metrics pubsub.MetricsCollector
}

type subscription struct {
	topic      string
	handler    pubsub.MessageHandler
	options    *pubsub.SubscribeOptions
	channel    *amqp.Channel
	deliveries <-chan amqp.Delivery
	cancel     context.CancelFunc
	closeChan  chan struct{}
}

// New creates a new RabbitMQ client
func New(config *pubsub.Config) (*Client, error) {
	if config == nil {
		config = pubsub.DefaultConfig()
	}

	if config.URL == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}

	if config.Logger == nil {
		return nil, fmt.Errorf("Logger is required in config")
	}

	client := &Client{
		config:        config,
		subscriptions: make(map[string]*subscription),
		reconnectChan: make(chan struct{}, 1),
		closeChan:     make(chan struct{}),
		logger:        config.Logger,
		metrics:       &pubsub.NoOpMetrics{},
	}

	return client, nil
}

// SetMetrics sets the metrics collector
func (c *Client) SetMetrics(metrics pubsub.MetricsCollector) {
	c.metrics = metrics
}

// Connect establishes connection to RabbitMQ
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		return pubsub.ErrAlreadyConnected
	}

	conn, err := amqp.Dial(c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.conn = conn
	c.logger.Info(fmt.Sprintf("Connected to RabbitMQ at %s", c.config.URL))

	// Create publisher channel
	publishChan, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create publish channel: %w", err)
	}

	if c.config.PublisherConfirm {
		if err := publishChan.Confirm(false); err != nil {
			conn.Close()
			return fmt.Errorf("failed to enable publisher confirms: %w", err)
		}
	}

	c.publishChan = publishChan

	// Start connection monitor
	c.wg.Add(1)
	go c.monitorConnection()

	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}

	c.logger.Info("Disconnecting from RabbitMQ")

	if c.publishChan != nil {
		c.publishChan.Close()
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.conn = nil
	c.publishChan = nil

	return nil
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.conn.IsClosed()
}

// Publish publishes a message to a topic
func (c *Client) Publish(ctx context.Context, topic string, message *pubsub.Message, opts ...pubsub.PublishOption) error {
	if c.closed {
		return pubsub.ErrClosed
	}

	if !c.IsConnected() {
		return pubsub.ErrNotConnected
	}

	// Apply options
	options := &pubsub.PublishOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Prepare message
	if message.ID == "" {
		message.ID = generateID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	message.Topic = topic

	// Build AMQP message
	publishing := amqp.Publishing{
		MessageId:     message.ID,
		Timestamp:     message.Timestamp,
		ContentType:   message.ContentType,
		Body:          message.Payload,
		DeliveryMode:  amqp.Transient,
		CorrelationId: options.CorrelationID,
		ReplyTo:       options.ReplyTo,
		Priority:      options.Priority,
	}

	if options.Persistent {
		publishing.DeliveryMode = amqp.Persistent
	}

	if options.Expiration > 0 {
		publishing.Expiration = fmt.Sprintf("%d", options.Expiration.Milliseconds())
	}

	// Add headers
	if len(message.Headers) > 0 || len(options.Headers) > 0 {
		publishing.Headers = make(amqp.Table)
		for k, v := range message.Headers {
			publishing.Headers[k] = v
		}
		for k, v := range options.Headers {
			publishing.Headers[k] = v
		}
	}

	// Publish with timeout
	publishCtx, cancel := context.WithTimeout(ctx, c.config.PublishTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		c.mu.RLock()
		defer c.mu.RUnlock()

		if c.publishChan == nil {
			errChan <- pubsub.ErrNotConnected
			return
		}

		err := c.publishChan.Publish(
			topic, // exchange
			"",    // routing key (use default)
			options.Mandatory,
			options.Immediate,
			publishing,
		)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to publish to topic %s: %v", topic, err))
			return fmt.Errorf("%w: %v", pubsub.ErrPublishFailed, err)
		}
		c.metrics.IncMessagesPublished(topic)
		c.logger.Debug(fmt.Sprintf("Published message to topic %s with id %s", topic, message.ID))
		return nil
	case <-publishCtx.Done():
		return pubsub.ErrTimeout
	}
}

// Subscribe subscribes to a topic
func (c *Client) Subscribe(ctx context.Context, topic string, handler pubsub.MessageHandler, opts ...pubsub.SubscribeOption) error {
	if c.closed {
		return pubsub.ErrClosed
	}

	if !c.IsConnected() {
		return pubsub.ErrNotConnected
	}

	// Check if already subscribed
	c.subMu.RLock()
	if _, exists := c.subscriptions[topic]; exists {
		c.subMu.RUnlock()
		return fmt.Errorf("already subscribed to topic: %s", topic)
	}
	c.subMu.RUnlock()

	// Apply options
	options := &pubsub.SubscribeOptions{
		QueueName:     "",
		Durable:       true,
		AutoDelete:    false,
		Exclusive:     false,
		ConsumerTag:   "",
		AutoAck:       c.config.AutoAck,
		PrefetchCount: c.config.PrefetchCount,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Apply middleware
	finalHandler := handler
	if len(options.Middleware) > 0 {
		finalHandler = pubsub.ChainMiddleware(options.Middleware...)(handler)
	}

	// Create channel for this subscription
	c.mu.RLock()
	ch, err := c.conn.Channel()
	c.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Set QoS
	qos := c.config.QoS
	if options.QoS != nil {
		qos = *options.QoS
	} else {
		qos.PrefetchCount = options.PrefetchCount
	}

	if err := ch.Qos(qos.PrefetchCount, qos.PrefetchSize, qos.Global); err != nil {
		ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare exchange
	if err := ch.ExchangeDeclare(
		topic,
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Generate queue name if not provided
	queueName := options.QueueName
	if queueName == "" {
		queueName = fmt.Sprintf("%s.%s", topic, generateID())
	}

	// Prepare queue arguments
	queueArgs := amqp.Table{}

	// Configure dead letter queue if enabled
	if options.DeadLetter != nil && options.DeadLetter.Enabled {
		queueArgs["x-dead-letter-exchange"] = options.DeadLetter.ExchangeName
		if options.DeadLetter.RoutingKey != "" {
			queueArgs["x-dead-letter-routing-key"] = options.DeadLetter.RoutingKey
		}
		if options.DeadLetter.TTL > 0 {
			queueArgs["x-message-ttl"] = int32(options.DeadLetter.TTL.Milliseconds())
		}
	}

	// Declare queue
	queue, err := ch.QueueDeclare(
		queueName,
		options.Durable,
		options.AutoDelete,
		options.Exclusive,
		options.NoWait,
		queueArgs,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	if err := ch.QueueBind(
		queue.Name,
		"",    // routing key
		topic, // exchange
		false,
		nil,
	); err != nil {
		ch.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming
	consumerTag := options.ConsumerTag
	if consumerTag == "" {
		consumerTag = fmt.Sprintf("consumer-%s", generateID())
	}

	deliveries, err := ch.Consume(
		queue.Name,
		consumerTag,
		options.AutoAck,
		options.Exclusive,
		false, // no-local
		options.NoWait,
		nil,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// Create subscription
	subCtx, cancel := context.WithCancel(ctx)
	sub := &subscription{
		topic:      topic,
		handler:    finalHandler,
		options:    options,
		channel:    ch,
		deliveries: deliveries,
		cancel:     cancel,
		closeChan:  make(chan struct{}),
	}

	// Store subscription
	c.subMu.Lock()
	c.subscriptions[topic] = sub
	c.subMu.Unlock()

	// Start processing messages
	c.wg.Add(1)
	go c.processMessages(subCtx, sub)

	c.logger.Info(fmt.Sprintf("Subscribed to topic %s with queue %s", topic, queue.Name))
	return nil
}

// processMessages processes incoming messages
func (c *Client) processMessages(ctx context.Context, sub *subscription) {
	defer c.wg.Done()
	defer close(sub.closeChan)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(fmt.Sprintf("Stopping message processing for topic %s", sub.topic))
			return
		case <-c.closeChan:
			return
		case delivery, ok := <-sub.deliveries:
			if !ok {
				c.logger.Warn(fmt.Sprintf("Delivery channel closed for topic %s", sub.topic))
				return
			}

			c.handleDelivery(ctx, sub, delivery)
		}
	}
}

// handleDelivery handles a single message delivery
func (c *Client) handleDelivery(ctx context.Context, sub *subscription, delivery amqp.Delivery) {
	c.metrics.IncMessagesReceived(sub.topic)

	// Convert AMQP delivery to pubsub.Message
	msg := &pubsub.Message{
		ID:          delivery.MessageId,
		Topic:       sub.topic,
		Payload:     delivery.Body,
		Timestamp:   delivery.Timestamp,
		ContentType: delivery.ContentType,
		Headers:     make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}

	// Convert headers
	for k, v := range delivery.Headers {
		if str, ok := v.(string); ok {
			msg.Headers[k] = str
		}
	}

	// Add metadata
	msg.Metadata["correlation_id"] = delivery.CorrelationId
	msg.Metadata["reply_to"] = delivery.ReplyTo
	msg.Metadata["priority"] = delivery.Priority
	msg.Metadata["delivery_tag"] = delivery.DeliveryTag

	// Set ack/nack functions
	msg.SetAckFunc(func() error {
		return delivery.Ack(false)
	})
	msg.SetNackFunc(func(requeue bool) error {
		return delivery.Nack(false, requeue)
	})

	// Process message
	start := time.Now()
	err := sub.handler(ctx, msg)
	duration := time.Since(start)

	c.metrics.ObserveProcessingDuration(sub.topic, duration)

	if err != nil {
		c.metrics.IncMessagesFailed(sub.topic)
		c.logger.Error(fmt.Sprintf("Handler error for topic %s, message_id %s: %v", sub.topic, msg.ID, err))

		// Handle error
		if sub.options.ErrorHandler != nil {
			if handlerErr := sub.options.ErrorHandler(ctx, msg, err); handlerErr != nil {
				c.logger.Error(fmt.Sprintf("Error handler failed: %v", handlerErr))
			}
		}

		// Nack with requeue if not auto-ack
		if !sub.options.AutoAck {
			if nackErr := msg.Nack(true); nackErr != nil {
				c.logger.Error(fmt.Sprintf("Failed to nack message: %v", nackErr))
			}
		}
		return
	}

	c.metrics.IncMessagesProcessed(sub.topic)

	// Ack if not auto-ack
	if !sub.options.AutoAck {
		if ackErr := msg.Ack(); ackErr != nil {
			c.logger.Error(fmt.Sprintf("Failed to ack message: %v", ackErr))
		}
	}
}

// Unsubscribe removes a subscription
func (c *Client) Unsubscribe(topic string) error {
	c.subMu.Lock()
	sub, exists := c.subscriptions[topic]
	if !exists {
		c.subMu.Unlock()
		return fmt.Errorf("not subscribed to topic: %s", topic)
	}
	delete(c.subscriptions, topic)
	c.subMu.Unlock()

	// Cancel context
	sub.cancel()

	// Wait for processing to stop
	<-sub.closeChan

	// Close channel
	if err := sub.channel.Close(); err != nil {
		c.logger.Error(fmt.Sprintf("Failed to close subscription channel: %v", err))
	}

	c.logger.Info(fmt.Sprintf("Unsubscribed from topic %s", topic))
	return nil
}

// Close gracefully shuts down the client
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	c.logger.Info("Closing RabbitMQ client")

	// Signal close
	close(c.closeChan)

	// Unsubscribe from all topics
	c.subMu.RLock()
	topics := make([]string, 0, len(c.subscriptions))
	for topic := range c.subscriptions {
		topics = append(topics, topic)
	}
	c.subMu.RUnlock()

	for _, topic := range topics {
		if err := c.Unsubscribe(topic); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to unsubscribe from topic %s: %v", topic, err))
		}
	}

	// Wait for goroutines
	c.wg.Wait()

	// Disconnect
	return c.Disconnect()
}

// HealthCheck checks the connection health
func (c *Client) HealthCheck(ctx context.Context) error {
	if !c.IsConnected() {
		return pubsub.ErrNotConnected
	}

	// Try to create a temporary channel
	c.mu.RLock()
	ch, err := c.conn.Channel()
	c.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	ch.Close()

	return nil
}

// monitorConnection monitors the connection and handles reconnection
func (c *Client) monitorConnection() {
	defer c.wg.Done()

	for {
		select {
		case <-c.closeChan:
			return
		default:
			if !c.IsConnected() && !c.closed {
				c.logger.Warn("Connection lost, attempting to reconnect...")
				c.handleReconnect()
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// handleReconnect attempts to reconnect
func (c *Client) handleReconnect() {
	c.mu.Lock()
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	for attempt := 0; attempt < c.config.MaxReconnects; attempt++ {
		if c.closed {
			return
		}

		c.logger.Info(fmt.Sprintf("Reconnection attempt %d", attempt+1))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := c.Connect(ctx)
		cancel()

		if err == nil {
			c.logger.Info("Reconnected successfully")
			c.resubscribeAll()
			return
		}

		c.logger.Error(fmt.Sprintf("Reconnection failed: %v", err))
		time.Sleep(c.config.ReconnectDelay)
	}

	c.logger.Error("Max reconnection attempts reached")
}

// resubscribeAll re-subscribes to all topics after reconnection
func (c *Client) resubscribeAll() {
	c.subMu.RLock()
	subscriptions := make([]*subscription, 0, len(c.subscriptions))
	for _, sub := range c.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	c.subMu.RUnlock()

	for _, sub := range subscriptions {
		opts := buildSubscribeOptions(sub.options)
		if err := c.Subscribe(context.Background(), sub.topic, sub.handler, opts...); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to resubscribe to topic %s: %v", sub.topic, err))
		}
	}
}

// Helper functions

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func buildSubscribeOptions(opts *pubsub.SubscribeOptions) []pubsub.SubscribeOption {
	var options []pubsub.SubscribeOption

	if opts.QueueName != "" {
		options = append(options, pubsub.WithQueueName(opts.QueueName))
	}
	if opts.Durable {
		options = append(options, pubsub.WithDurable())
	}
	if opts.AutoDelete {
		options = append(options, pubsub.WithAutoDelete())
	}
	if opts.Exclusive {
		options = append(options, pubsub.WithExclusive())
	}
	if opts.ConsumerTag != "" {
		options = append(options, pubsub.WithConsumerTag(opts.ConsumerTag))
	}
	if opts.QoS != nil {
		options = append(options, pubsub.WithQoS(*opts.QoS))
	}
	if opts.RetryPolicy != nil {
		options = append(options, pubsub.WithRetryPolicy(*opts.RetryPolicy))
	}
	if opts.DeadLetter != nil {
		options = append(options, pubsub.WithDeadLetter(*opts.DeadLetter))
	}
	if len(opts.Middleware) > 0 {
		options = append(options, pubsub.WithMiddleware(opts.Middleware...))
	}
	if opts.ErrorHandler != nil {
		options = append(options, pubsub.WithErrorHandler(opts.ErrorHandler))
	}

	return options
}
