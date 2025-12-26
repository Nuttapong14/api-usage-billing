package notification

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/Nuttapong14/api-usage-billing/internal/domain/webhook"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/email"
	"github.com/Nuttapong14/api-usage-billing/internal/repository"
)

const (
	defaultMaxAttempts = 5
	secretPrefix       = "whsec_"
)

var validEvents = map[string]struct{}{
	string(EventTestWebhook):            {},
	string(EventUsageQuotaWarning):      {},
	string(EventUsageQuotaExceeded):     {},
	string(EventInvoiceCreated):         {},
	string(EventInvoicePaid):            {},
	string(EventPaymentReceived):        {},
	string(EventSubscriptionUpgraded):   {},
	string(EventSubscriptionDowngraded): {},
	string(EventSubscriptionCancelled):  {},
	string(EventAPIKeyCreated):          {},
	string(EventAPIKeyRevoked):          {},
}

// ServiceImpl implements notification operations.
type ServiceImpl struct {
	repo       repository.WebhookRepository
	dispatcher Dispatcher
	sender     email.Sender
	clock      func() time.Time
}

// NewService creates a new notification service.
func NewService(repo repository.WebhookRepository, dispatcher Dispatcher, sender email.Sender) *ServiceImpl {
	return &ServiceImpl{
		repo:       repo,
		dispatcher: dispatcher,
		sender:     sender,
		clock:      time.Now,
	}
}

// ListWebhooks returns webhook endpoints for a customer.
func (s *ServiceImpl) ListWebhooks(ctx context.Context, params ListWebhooksParams) (*WebhookList, error) {
	if params.OrganizationID == uuid.Nil || params.CustomerID == uuid.Nil {
		return nil, ErrInvalidContext
	}

	endpoints, total, err := s.repo.ListEndpoints(ctx, params.OrganizationID, params.CustomerID, params.IsActive, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	webhooks := make([]Webhook, 0, len(endpoints))
	for _, endpoint := range endpoints {
		webhooks = append(webhooks, toWebhook(endpoint))
	}

	hasMore := int64(params.Offset+params.Limit) < total
	return &WebhookList{
		Data: webhooks,
		Pagination: Pagination{
			Total:   total,
			Limit:   params.Limit,
			Offset:  params.Offset,
			HasMore: hasMore,
		},
	}, nil
}

// CreateWebhook creates a webhook endpoint.
func (s *ServiceImpl) CreateWebhook(ctx context.Context, params CreateWebhookParams) (*WebhookWithSecret, error) {
	if params.OrganizationID == uuid.Nil || params.CustomerID == uuid.Nil {
		return nil, ErrInvalidContext
	}

	events, err := normalizeEvents(params.Events)
	if err != nil {
		return nil, err
	}

	limit := DefaultWebhookLimit
	_, total, err := s.repo.ListEndpoints(ctx, params.OrganizationID, params.CustomerID, nil, 1, 0)
	if err != nil {
		return nil, err
	}
	if total >= int64(limit) {
		return nil, ErrWebhookLimitReached
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	endpoint := webhook.WebhookEndpoint{
		CustomerID:     params.CustomerID,
		OrganizationID: params.OrganizationID,
		URL:            strings.TrimSpace(params.URL),
		Secret:         secret,
		Events:         events,
		Description:    params.Description,
		IsActive:       true,
	}

	if err := s.repo.CreateEndpoint(ctx, &endpoint); err != nil {
		return nil, err
	}

	return &WebhookWithSecret{
		Webhook: toWebhook(endpoint),
		Secret:  secret,
	}, nil
}

// SendTestWebhook sends a test event to a single endpoint.
func (s *ServiceImpl) SendTestWebhook(ctx context.Context, params TestWebhookParams) (*TestWebhookResponse, error) {
	if params.OrganizationID == uuid.Nil || params.CustomerID == uuid.Nil || params.WebhookID == uuid.Nil {
		return nil, ErrInvalidContext
	}

	eventType := strings.TrimSpace(params.EventType)
	if eventType == "" {
		eventType = string(EventTestWebhook)
	}
	if !isValidEvent(eventType) {
		return nil, ErrInvalidWebhookEvent
	}

	endpoint, err := s.repo.GetEndpoint(ctx, params.OrganizationID, params.CustomerID, params.WebhookID)
	if err != nil {
		return nil, err
	}
	if !endpoint.IsActive {
		return nil, ErrWebhookInactive
	}

	payload := WebhookPayload{
		ID:        uuid.New(),
		Type:      eventType,
		CreatedAt: s.clock().UTC(),
		Data: map[string]any{
			"message": "This is a test webhook delivery.",
		},
	}

	result, err := s.dispatchEvent(ctx, endpoint, payload)
	if err != nil && result == nil {
		return nil, err
	}

	response := &TestWebhookResponse{
		DeliveryID:     result.DeliveryID,
		EventType:      result.EventType,
		Status:         result.Status,
		ResponseStatus: result.ResponseStatus,
		ResponseTimeMs: result.ResponseTimeMs,
		Message:        statusMessage(result.Status),
	}
	return response, nil
}

// TriggerWebhook notifies all matching webhook endpoints for an event.
func (s *ServiceImpl) TriggerWebhook(ctx context.Context, params TriggerWebhookParams) ([]DeliveryResult, error) {
	if params.OrganizationID == uuid.Nil || params.CustomerID == uuid.Nil {
		return nil, ErrInvalidContext
	}
	eventType := strings.TrimSpace(params.EventType)
	if !isValidEvent(eventType) {
		return nil, ErrInvalidWebhookEvent
	}

	active := true
	endpoints, _, err := s.repo.ListEndpoints(ctx, params.OrganizationID, params.CustomerID, &active, 100, 0)
	if err != nil {
		return nil, err
	}

	results := make([]DeliveryResult, 0)
	var firstErr error
	for _, endpoint := range endpoints {
		if !eventSubscribed(endpoint.Events, eventType) {
			continue
		}
		payload := WebhookPayload{
			ID:        uuid.New(),
			Type:      eventType,
			CreatedAt: s.clock().UTC(),
			Data:      params.Data,
		}
		result, err := s.dispatchEvent(ctx, &endpoint, payload)
		if result != nil {
			results = append(results, *result)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}

// DispatchPendingDeliveries sends pending webhook deliveries.
func (s *ServiceImpl) DispatchPendingDeliveries(ctx context.Context, params DispatchPendingParams) ([]DeliveryResult, error) {
	deliveries, err := s.repo.ListPendingDeliveries(ctx, params.AsOf, params.Limit)
	if err != nil {
		return nil, err
	}

	results := make([]DeliveryResult, 0, len(deliveries))
	var firstErr error

	for _, delivery := range deliveries {
		if params.RetryOnly {
			if delivery.NextRetryAt == nil || delivery.NextRetryAt.After(params.AsOf) {
				continue
			}
		} else if delivery.NextRetryAt != nil {
			continue
		}

		endpoint, err := s.repo.GetEndpointByID(ctx, delivery.EndpointID)
		if err != nil {
			reason := "endpoint not found"
			now := s.clock().UTC()
			delivery.Status = webhook.DeliveryStatusFailed
			delivery.FailureReason = &reason
			delivery.FailedAt = &now
			_ = s.repo.UpdateDelivery(ctx, &delivery)
			result := DeliveryResult{
				DeliveryID:    delivery.ID,
				EventType:     delivery.EventType,
				Status:        string(webhook.DeliveryStatusFailed),
				FailureReason: &reason,
			}
			results = append(results, result)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		payload := []byte(delivery.Payload)
		result, err := s.dispatchDelivery(ctx, endpoint, &delivery, payload)
		if result != nil {
			results = append(results, *result)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}

// SendEmail sends a notification email.
func (s *ServiceImpl) SendEmail(ctx context.Context, params SendEmailParams) error {
	if s.sender == nil {
		return errors.New("email sender not configured")
	}
	return s.sender.Send(ctx, email.Message{
		To:          params.To,
		Subject:     params.Subject,
		Body:        params.Body,
		ContentType: params.ContentType,
	})
}

func (s *ServiceImpl) dispatchEvent(ctx context.Context, endpoint *webhook.WebhookEndpoint, payload WebhookPayload) (*DeliveryResult, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	delivery := webhook.WebhookDelivery{
		EndpointID:     endpoint.ID,
		OrganizationID: endpoint.OrganizationID,
		EventType:      payload.Type,
		EventID:        payload.ID,
		Payload:        datatypes.JSON(payloadBytes),
		Attempts:       1,
		MaxAttempts:    defaultMaxAttempts,
		Status:         webhook.DeliveryStatusPending,
	}

	if err := s.repo.CreateDelivery(ctx, &delivery); err != nil {
		return nil, err
	}

	return s.dispatchDelivery(ctx, endpoint, &delivery, payloadBytes)
}

func (s *ServiceImpl) dispatchDelivery(ctx context.Context, endpoint *webhook.WebhookEndpoint, delivery *webhook.WebhookDelivery, payload []byte) (*DeliveryResult, error) {
	now := s.clock().UTC()

	if s.dispatcher == nil {
		return nil, errors.New("webhook dispatcher not configured")
	}

	if !endpoint.IsActive {
		reason := "endpoint inactive"
		delivery.Status = webhook.DeliveryStatusFailed
		delivery.FailureReason = &reason
		delivery.FailedAt = &now
		_ = s.repo.UpdateDelivery(ctx, delivery)
		return &DeliveryResult{
			DeliveryID:    delivery.ID,
			EventType:     delivery.EventType,
			Status:        string(delivery.Status),
			FailureReason: &reason,
		}, ErrWebhookInactive
	}

	attempts := delivery.Attempts
	if delivery.NextRetryAt != nil {
		attempts++
	}
	if delivery.MaxAttempts <= 0 {
		delivery.MaxAttempts = defaultMaxAttempts
	}
	if attempts > delivery.MaxAttempts {
		reason := "max attempts reached"
		delivery.Status = webhook.DeliveryStatusFailed
		delivery.FailureReason = &reason
		delivery.FailedAt = &now
		_ = s.repo.UpdateDelivery(ctx, delivery)
		return &DeliveryResult{
			DeliveryID:    delivery.ID,
			EventType:     delivery.EventType,
			Status:        string(delivery.Status),
			FailureReason: &reason,
		}, nil
	}
	delivery.Attempts = attempts

	result, err := s.dispatcher.Dispatch(ctx, *endpoint, payload, delivery.EventType, delivery.EventID)
	if result != nil {
		statusCode := result.StatusCode
		delivery.ResponseStatus = &statusCode
		responseTime := int(result.Duration.Milliseconds())
		delivery.ResponseTimeMs = &responseTime
		if result.ResponseBody != "" {
			body := result.ResponseBody
			delivery.ResponseBody = &body
		}
	}

	isSuccess := err == nil && result != nil && result.StatusCode >= 200 && result.StatusCode < 300
	if isSuccess {
		delivery.Status = webhook.DeliveryStatusDelivered
		delivery.DeliveredAt = &now
		delivery.FailedAt = nil
		delivery.FailureReason = nil
		delivery.NextRetryAt = nil
		endpoint.ConsecutiveFailures = 0
	} else {
		if err != nil {
			reason := err.Error()
			delivery.FailureReason = &reason
		} else if result != nil {
			reason := "non-2xx response"
			delivery.FailureReason = &reason
		}
		endpoint.ConsecutiveFailures++

		if attempts < delivery.MaxAttempts {
			next := now.Add(retryDelay(attempts))
			delivery.NextRetryAt = &next
			delivery.Status = webhook.DeliveryStatusPending
		} else {
			delivery.Status = webhook.DeliveryStatusFailed
			delivery.FailedAt = &now
		}
	}

	endpoint.LastTriggeredAt = &now
	if updateErr := s.repo.UpdateEndpoint(ctx, endpoint); updateErr != nil && err == nil {
		err = updateErr
	}
	if updateErr := s.repo.UpdateDelivery(ctx, delivery); updateErr != nil && err == nil {
		err = updateErr
	}

	return &DeliveryResult{
		DeliveryID:     delivery.ID,
		EventType:      delivery.EventType,
		Status:         string(delivery.Status),
		ResponseStatus: delivery.ResponseStatus,
		ResponseTimeMs: delivery.ResponseTimeMs,
		FailureReason:  delivery.FailureReason,
	}, err
}

func generateSecret() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return secretPrefix + base64.RawURLEncoding.EncodeToString(bytes), nil
}

func normalizeEvents(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, ErrInvalidWebhookEvent
	}
	seen := make(map[string]struct{})
	events := make([]string, 0, len(values))
	for _, value := range values {
		event := strings.ToLower(strings.TrimSpace(value))
		if event == "" {
			continue
		}
		if !isValidEvent(event) {
			return nil, ErrInvalidWebhookEvent
		}
		if _, ok := seen[event]; ok {
			continue
		}
		seen[event] = struct{}{}
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil, ErrInvalidWebhookEvent
	}
	return events, nil
}

func eventSubscribed(events []string, event string) bool {
	for _, candidate := range events {
		if strings.EqualFold(candidate, event) {
			return true
		}
	}
	return false
}

func isValidEvent(event string) bool {
	_, ok := validEvents[strings.ToLower(event)]
	return ok
}

func retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(attempt*attempt) * time.Minute
}

func toWebhook(endpoint webhook.WebhookEndpoint) Webhook {
	return Webhook{
		ID:                  endpoint.ID,
		URL:                 endpoint.URL,
		Description:         endpoint.Description,
		Events:              endpoint.Events,
		IsActive:            endpoint.IsActive,
		LastTriggeredAt:     endpoint.LastTriggeredAt,
		ConsecutiveFailures: endpoint.ConsecutiveFailures,
		CreatedAt:           endpoint.CreatedAt,
		UpdatedAt:           endpoint.UpdatedAt,
	}
}

func statusMessage(status string) string {
	switch status {
	case string(webhook.DeliveryStatusDelivered):
		return "Test event delivered"
	case string(webhook.DeliveryStatusFailed):
		return "Test event failed"
	default:
		return "Test event queued"
	}
}

func ptr(value string) *string {
	return &value
}
