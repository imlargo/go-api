package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPubSub implements PubSub interface using RabbitMQ
type RabbitMQPubSub struct {
	config       *Config
	conn         *amqp.Connection
	channel      *amqp.Channel
	subscribers  map[string]*subscription
	mu           sync.RWMutex
	done         chan struct{}
	reconnecting bool
	reconnectMu  sync.Mutex
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

type subscription struct {
	topic   string
	handler HandlerFunc
	opts    subscribeOptions
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewRabbitMQPubSub creates a new RabbitMQ pub/sub instance
func NewRabbitMQPubSub(config *Config) (PubSub, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ps := &RabbitMQPubSub{
		config:      config,
		subscribers: make(map[string]*subscription),
		done:        make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	if err := ps.Connect(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Start connection monitor
	go ps.monitorConnection()

	return ps, nil
}

// Connect establishes connection to RabbitMQ
func (ps *RabbitMQPubSub) Connect(ctx context.Context) error {
	ps.reconnectMu.Lock()
	defer ps.reconnectMu.Unlock()

	connectCtx, cancel := context.WithTimeout(ctx, ps.config.ConnectionTimeout)
	defer cancel()

	connChan := make(chan *amqp.Connection, 1)
	errChan := make(chan error, 1)

	go func() {
		conn, err := amqp.Dial(ps.config.URL)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	select {
	case conn := <-connChan:
		ps.conn = conn
	case err := <-errChan:
		return fmt.Errorf("failed to dial: %w", err)
	case <-connectCtx.Done():
		return fmt.Errorf("connection timeout")
	}

	channel, err := ps.conn.Channel()
	if err != nil {
		ps.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	ps.channel = channel

	// Declare exchange
	err = ps.channel.ExchangeDeclare(
		ps.config.ExchangeName,
		ps.config.ExchangeType,
		ps.config.Durable,
		ps.config.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		ps.channel.Close()
		ps.conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	return nil
}

// Disconnect closes the connection
func (ps *RabbitMQPubSub) Disconnect() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.channel != nil {
		ps.channel.Close()
	}
	if ps.conn != nil {
		return ps.conn.Close()
	}
	return nil
}

// IsConnected checks if connection is active
func (ps *RabbitMQPubSub) IsConnected() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.conn != nil && !ps.conn.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (ps *RabbitMQPubSub) Reconnect(ctx context.Context) error {
	ps.reconnectMu.Lock()
	if ps.reconnecting {
		ps.reconnectMu.Unlock()
		return nil
	}
	ps.reconnecting = true
	ps.reconnectMu.Unlock()

	defer func() {
		ps.reconnectMu.Lock()
		ps.reconnecting = false
		ps.reconnectMu.Unlock()
	}()

	ps.config.Logger.Info("Attempting to reconnect...")

	for attempt := 1; attempt <= ps.config.MaxReconnect; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ps.config.Logger.Info(fmt.Sprintf("Reconnection attempt %d/%d", attempt, ps.config.MaxReconnect))

		if err := ps.Connect(ctx); err != nil {
			ps.config.Logger.Error(fmt.Sprintf("Reconnection attempt %d failed: %v", attempt, err))
			if attempt < ps.config.MaxReconnect {
				time.Sleep(ps.config.ReconnectDelay)
			}
			continue
		}

		ps.config.Logger.Info("Reconnection successful")

		// Resubscribe to all topics
		ps.mu.RLock()
		subs := make([]*subscription, 0, len(ps.subscribers))
		for _, sub := range ps.subscribers {
			subs = append(subs, sub)
		}
		ps.mu.RUnlock()

		for _, sub := range subs {
			opts := make([]SubscribeOption, 0)
			opts = append(opts, WithQueueName(sub.opts.queueName))
			opts = append(opts, WithRoutingKey(sub.opts.routingKey))
			opts = append(opts, WithAutoAck(sub.opts.autoAck))
			opts = append(opts, WithConcurrency(sub.opts.concurrency))

			if err := ps.Subscribe(sub.topic, sub.handler, opts...); err != nil {
				ps.config.Logger.Error(fmt.Sprintf("Failed to resubscribe to %s: %v", sub.topic, err))
			}
		}

		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts", ps.config.MaxReconnect)
}

// monitorConnection monitors the connection and attempts reconnection
func (ps *RabbitMQPubSub) monitorConnection() {
	for {
		select {
		case <-ps.done:
			return
		case <-time.After(5 * time.Second):
			if !ps.IsConnected() {
				ps.config.Logger.Warn("Connection lost, attempting reconnection...")
				if err := ps.Reconnect(ps.ctx); err != nil {
					ps.config.Logger.Error(fmt.Sprintf("Reconnection failed: %v", err))
				}
			}
		}
	}
}

// Publish publishes a message to a topic
func (ps *RabbitMQPubSub) Publish(ctx context.Context, topic string, payload interface{}, opts ...PublishOption) error {
	if !ps.IsConnected() {
		return errors.New("not connected to RabbitMQ")
	}

	options := &publishOptions{
		headers:    make(map[string]string),
		persistent: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	var data []byte
	var err error

	switch v := payload.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	msg := &Message{
		ID:          generateID(),
		Topic:       topic,
		Payload:     data,
		Headers:     options.headers,
		PublishedAt: time.Now(),
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range options.headers {
		headers[k] = v
	}

	deliveryMode := uint8(1) // Non-persistent
	if options.persistent {
		deliveryMode = 2 // Persistent
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         msgData,
		Headers:      headers,
		Priority:     options.priority,
		Expiration:   options.expiration,
		DeliveryMode: deliveryMode,
		Timestamp:    time.Now(),
		MessageId:    msg.ID,
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	err = ps.channel.PublishWithContext(
		ctx,
		ps.config.ExchangeName,
		topic,
		false,
		false,
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	ps.config.Logger.Debug(fmt.Sprintf("Published message to topic %s: %s", topic, msg.ID))
	return nil
}

// Subscribe subscribes to a topic with a handler
func (ps *RabbitMQPubSub) Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error {
	if !ps.IsConnected() {
		return errors.New("not connected to RabbitMQ")
	}

	options := subscribeOptions{
		queueName:   "",
		routingKey:  topic,
		autoAck:     false,
		exclusive:   false,
		concurrency: 1,
		retryStrategy: &ExponentialBackoff{
			MaxRetries:   ps.config.DefaultRetries,
			InitialDelay: ps.config.RetryDelay,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}

	for _, opt := range opts {
		opt(&options)
	}

	if options.queueName == "" {
		options.queueName = fmt.Sprintf("queue.%s", topic)
	}

	ps.mu.Lock()

	// Cancel existing subscription if any
	if existing, ok := ps.subscribers[topic]; ok {
		existing.cancel()
		<-existing.done
		delete(ps.subscribers, topic)
	}

	queue, err := ps.channel.QueueDeclare(
		options.queueName,
		ps.config.QueueDurable,
		ps.config.QueueAutoDelete,
		ps.config.QueueExclusive,
		false,
		nil,
	)
	if err != nil {
		ps.mu.Unlock()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ps.channel.QueueBind(
		queue.Name,
		options.routingKey,
		ps.config.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		ps.mu.Unlock()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	err = ps.channel.Qos(
		ps.config.PrefetchCount,
		ps.config.PrefetchSize,
		false,
	)
	if err != nil {
		ps.mu.Unlock()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ps.channel.Consume(
		queue.Name,
		"",
		options.autoAck,
		options.exclusive,
		false,
		false,
		nil,
	)
	if err != nil {
		ps.mu.Unlock()
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(ps.ctx)
	sub := &subscription{
		topic:   topic,
		handler: handler,
		opts:    options,
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	ps.subscribers[topic] = sub
	ps.mu.Unlock()

	ps.config.Logger.Info(fmt.Sprintf("Subscribed to topic %s with queue %s", topic, queue.Name))

	// Start consumer workers
	for i := 0; i < options.concurrency; i++ {
		ps.wg.Add(1)
		go ps.consumeMessages(ctx, sub, msgs, options)
	}

	return nil
}

// consumeMessages processes messages from the queue
func (ps *RabbitMQPubSub) consumeMessages(ctx context.Context, sub *subscription, msgs <-chan amqp.Delivery, opts subscribeOptions) {
	defer ps.wg.Done()
	defer close(sub.done)

	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-msgs:
			if !ok {
				ps.config.Logger.Warn(fmt.Sprintf("Message channel closed for topic %s", sub.topic))
				return
			}

			ps.processMessage(ctx, sub, delivery, opts)
		}
	}
}

// processMessage handles individual message processing
func (ps *RabbitMQPubSub) processMessage(ctx context.Context, sub *subscription, delivery amqp.Delivery, opts subscribeOptions) {
	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		ps.config.Logger.Error(fmt.Sprintf("Failed to unmarshal message: %v", err))
		delivery.Nack(false, false)
		return
	}

	ps.config.Logger.Debug(fmt.Sprintf("Processing message %s from topic %s", msg.ID, msg.Topic))

	handlerErr := sub.handler(ctx, &msg)

	if handlerErr != nil {
		ps.config.Logger.Error(fmt.Sprintf("Handler error for message %s: %v", msg.ID, handlerErr))

		if opts.retryStrategy != nil && opts.retryStrategy.ShouldRetry(&msg, handlerErr) {
			msg.RetryCount++
			delay := opts.retryStrategy.NextDelay(&msg)
			ps.config.Logger.Info(fmt.Sprintf("Retrying message %s after %v (attempt %d)", msg.ID, delay, msg.RetryCount))

			time.Sleep(delay)

			// Republish for retry
			if err := ps.Publish(ctx, msg.Topic, msg.Payload, WithHeaders(msg.Headers)); err != nil {
				ps.config.Logger.Error(fmt.Sprintf("Failed to republish message %s: %v", msg.ID, err))
			}
		}

		delivery.Nack(false, false)
		return
	}

	if !opts.autoAck {
		delivery.Ack(false)
	}

	ps.config.Logger.Debug(fmt.Sprintf("Successfully processed message %s", msg.ID))
}

// Unsubscribe removes a subscription
func (ps *RabbitMQPubSub) Unsubscribe(topic string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	sub, ok := ps.subscribers[topic]
	if !ok {
		return fmt.Errorf("no subscription found for topic: %s", topic)
	}

	sub.cancel()
	<-sub.done
	delete(ps.subscribers, topic)

	ps.config.Logger.Info(fmt.Sprintf("Unsubscribed from topic %s", topic))
	return nil
}

// Close gracefully shuts down the pub/sub system
func (ps *RabbitMQPubSub) Close() error {
	ps.config.Logger.Info("Shutting down pub/sub system...")

	close(ps.done)
	ps.cancel()

	// Wait for all consumers to finish with timeout
	done := make(chan struct{})
	go func() {
		ps.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		ps.config.Logger.Info("All consumers shut down gracefully")
	case <-time.After(ps.config.ShutdownTimeout):
		ps.config.Logger.Warn("Shutdown timeout exceeded, forcing close")
	}

	return ps.Disconnect()
}
