package usage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	infraevents "github.com/Nuttapong14/api-usage-billing/internal/infrastructure/events"
)

// Event represents a usage event emitted by the API gateway.
type Event struct {
	ID                string                 `json:"id"`
	OrganizationID    uuid.UUID              `json:"organization_id"`
	CustomerID        uuid.UUID              `json:"customer_id"`
	APIKeyID          *uuid.UUID             `json:"api_key_id,omitempty"`
	RequestID         string                 `json:"request_id"`
	Endpoint          string                 `json:"endpoint"`
	Method            string                 `json:"method"`
	StatusCode        int                    `json:"status_code"`
	RequestSizeBytes  int                    `json:"request_size_bytes"`
	ResponseSizeBytes int                    `json:"response_size_bytes"`
	LatencyMs         int                    `json:"latency_ms"`
	RecordedAt        time.Time              `json:"recorded_at"`
	UserAgent         string                 `json:"user_agent,omitempty"`
	ClientIP          string                 `json:"client_ip,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// NewEvent creates a new usage event with defaults.
func NewEvent(orgID, customerID uuid.UUID) *Event {
	return &Event{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		CustomerID:     customerID,
		RecordedAt:     time.Now().UTC(),
		Metadata:       make(map[string]interface{}),
	}
}

// IsSuccess returns true if the event represents a successful request.
func (e *Event) IsSuccess() bool {
	return e.StatusCode > 0 && e.StatusCode < 400
}

// TotalBytes returns total transferred bytes for the request.
func (e *Event) TotalBytes() int64 {
	return int64(e.RequestSizeBytes + e.ResponseSizeBytes)
}

// FromEvent converts a generic infrastructure event into a usage event.
func FromEvent(evt *infraevents.Event) (*Event, error) {
	if evt == nil {
		return nil, errors.New("nil event")
	}
	if evt.CustomerID == nil {
		return nil, errors.New("missing customer_id")
	}

	payloadBytes, err := json.Marshal(evt.Payload)
	if err != nil {
		return nil, err
	}

	var payload struct {
		RequestID         string                 `json:"request_id"`
		Endpoint          string                 `json:"endpoint"`
		Method            string                 `json:"method"`
		StatusCode        int                    `json:"status_code"`
		RequestSizeBytes  int                    `json:"request_size_bytes"`
		ResponseSizeBytes int                    `json:"response_size_bytes"`
		LatencyMs         int                    `json:"latency_ms"`
		RecordedAt        string                 `json:"recorded_at"`
		UserAgent         string                 `json:"user_agent"`
		ClientIP          string                 `json:"client_ip"`
		Metadata          map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	recordedAt := evt.Timestamp
	if payload.RecordedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, payload.RecordedAt); err == nil {
			recordedAt = parsed
		}
	}

	return &Event{
		ID:                evt.ID,
		OrganizationID:    evt.OrganizationID,
		CustomerID:        *evt.CustomerID,
		APIKeyID:          evt.APIKeyID,
		RequestID:         payload.RequestID,
		Endpoint:          payload.Endpoint,
		Method:            payload.Method,
		StatusCode:        payload.StatusCode,
		RequestSizeBytes:  payload.RequestSizeBytes,
		ResponseSizeBytes: payload.ResponseSizeBytes,
		LatencyMs:         payload.LatencyMs,
		RecordedAt:        recordedAt,
		UserAgent:         payload.UserAgent,
		ClientIP:          payload.ClientIP,
		Metadata:          payload.Metadata,
	}, nil
}
