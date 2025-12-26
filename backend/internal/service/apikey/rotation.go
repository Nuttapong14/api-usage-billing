package apikey

import (
	"context"
	"time"

	apikeydomain "github.com/Nuttapong14/api-usage-billing/internal/domain/apikey"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/hash"
)

// RotateAPIKey creates a new key and schedules the old one for expiration.
func (s *ServiceImpl) RotateAPIKey(ctx context.Context, params RotateAPIKeyParams) (*RotateAPIKeyResponse, error) {
	orgID, err := s.resolveOrganization(ctx, params.OrganizationID, params.CustomerID)
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
	if key.RotationGraceUntil != nil && now.Before(*key.RotationGraceUntil) {
		return nil, ErrAPIKeyRotationInProgress
	}

	gracePeriod := s.gracePeriod
	if params.GracePeriodHours != 0 {
		if params.GracePeriodHours < 1 || params.GracePeriodHours > 168 {
			return nil, ErrInvalidGracePeriod
		}
		gracePeriod = time.Duration(params.GracePeriodHours) * time.Hour
	}

	plainKey, hashedKey, err := hash.GenerateAPIKey(s.keyPrefix)
	if err != nil {
		return nil, err
	}
	prefix := hash.APIKeyPrefix(plainKey)

	newKey := &apikeydomain.APIKey{
		CustomerID:     key.CustomerID,
		KeyHash:        hashedKey,
		KeyPrefix:      prefix,
		Name:           key.Name,
		Description:    key.Description,
		Permissions:    key.Permissions,
		Scopes:         key.Scopes,
		IPWhitelist:    key.IPWhitelist,
		AllowedOrigins: key.AllowedOrigins,
		ExpiresAt:      key.ExpiresAt,
		RotatedFromID:  &key.ID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, newKey); err != nil {
		return nil, err
	}

	graceUntil := now.Add(gracePeriod)
	key.RotationGraceUntil = &graceUntil
	key.UpdatedAt = now
	if err := s.repo.Update(ctx, key); err != nil {
		return nil, err
	}

	permissions, err := decodeStringSlice(key.Permissions)
	if err != nil {
		return nil, err
	}
	scopes, err := decodeStringSlice(key.Scopes)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, hashedKey, buildCachedAPIKey(newKey, orgID, permissions, scopes))
		_ = s.cache.Set(ctx, key.KeyHash, buildCachedAPIKey(key, orgID, permissions, scopes))
	}

	newKeyResponse := buildAPIKeyResponse(newKey, permissions, scopes, now)
	return &RotateAPIKeyResponse{
		NewKey:          plainKey,
		NewAPIKey:       newKeyResponse,
		OldKeyID:        key.ID,
		OldKeyExpiresAt: graceUntil,
		Message:         "New key created. Old key remains valid until expiration.",
	}, nil
}
