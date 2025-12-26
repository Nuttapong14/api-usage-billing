package usage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	infraevents "github.com/your-org/api-usage-billing/backend/internal/infrastructure/events"
)

// Handler processes a usage event.
type Handler func(ctx context.Context, event *Event) error

// ConsumerConfig configures the usage event consumer.
type ConsumerConfig struct {
	StreamPrefix   string
	Stream         string
	ConsumerGroup  string
	ConsumerName   string
	BatchSize      int64
	BlockTimeout   time.Duration
	RetryDelay     time.Duration
	MaxRetries     int
	ClaimMinIdle   time.Duration
	ClaimBatchSize int64
	CleanupInterval time.Duration
}

// DefaultConsumerConfig returns default consumer settings.
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		StreamPrefix:   "usage",
		Stream:         "requests",
		ConsumerGroup:  "usage-tracker",
		ConsumerName:   "usage-consumer",
		BatchSize:      50,
		BlockTimeout:   5 * time.Second,
		RetryDelay:     1 * time.Second,
		MaxRetries:     3,
		ClaimMinIdle:   30 * time.Second,
		ClaimBatchSize: 50,
		CleanupInterval: 1 * time.Minute,
	}
}

// Consumer consumes usage events from Redis Streams.
type Consumer struct {
	consumer *infraevents.Consumer
}

// NewConsumer creates a new usage event consumer.
func NewConsumer(rdb redis.UniversalClient, config ...ConsumerConfig) *Consumer {
	cfg := DefaultConsumerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	eventsConfig := infraevents.DefaultConsumerConfig()
	eventsConfig.StreamPrefix = cfg.StreamPrefix
	eventsConfig.Streams = []string{cfg.Stream}
	eventsConfig.ConsumerGroup = cfg.ConsumerGroup
	eventsConfig.ConsumerName = cfg.ConsumerName
	eventsConfig.BatchSize = cfg.BatchSize
	eventsConfig.BlockTimeout = cfg.BlockTimeout
	eventsConfig.RetryDelay = cfg.RetryDelay
	eventsConfig.MaxRetries = cfg.MaxRetries
	eventsConfig.ClaimMinIdle = cfg.ClaimMinIdle
	eventsConfig.ClaimBatchSize = cfg.ClaimBatchSize
	eventsConfig.CleanupInterval = cfg.CleanupInterval

	consumer := infraevents.NewConsumer(rdb, eventsConfig)

	return &Consumer{consumer: consumer}
}

// RegisterHandler registers a handler for usage events.
func (c *Consumer) RegisterHandler(handler Handler) {
	c.consumer.RegisterHandler(infraevents.EventTypeAPIRequest, func(ctx context.Context, event *infraevents.Event) error {
		usageEvent, err := FromEvent(event)
		if err != nil {
			return err
		}
		return handler(ctx, usageEvent)
	})
}

// Start begins consuming events.
func (c *Consumer) Start() error {
	return c.consumer.Start()
}

// Stop stops consuming events.
func (c *Consumer) Stop() error {
	return c.consumer.Stop()
}
