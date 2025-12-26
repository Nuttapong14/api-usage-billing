package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrConsumerClosed    = errors.New("consumer is closed")
	ErrProcessingFailed  = errors.New("event processing failed")
	ErrAckFailed         = errors.New("event acknowledgment failed")
)

// EventHandler handles an event
type EventHandler func(ctx context.Context, event *Event) error

// ConsumerConfig configures the event consumer
type ConsumerConfig struct {
	StreamPrefix     string
	ConsumerGroup    string
	ConsumerName     string
	Streams          []string
	BatchSize        int64
	BlockTimeout     time.Duration
	RetryDelay       time.Duration
	MaxRetries       int
	ClaimMinIdle     time.Duration
	ClaimBatchSize   int64
	CleanupInterval  time.Duration
}

// DefaultConsumerConfig returns default consumer configuration
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		StreamPrefix:     "events",
		ConsumerGroup:    "default",
		ConsumerName:     "consumer-1",
		Streams:          []string{"main"},
		BatchSize:        10,
		BlockTimeout:     5 * time.Second,
		RetryDelay:       1 * time.Second,
		MaxRetries:       3,
		ClaimMinIdle:     30 * time.Second,
		ClaimBatchSize:   10,
		CleanupInterval:  1 * time.Minute,
	}
}

// Consumer consumes events from Redis Streams
type Consumer struct {
	rdb           redis.UniversalClient
	config        ConsumerConfig
	handlers      map[EventType][]EventHandler
	defaultHandler EventHandler
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	running       bool
}

// NewConsumer creates a new event consumer
func NewConsumer(rdb redis.UniversalClient, config ...ConsumerConfig) *Consumer {
	cfg := DefaultConsumerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		rdb:      rdb,
		config:   cfg,
		handlers: make(map[EventType][]EventHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// streamName returns the full stream name
func (c *Consumer) streamName(stream string) string {
	return fmt.Sprintf("%s:%s", c.config.StreamPrefix, stream)
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(eventType EventType, handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[eventType] = append(c.handlers[eventType], handler)
}

// RegisterDefaultHandler registers a default handler for unhandled events
func (c *Consumer) RegisterDefaultHandler(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultHandler = handler
}

// Start starts consuming events
func (c *Consumer) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// Ensure consumer groups exist
	for _, stream := range c.config.Streams {
		if err := c.ensureConsumerGroup(stream); err != nil {
			return fmt.Errorf("failed to create consumer group for stream %s: %w", stream, err)
		}
	}

	// Start stream consumers
	for _, stream := range c.config.Streams {
		c.wg.Add(1)
		go c.consumeStream(stream)
	}

	// Start pending message claimer
	c.wg.Add(1)
	go c.claimPendingMessages()

	return nil
}

// Stop stops consuming events
func (c *Consumer) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	c.cancel()
	c.wg.Wait()
	return nil
}

// ensureConsumerGroup ensures the consumer group exists
func (c *Consumer) ensureConsumerGroup(stream string) error {
	streamName := c.streamName(stream)

	_, err := c.rdb.XGroupCreateMkStream(c.ctx, streamName, c.config.ConsumerGroup, "0").Result()
	if err != nil {
		// Ignore BUSYGROUP error (group already exists)
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return err
		}
	}

	return nil
}

// consumeStream consumes messages from a single stream
func (c *Consumer) consumeStream(stream string) {
	defer c.wg.Done()

	streamName := c.streamName(stream)
	lastID := ">" // Start with new messages

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Read messages
		streams, err := c.rdb.XReadGroup(c.ctx, &redis.XReadGroupArgs{
			Group:    c.config.ConsumerGroup,
			Consumer: c.config.ConsumerName,
			Streams:  []string{streamName, lastID},
			Count:    c.config.BatchSize,
			Block:    c.config.BlockTimeout,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
				continue
			}
			// Log error and continue
			time.Sleep(c.config.RetryDelay)
			continue
		}

		// Process messages
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				c.processMessage(streamName, msg)
			}
		}
	}
}

// processMessage processes a single message
func (c *Consumer) processMessage(stream string, msg redis.XMessage) {
	// Parse event data
	data, ok := msg.Values["data"].(string)
	if !ok {
		// Invalid message format, acknowledge and skip
		c.acknowledge(stream, msg.ID)
		return
	}

	var event Event
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		// Invalid JSON, acknowledge and skip
		c.acknowledge(stream, msg.ID)
		return
	}

	// Process event with handlers
	if err := c.handleEvent(&event); err != nil {
		// Processing failed, will be retried via pending messages
		return
	}

	// Acknowledge successful processing
	c.acknowledge(stream, msg.ID)
}

// handleEvent dispatches an event to registered handlers
func (c *Consumer) handleEvent(event *Event) error {
	c.mu.RLock()
	handlers := c.handlers[event.Type]
	defaultHandler := c.defaultHandler
	c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	// Call type-specific handlers
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	// Call default handler if no specific handlers exist
	if len(handlers) == 0 && defaultHandler != nil {
		if err := defaultHandler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// acknowledge acknowledges a message
func (c *Consumer) acknowledge(stream, msgID string) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	c.rdb.XAck(ctx, stream, c.config.ConsumerGroup, msgID)
}

// claimPendingMessages claims and reprocesses pending messages
func (c *Consumer) claimPendingMessages() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			for _, stream := range c.config.Streams {
				c.claimStreamPending(stream)
			}
		}
	}
}

// claimStreamPending claims pending messages from a stream
func (c *Consumer) claimStreamPending(stream string) {
	streamName := c.streamName(stream)

	// Get pending messages
	pending, err := c.rdb.XPendingExt(c.ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  c.config.ConsumerGroup,
		Start:  "-",
		End:    "+",
		Count:  c.config.ClaimBatchSize,
	}).Result()

	if err != nil {
		return
	}

	// Filter messages that are idle enough to claim
	messageIDs := make([]string, 0)
	for _, p := range pending {
		if p.Idle >= c.config.ClaimMinIdle {
			// Check retry count
			if p.RetryCount < int64(c.config.MaxRetries) {
				messageIDs = append(messageIDs, p.ID)
			} else {
				// Max retries exceeded, acknowledge and move to dead letter
				c.moveToDeadLetter(streamName, p.ID)
			}
		}
	}

	if len(messageIDs) == 0 {
		return
	}

	// Claim messages
	claimed, err := c.rdb.XClaim(c.ctx, &redis.XClaimArgs{
		Stream:   streamName,
		Group:    c.config.ConsumerGroup,
		Consumer: c.config.ConsumerName,
		MinIdle:  c.config.ClaimMinIdle,
		Messages: messageIDs,
	}).Result()

	if err != nil {
		return
	}

	// Reprocess claimed messages
	for _, msg := range claimed {
		c.processMessage(streamName, msg)
	}
}

// moveToDeadLetter moves a failed message to the dead letter stream
func (c *Consumer) moveToDeadLetter(stream, msgID string) {
	// Get the original message
	messages, err := c.rdb.XRange(c.ctx, stream, msgID, msgID).Result()
	if err != nil || len(messages) == 0 {
		// Just acknowledge if we can't retrieve
		c.acknowledge(stream, msgID)
		return
	}

	msg := messages[0]

	// Add to dead letter stream
	deadLetterStream := stream + ":deadletter"
	c.rdb.XAdd(c.ctx, &redis.XAddArgs{
		Stream: deadLetterStream,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"original_stream": stream,
			"original_id":     msgID,
			"data":            msg.Values["data"],
			"failed_at":       time.Now().Unix(),
		},
	})

	// Acknowledge original message
	c.acknowledge(stream, msgID)
}

// GetPendingCount returns the number of pending messages
func (c *Consumer) GetPendingCount(stream string) (int64, error) {
	pending, err := c.rdb.XPending(c.ctx, c.streamName(stream), c.config.ConsumerGroup).Result()
	if err != nil {
		return 0, err
	}
	return pending.Count, nil
}

// GetLag returns the consumer lag (messages not yet delivered)
func (c *Consumer) GetLag(stream string) (int64, error) {
	info, err := c.rdb.XInfoGroups(c.ctx, c.streamName(stream)).Result()
	if err != nil {
		return 0, err
	}

	for _, g := range info {
		if g.Name == c.config.ConsumerGroup {
			return g.Lag, nil
		}
	}

	return 0, nil
}

// HealthCheck returns the health status of the consumer
func (c *Consumer) HealthCheck(ctx context.Context) error {
	for _, stream := range c.config.Streams {
		streamName := c.streamName(stream)
		if _, err := c.rdb.XLen(ctx, streamName).Result(); err != nil {
			return fmt.Errorf("stream %s health check failed: %w", stream, err)
		}
	}
	return nil
}
