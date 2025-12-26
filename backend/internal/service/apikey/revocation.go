package apikey

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/infrastructure/redis"
)

const RevocationChannel = "apikey:revoked"

// RevocationEvent is broadcast when a key is revoked.
type RevocationEvent struct {
	KeyHash    string    `json:"key_hash"`
	CustomerID uuid.UUID `json:"customer_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// RevokeAPIKey immediately revokes an API key.
func (s *ServiceImpl) RevokeAPIKey(ctx context.Context, params RevokeAPIKeyParams) (*RevokeAPIKeyResponse, error) {
	_, err := s.resolveOrganization(ctx, params.OrganizationID, params.CustomerID)
	if err != nil {
		return nil, err
	}

	key, err := s.repo.GetByIDForCustomer(ctx, params.CustomerID, params.APIKeyID)
	if err != nil {
		return nil, err
	}

	if key.RevokedAt != nil || !key.IsActive {
		return nil, ErrAPIKeyAlreadyRevoked
	}

	now := s.clock().UTC()
	key.IsActive = false
	key.RevokedAt = &now
	key.RevokedReason = params.Reason
	key.UpdatedAt = now

	if err := s.repo.Update(ctx, key); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, key.KeyHash)
	}

	if s.pubsub != nil {
		_ = s.pubsub.Publish(ctx, RevocationChannel, RevocationEvent{
			KeyHash:    key.KeyHash,
			CustomerID: key.CustomerID,
			Timestamp:  now,
		})
	}

	return &RevokeAPIKeyResponse{
		ID:        key.ID,
		RevokedAt: now,
		Message:   "API key revoked successfully. Changes take effect within 1 second.",
	}, nil
}

// RevocationSubscriber listens for revocation events.
type RevocationSubscriber struct {
	pubsub *redis.PubSub
	cache  *Cache
}

// NewRevocationSubscriber creates a new subscriber.
func NewRevocationSubscriber(pubsub *redis.PubSub, cache *Cache) *RevocationSubscriber {
	return &RevocationSubscriber{pubsub: pubsub, cache: cache}
}

// Start begins listening for revocation events.
func (s *RevocationSubscriber) Start() error {
	if s.pubsub == nil {
		return nil
	}
	return s.pubsub.Subscribe(context.Background(), s.handleMessage, RevocationChannel)
}

func (s *RevocationSubscriber) handleMessage(ctx context.Context, msg *redis.Message) error {
	if s.cache == nil {
		return nil
	}

	var event RevocationEvent
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		return err
	}
	if event.KeyHash == "" {
		return nil
	}
	return s.cache.Delete(ctx, event.KeyHash)
}
