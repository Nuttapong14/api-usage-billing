package subscription

import (
	"context"
	"errors"

	"github.com/google/uuid"

	subscriptionmodel "github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

// TierServiceImpl implements tier operations.
type TierServiceImpl struct {
	tiers          repository.TierRepository
	subscriptions repository.SubscriptionRepository
}

// NewTierService creates a new tier service.
func NewTierService(tiers repository.TierRepository, subscriptions repository.SubscriptionRepository) *TierServiceImpl {
	return &TierServiceImpl{
		tiers:          tiers,
		subscriptions: subscriptions,
	}
}

// ListAvailableTiers returns public tiers plus the current tier if requested.
func (s *TierServiceImpl) ListAvailableTiers(ctx context.Context, params ListTiersParams) (*TierList, error) {
	tiers, err := s.tiers.ListPublicForOrg(ctx, params.OrganizationID)
	if err != nil {
		return nil, err
	}

	currentTierID := uuid.Nil
	if params.CustomerID != uuid.Nil {
		sub, err := s.subscriptions.GetByCustomerID(ctx, params.CustomerID)
		if err != nil {
			if !errors.Is(err, repository.ErrSubscriptionNotFound) {
				return nil, err
			}
		} else {
			currentTierID = sub.TierID
			if params.IncludeCurrent && !hasTier(tiers, currentTierID) {
				currentTier, err := s.tiers.GetByID(ctx, currentTierID)
				if err != nil {
					if !errors.Is(err, repository.ErrTierNotFound) {
						return nil, err
					}
				} else if currentTier != nil {
					tiers = append(tiers, *currentTier)
				}
			}
		}
	}

	response := make([]Tier, 0, len(tiers))
	for i := range tiers {
		isCurrent := currentTierID != uuid.Nil && tiers[i].ID == currentTierID
		response = append(response, toTier(&tiers[i], isCurrent))
	}

	return &TierList{Tiers: response}, nil
}

func hasTier(tiers []subscriptionmodel.SubscriptionTier, id uuid.UUID) bool {
	for _, tier := range tiers {
		if tier.ID == id {
			return true
		}
	}
	return false
}
