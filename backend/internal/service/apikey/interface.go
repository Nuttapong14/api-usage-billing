package apikey

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/middleware"
)

var (
	ErrAPIKeyAlreadyRevoked    = errors.New("api key already revoked")
	ErrAPIKeyRotationInProgress = errors.New("api key rotation already in progress")
	ErrAPIKeyLimitReached      = errors.New("api key limit reached")
	ErrInvalidGracePeriod      = errors.New("invalid grace period")
	ErrInvalidPermissions      = errors.New("invalid permissions")
)

// APIKey represents an API key response.
type APIKey struct {
	ID                 uuid.UUID `json:"id"`
	KeyPrefix          string    `json:"key_prefix"`
	Name               string    `json:"name"`
	Description        *string   `json:"description,omitempty"`
	Permissions        []string  `json:"permissions"`
	Scopes             []string  `json:"scopes,omitempty"`
	IPWhitelist        []string  `json:"ip_whitelist,omitempty"`
	AllowedOrigins     []string  `json:"allowed_origins,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP         *string    `json:"last_used_ip,omitempty"`
	IsActive           bool       `json:"is_active"`
	IsRotating         bool       `json:"is_rotating"`
	RotationGraceUntil *time.Time `json:"rotation_grace_until,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
}

// Pagination represents limit/offset pagination.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// APIKeyList represents a list response of API keys.
type APIKeyList struct {
	Data       []APIKey   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// ListAPIKeysParams contains list filters.
type ListAPIKeysParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	IsActive       *bool
	Limit          int
	Offset         int
}

// CreateAPIKeyParams contains inputs for creating a key.
type CreateAPIKeyParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	Name           string
	Description    *string
	Permissions    []string
	Scopes         []string
	IPWhitelist    []string
	AllowedOrigins []string
	ExpiresAt      *time.Time
}

// CreateAPIKeyResponse represents the created key response.
type CreateAPIKeyResponse struct {
	Key     string `json:"key"`
	APIKey  APIKey `json:"api_key"`
	Warning string `json:"warning,omitempty"`
}

// RotateAPIKeyParams contains inputs for rotation.
type RotateAPIKeyParams struct {
	OrganizationID    uuid.UUID
	CustomerID        uuid.UUID
	APIKeyID          uuid.UUID
	GracePeriodHours  int
}

// RotateAPIKeyResponse represents rotation output.
type RotateAPIKeyResponse struct {
	NewKey          string  `json:"new_key"`
	NewAPIKey       APIKey  `json:"new_api_key"`
	OldKeyID        uuid.UUID `json:"old_key_id"`
	OldKeyExpiresAt time.Time `json:"old_key_expires_at"`
	Message         string    `json:"message,omitempty"`
}

// RevokeAPIKeyParams contains revoke inputs.
type RevokeAPIKeyParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	APIKeyID       uuid.UUID
	Reason         *string
}

// RevokeAPIKeyResponse represents revoke output.
type RevokeAPIKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	RevokedAt time.Time `json:"revoked_at"`
	Message   string    `json:"message,omitempty"`
}

// Service defines API key operations.
type Service interface {
	ListAPIKeys(ctx context.Context, params ListAPIKeysParams) (*APIKeyList, error)
	CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*CreateAPIKeyResponse, error)
	RotateAPIKey(ctx context.Context, params RotateAPIKeyParams) (*RotateAPIKeyResponse, error)
	RevokeAPIKey(ctx context.Context, params RevokeAPIKeyParams) (*RevokeAPIKeyResponse, error)

	// API key validation for middleware.
	ValidateKey(ctx context.Context, keyHash string) (*middleware.APIKeyInfo, error)
	RecordUsage(ctx context.Context, keyID uuid.UUID, ip string) error
}
