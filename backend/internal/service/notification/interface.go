package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidContext      = errors.New("invalid organization or customer context")
	ErrWebhookLimitReached = errors.New("webhook limit reached")
	ErrWebhookInactive     = errors.New("webhook is inactive")
	ErrInvalidWebhookEvent = errors.New("invalid webhook event type")
)

const (
	DefaultWebhookLimit = 10
)

// EventType represents a webhook event type.
type EventType string

const (
	EventTestWebhook           EventType = "test.webhook"
	EventUsageQuotaWarning     EventType = "usage.quota_warning"
	EventUsageQuotaExceeded    EventType = "usage.quota_exceeded"
	EventInvoiceCreated        EventType = "invoice.created"
	EventInvoicePaid           EventType = "invoice.paid"
	EventPaymentReceived       EventType = "payment.received"
	EventSubscriptionUpgraded  EventType = "subscription.upgraded"
	EventSubscriptionDowngraded EventType = "subscription.downgraded"
	EventSubscriptionCancelled EventType = "subscription.cancelled"
	EventAPIKeyCreated         EventType = "api_key.created"
	EventAPIKeyRevoked         EventType = "api_key.revoked"
)

// Webhook represents a webhook endpoint response.
type Webhook struct {
	ID                  uuid.UUID  `json:"id"`
	URL                 string     `json:"url"`
	Description         *string    `json:"description,omitempty"`
	Events              []string   `json:"events"`
	IsActive            bool       `json:"is_active"`
	LastTriggeredAt     *time.Time `json:"last_triggered_at,omitempty"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// WebhookWithSecret includes the signing secret.
type WebhookWithSecret struct {
	Webhook
	Secret string `json:"secret"`
}

// Pagination represents limit/offset pagination.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// WebhookList represents a list response of webhooks.
type WebhookList struct {
	Data       []Webhook `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// WebhookPayload represents the payload sent to webhook endpoints.
type WebhookPayload struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Data      any       `json:"data"`
}

// CreateWebhookParams contains create inputs.
type CreateWebhookParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	URL            string
	Description    *string
	Events         []string
}

// ListWebhooksParams contains list filters.
type ListWebhooksParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	IsActive       *bool
	Limit          int
	Offset         int
}

// TestWebhookParams contains test inputs.
type TestWebhookParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	WebhookID      uuid.UUID
	EventType      string
}

// TestWebhookResponse represents a test webhook response.
type TestWebhookResponse struct {
	DeliveryID     uuid.UUID `json:"delivery_id"`
	EventType      string    `json:"event_type"`
	Status         string    `json:"status"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty"`
	Message        string    `json:"message"`
}

// TriggerWebhookParams represents inputs for triggering an event.
type TriggerWebhookParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	EventType      string
	Data           any
}

// DeliveryResult represents the outcome of a webhook delivery.
type DeliveryResult struct {
	DeliveryID     uuid.UUID `json:"delivery_id"`
	EventType      string    `json:"event_type"`
	Status         string    `json:"status"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty"`
	FailureReason  *string   `json:"failure_reason,omitempty"`
}

// DispatchPendingParams controls delivery dispatching.
type DispatchPendingParams struct {
	AsOf      time.Time
	Limit     int
	RetryOnly bool
}

// SendEmailParams defines email inputs.
type SendEmailParams struct {
	To          []string
	Subject     string
	Body        string
	ContentType string
}

// Service defines notification operations.
type Service interface {
	ListWebhooks(ctx context.Context, params ListWebhooksParams) (*WebhookList, error)
	CreateWebhook(ctx context.Context, params CreateWebhookParams) (*WebhookWithSecret, error)
	SendTestWebhook(ctx context.Context, params TestWebhookParams) (*TestWebhookResponse, error)
	TriggerWebhook(ctx context.Context, params TriggerWebhookParams) ([]DeliveryResult, error)
	DispatchPendingDeliveries(ctx context.Context, params DispatchPendingParams) ([]DeliveryResult, error)
	SendEmail(ctx context.Context, params SendEmailParams) error
}
