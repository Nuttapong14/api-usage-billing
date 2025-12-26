package usage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	infraevents "github.com/Nuttapong14/api-usage-billing/internal/infrastructure/events"
)

// PublisherConfig configures usage event publishing.
type PublisherConfig struct {
	StreamPrefix string
	Stream       string
	MaxLen       int64
}

// DefaultPublisherConfig returns default publisher settings.
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		StreamPrefix: "usage",
		Stream:       "requests",
		MaxLen:       100000,
	}
}

// Publisher publishes usage events to Redis Streams.
type Publisher struct {
	publisher *infraevents.Publisher
	stream    string
}

// NewPublisher creates a new usage event publisher.
func NewPublisher(rdb redis.UniversalClient, config ...PublisherConfig) *Publisher {
	cfg := DefaultPublisherConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	publisher := infraevents.NewPublisher(rdb, infraevents.PublisherConfig{
		StreamPrefix:  cfg.StreamPrefix,
		MaxLen:        cfg.MaxLen,
		DefaultStream: cfg.Stream,
	})

	return &Publisher{publisher: publisher, stream: cfg.Stream}
}

// Publish sends a usage event to the configured stream.
func (p *Publisher) Publish(ctx context.Context, event *Event) (string, error) {
	if event.RecordedAt.IsZero() {
		event.RecordedAt = time.Now().UTC()
	}

	payload := map[string]interface{}{
		"request_id":          event.RequestID,
		"endpoint":            event.Endpoint,
		"method":              event.Method,
		"status_code":         event.StatusCode,
		"request_size_bytes":  event.RequestSizeBytes,
		"response_size_bytes": event.ResponseSizeBytes,
		"latency_ms":          event.LatencyMs,
		"recorded_at":         event.RecordedAt.Format(time.RFC3339Nano),
		"user_agent":          event.UserAgent,
		"client_ip":           event.ClientIP,
		"metadata":            event.Metadata,
	}

	domainEvent := infraevents.NewEvent(infraevents.EventTypeAPIRequest, event.OrganizationID, payload)
	domainEvent.Timestamp = event.RecordedAt
	if event.ID != "" {
		domainEvent.ID = event.ID
	}
	domainEvent.WithCustomer(event.CustomerID)
	if event.APIKeyID != nil {
		domainEvent.WithAPIKey(*event.APIKeyID)
	}

	return p.publisher.PublishToStream(ctx, p.stream, domainEvent)
}
