package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EventType defines the type of event
type EventType string

const (
	// API Events
	EventTypeAPIRequest      EventType = "api.request"
	EventTypeAPIError        EventType = "api.error"
	EventTypeRateLimitHit    EventType = "api.rate_limit"

	// Customer Events
	EventTypeCustomerCreated   EventType = "customer.created"
	EventTypeCustomerUpdated   EventType = "customer.updated"
	EventTypeCustomerSuspended EventType = "customer.suspended"
	EventTypeCustomerDeleted   EventType = "customer.deleted"

	// Subscription Events
	EventTypeSubscriptionCreated   EventType = "subscription.created"
	EventTypeSubscriptionUpdated   EventType = "subscription.updated"
	EventTypeSubscriptionCanceled  EventType = "subscription.canceled"
	EventTypeSubscriptionRenewed   EventType = "subscription.renewed"
	EventTypeQuotaExceeded         EventType = "subscription.quota_exceeded"
	EventTypeQuotaWarning          EventType = "subscription.quota_warning"

	// Billing Events
	EventTypeInvoiceCreated    EventType = "billing.invoice_created"
	EventTypeInvoicePaid       EventType = "billing.invoice_paid"
	EventTypeInvoiceFailed     EventType = "billing.invoice_failed"
	EventTypePaymentSucceeded  EventType = "billing.payment_succeeded"
	EventTypePaymentFailed     EventType = "billing.payment_failed"

	// API Key Events
	EventTypeAPIKeyCreated     EventType = "api_key.created"
	EventTypeAPIKeyRotated     EventType = "api_key.rotated"
	EventTypeAPIKeyRevoked     EventType = "api_key.revoked"
	EventTypeAPIKeyExpired     EventType = "api_key.expired"

	// Organization Events
	EventTypeOrgCreated        EventType = "organization.created"
	EventTypeOrgUpdated        EventType = "organization.updated"
	EventTypeOrgSuspended      EventType = "organization.suspended"
)

// Event represents a domain event
type Event struct {
	ID             string                 `json:"id"`
	Type           EventType              `json:"type"`
	Timestamp      time.Time              `json:"timestamp"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	CustomerID     *uuid.UUID             `json:"customer_id,omitempty"`
	APIKeyID       *uuid.UUID             `json:"api_key_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	Payload        map[string]interface{} `json:"payload"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, orgID uuid.UUID, payload map[string]interface{}) *Event {
	return &Event{
		ID:             uuid.NewString(),
		Type:           eventType,
		Timestamp:      time.Now().UTC(),
		OrganizationID: orgID,
		Payload:        payload,
		Metadata:       make(map[string]string),
	}
}

// WithCustomer sets the customer ID
func (e *Event) WithCustomer(customerID uuid.UUID) *Event {
	e.CustomerID = &customerID
	return e
}

// WithAPIKey sets the API key ID
func (e *Event) WithAPIKey(apiKeyID uuid.UUID) *Event {
	e.APIKeyID = &apiKeyID
	return e
}

// WithCorrelation sets the correlation ID
func (e *Event) WithCorrelation(correlationID string) *Event {
	e.CorrelationID = correlationID
	return e
}

// WithMetadata adds metadata
func (e *Event) WithMetadata(key, value string) *Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// Publisher publishes events to Redis Streams
type Publisher struct {
	rdb           redis.UniversalClient
	streamPrefix  string
	maxLen        int64
	defaultStream string
}

// PublisherConfig configures the event publisher
type PublisherConfig struct {
	StreamPrefix  string
	MaxLen        int64  // Maximum stream length (approximate)
	DefaultStream string // Default stream name
}

// DefaultPublisherConfig returns default publisher configuration
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		StreamPrefix:  "events",
		MaxLen:        100000,
		DefaultStream: "main",
	}
}

// NewPublisher creates a new event publisher
func NewPublisher(rdb redis.UniversalClient, config ...PublisherConfig) *Publisher {
	cfg := DefaultPublisherConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &Publisher{
		rdb:           rdb,
		streamPrefix:  cfg.StreamPrefix,
		maxLen:        cfg.MaxLen,
		defaultStream: cfg.DefaultStream,
	}
}

// streamName returns the full stream name
func (p *Publisher) streamName(streamType string) string {
	return fmt.Sprintf("%s:%s", p.streamPrefix, streamType)
}

// Publish publishes an event to a stream
func (p *Publisher) Publish(ctx context.Context, event *Event) (string, error) {
	return p.PublishToStream(ctx, p.defaultStream, event)
}

// PublishToStream publishes an event to a specific stream
func (p *Publisher) PublishToStream(ctx context.Context, stream string, event *Event) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event: %w", err)
	}

	streamName := p.streamName(stream)

	// Use XADD with MAXLEN to prevent unbounded growth
	result, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		MaxLen: p.maxLen,
		Approx: true,
		Values: map[string]interface{}{
			"event_id":   event.ID,
			"event_type": string(event.Type),
			"org_id":     event.OrganizationID.String(),
			"timestamp":  event.Timestamp.Unix(),
			"data":       string(data),
		},
	}).Result()

	if err != nil {
		return "", fmt.Errorf("failed to publish event: %w", err)
	}

	return result, nil
}

// PublishBatch publishes multiple events atomically
func (p *Publisher) PublishBatch(ctx context.Context, events []*Event) ([]string, error) {
	return p.PublishBatchToStream(ctx, p.defaultStream, events)
}

// PublishBatchToStream publishes multiple events to a specific stream
func (p *Publisher) PublishBatchToStream(ctx context.Context, stream string, events []*Event) ([]string, error) {
	streamName := p.streamName(stream)
	pipe := p.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(events))

	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event %d: %w", i, err)
		}

		cmds[i] = pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			MaxLen: p.maxLen,
			Approx: true,
			Values: map[string]interface{}{
				"event_id":   event.ID,
				"event_type": string(event.Type),
				"org_id":     event.OrganizationID.String(),
				"timestamp":  event.Timestamp.Unix(),
				"data":       string(data),
			},
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to publish events: %w", err)
	}

	ids := make([]string, len(cmds))
	for i, cmd := range cmds {
		ids[i] = cmd.Val()
	}

	return ids, nil
}

// PublishToOrg publishes an event to an organization-specific stream
func (p *Publisher) PublishToOrg(ctx context.Context, orgID uuid.UUID, event *Event) (string, error) {
	stream := fmt.Sprintf("org:%s", orgID.String())
	return p.PublishToStream(ctx, stream, event)
}

// StreamInfo returns information about a stream
func (p *Publisher) StreamInfo(ctx context.Context, stream string) (*redis.XInfoStream, error) {
	return p.rdb.XInfoStream(ctx, p.streamName(stream)).Result()
}

// StreamLen returns the length of a stream
func (p *Publisher) StreamLen(ctx context.Context, stream string) (int64, error) {
	return p.rdb.XLen(ctx, p.streamName(stream)).Result()
}

// TrimStream trims a stream to the specified length
func (p *Publisher) TrimStream(ctx context.Context, stream string, maxLen int64) (int64, error) {
	return p.rdb.XTrimMaxLen(ctx, p.streamName(stream), maxLen).Result()
}

// CreateConsumerGroup creates a consumer group for a stream
func (p *Publisher) CreateConsumerGroup(ctx context.Context, stream, group string) error {
	streamName := p.streamName(stream)

	// Create stream if it doesn't exist by adding a dummy entry
	_, err := p.rdb.XGroupCreateMkStream(ctx, streamName, group, "0").Result()
	if err != nil {
		// Ignore BUSYGROUP error (group already exists)
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create consumer group: %w", err)
		}
	}

	return nil
}

// DeleteStream deletes a stream
func (p *Publisher) DeleteStream(ctx context.Context, stream string) error {
	return p.rdb.Del(ctx, p.streamName(stream)).Err()
}

// APIRequestEvent creates an API request event
func APIRequestEvent(orgID uuid.UUID, apiKeyID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypeAPIRequest, orgID, payload).WithAPIKey(apiKeyID)
}

// CustomerCreatedEvent creates a customer created event
func CustomerCreatedEvent(orgID, customerID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypeCustomerCreated, orgID, payload).WithCustomer(customerID)
}

// SubscriptionUpdatedEvent creates a subscription updated event
func SubscriptionUpdatedEvent(orgID, customerID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypeSubscriptionUpdated, orgID, payload).WithCustomer(customerID)
}

// QuotaExceededEvent creates a quota exceeded event
func QuotaExceededEvent(orgID, customerID, apiKeyID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypeQuotaExceeded, orgID, payload).
		WithCustomer(customerID).
		WithAPIKey(apiKeyID)
}

// InvoiceCreatedEvent creates an invoice created event
func InvoiceCreatedEvent(orgID, customerID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypeInvoiceCreated, orgID, payload).WithCustomer(customerID)
}

// PaymentSucceededEvent creates a payment succeeded event
func PaymentSucceededEvent(orgID, customerID uuid.UUID, payload map[string]interface{}) *Event {
	return NewEvent(EventTypePaymentSucceeded, orgID, payload).WithCustomer(customerID)
}
