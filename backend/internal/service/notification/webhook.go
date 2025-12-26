package notification

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/domain/webhook"
	cryptoutil "github.com/your-org/api-usage-billing/backend/internal/pkg/crypto"
)

const (
	defaultWebhookTimeout = 8 * time.Second
	maxResponseBytes      = 1000
	signatureHeader       = "X-Webhook-Signature"
	eventHeader           = "X-Webhook-Event"
	eventIDHeader         = "X-Webhook-Id"
)

// DispatchResult captures a webhook delivery attempt.
type DispatchResult struct {
	StatusCode   int
	ResponseBody string
	Duration     time.Duration
}

// Dispatcher delivers webhook payloads to endpoints.
type Dispatcher interface {
	Dispatch(ctx context.Context, endpoint webhook.WebhookEndpoint, payload []byte, eventType string, eventID uuid.UUID) (*DispatchResult, error)
}

// HTTPDispatcher sends webhook events via HTTP.
type HTTPDispatcher struct {
	client *http.Client
}

// NewHTTPDispatcher creates an HTTP dispatcher with a default timeout.
func NewHTTPDispatcher(client *http.Client) *HTTPDispatcher {
	if client == nil {
		client = &http.Client{Timeout: defaultWebhookTimeout}
	}
	return &HTTPDispatcher{client: client}
}

// Dispatch sends a webhook event to the endpoint.
func (d *HTTPDispatcher) Dispatch(ctx context.Context, endpoint webhook.WebhookEndpoint, payload []byte, eventType string, eventID uuid.UUID) (*DispatchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	signature := cryptoutil.SignSHA256WithPrefix(endpoint.Secret, payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AUB-Webhook/1.0")
	req.Header.Set(signatureHeader, signature)
	if eventType != "" {
		req.Header.Set(eventHeader, eventType)
	}
	if eventID != uuid.Nil {
		req.Header.Set(eventIDHeader, eventID.String())
	}

	start := time.Now()
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))

	return &DispatchResult{
		StatusCode:   resp.StatusCode,
		ResponseBody: string(body),
		Duration:     time.Since(start),
	}, nil
}
