package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"

	subscriptionmodel "github.com/Nuttapong14/api-usage-billing/internal/domain/subscription"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/quota"
	"github.com/Nuttapong14/api-usage-billing/internal/repository"
)

// ServiceImpl implements subscription operations.
type ServiceImpl struct {
	subscriptions repository.SubscriptionRepository
	tiers         repository.TierRepository
	usage         repository.UsageRepository
	clock         func() time.Time
}

// NewService creates a new subscription service.
func NewService(
	subscriptions repository.SubscriptionRepository,
	tiers repository.TierRepository,
	usage repository.UsageRepository,
) *ServiceImpl {
	return &ServiceImpl{
		subscriptions: subscriptions,
		tiers:         tiers,
		usage:         usage,
		clock:         time.Now,
	}
}

// GetCurrentSubscription returns the customer's current subscription.
func (s *ServiceImpl) GetCurrentSubscription(ctx context.Context, params GetSubscriptionParams) (*Subscription, error) {
	sub, err := s.subscriptions.GetByCustomerID(ctx, params.CustomerID)
	if err != nil {
		return nil, err
	}

	tier, err := s.tiers.GetByID(ctx, sub.TierID)
	if err != nil {
		return nil, err
	}
	if params.OrganizationID != uuid.Nil && tier.OrganizationID != params.OrganizationID {
		return nil, ErrInvalidTierChange
	}

	usageSummary, err := s.buildUsageSummary(ctx, params.OrganizationID, params.CustomerID, sub, tier)
	if err != nil {
		return nil, err
	}

	pending, err := s.pendingChangeFromMetadata(ctx, sub.Metadata)
	if err != nil {
		return nil, err
	}

	response := buildSubscriptionResponse(sub, tier, usageSummary, pending)
	return &response, nil
}

// UpgradeSubscription upgrades a subscription immediately.
func (s *ServiceImpl) UpgradeSubscription(ctx context.Context, params ChangeSubscriptionParams) (*SubscriptionChangeResponse, error) {
	sub, err := s.subscriptions.GetByCustomerID(ctx, params.CustomerID)
	if err != nil {
		return nil, err
	}

	targetTier, err := s.tiers.GetByID(ctx, params.TierID)
	if err != nil {
		return nil, err
	}
	if params.OrganizationID != uuid.Nil && targetTier.OrganizationID != params.OrganizationID {
		return nil, ErrInvalidTierChange
	}
	if sub.TierID == targetTier.ID {
		return nil, ErrTierAlreadyActive
	}

	currentTier, err := s.tiers.GetByID(ctx, sub.TierID)
	if err != nil {
		return nil, err
	}

	requestedCycle, err := normalizeBillingCycle(params.BillingCycle)
	if err != nil {
		return nil, err
	}
	if requestedCycle == "" {
		requestedCycle = string(sub.BillingCycle)
	}

	if !isUpgrade(currentTier, targetTier, requestedCycle) {
		return nil, ErrInvalidTierChange
	}

	sub.TierID = targetTier.ID
	if requestedCycle != "" {
		sub.BillingCycle = subscriptionmodel.BillingCycle(requestedCycle)
	}

	updatedMetadata, err := updatePendingChange(sub.Metadata, nil)
	if err != nil {
		return nil, err
	}
	sub.Metadata = updatedMetadata

	if err := s.subscriptions.Update(ctx, sub); err != nil {
		return nil, err
	}

	updatedSub, err := s.GetCurrentSubscription(ctx, GetSubscriptionParams{
		OrganizationID: params.OrganizationID,
		CustomerID:     params.CustomerID,
	})
	if err != nil {
		return nil, err
	}

	return &SubscriptionChangeResponse{
		Subscription:  *updatedSub,
		ChangeType:    "immediate",
		EffectiveDate: s.clock().UTC(),
		Message:       "Subscription upgraded",
	}, nil
}

// DowngradeSubscription schedules a downgrade at the next billing period.
func (s *ServiceImpl) DowngradeSubscription(ctx context.Context, params ChangeSubscriptionParams) (*SubscriptionChangeResponse, error) {
	sub, err := s.subscriptions.GetByCustomerID(ctx, params.CustomerID)
	if err != nil {
		return nil, err
	}

	targetTier, err := s.tiers.GetByID(ctx, params.TierID)
	if err != nil {
		return nil, err
	}
	if params.OrganizationID != uuid.Nil && targetTier.OrganizationID != params.OrganizationID {
		return nil, ErrInvalidTierChange
	}
	if sub.TierID == targetTier.ID {
		return nil, ErrTierAlreadyActive
	}

	currentTier, err := s.tiers.GetByID(ctx, sub.TierID)
	if err != nil {
		return nil, err
	}

	requestedCycle, err := normalizeBillingCycle(params.BillingCycle)
	if err != nil {
		return nil, err
	}
	if requestedCycle == "" {
		requestedCycle = string(sub.BillingCycle)
	}

	if !isDowngrade(currentTier, targetTier, requestedCycle) {
		return nil, ErrInvalidTierChange
	}

	effectiveDate := nextPeriodStart(sub.CurrentPeriodEnd)
	pending := pendingChangeMetadata{
		Type:          "downgrade",
		TierID:        targetTier.ID,
		EffectiveDate: effectiveDate,
		CreatedAt:     s.clock().UTC(),
	}
	if params.BillingCycle != "" {
		pending.BillingCycle = requestedCycle
	}

	updatedMetadata, err := updatePendingChange(sub.Metadata, &pending)
	if err != nil {
		return nil, err
	}
	sub.Metadata = updatedMetadata

	if err := s.subscriptions.Update(ctx, sub); err != nil {
		return nil, err
	}

	updatedSub, err := s.GetCurrentSubscription(ctx, GetSubscriptionParams{
		OrganizationID: params.OrganizationID,
		CustomerID:     params.CustomerID,
	})
	if err != nil {
		return nil, err
	}

	return &SubscriptionChangeResponse{
		Subscription:  *updatedSub,
		ChangeType:    "scheduled",
		EffectiveDate: effectiveDate,
		Message:       "Downgrade scheduled",
	}, nil
}

func (s *ServiceImpl) buildUsageSummary(
	ctx context.Context,
	orgID, customerID uuid.UUID,
	sub *subscriptionmodel.Subscription,
	tier *subscriptionmodel.SubscriptionTier,
) (*UsageSummary, error) {
	if s.usage == nil || tier == nil {
		return nil, nil
	}

	start := dateStart(sub.CurrentPeriodStart)
	endExclusive := dateStart(sub.CurrentPeriodEnd).AddDate(0, 0, 1)

	metrics, err := s.usage.GetUsageSummary(ctx, orgID, customerID, nil, start, endExclusive)
	if err != nil {
		return nil, err
	}

	requests := quota.Calculate(metrics.TotalRequests, tier.QuotaRequests)
	bandwidthUsedMB := quota.BytesToMB(metrics.TotalBandwidthBytes)
	bandwidth := quota.Calculate(bandwidthUsedMB, tier.QuotaBandwidthMB)

	return &UsageSummary{
		Requests:    toUsageQuota(requests),
		BandwidthMB: toUsageQuota(bandwidth),
	}, nil
}

func buildSubscriptionResponse(
	sub *subscriptionmodel.Subscription,
	tier *subscriptionmodel.SubscriptionTier,
	usageSummary *UsageSummary,
	pending *PendingChange,
) Subscription {
	responseTier := toTier(tier, true)
	return Subscription{
		ID:            sub.ID,
		CustomerID:    sub.CustomerID,
		Tier:          responseTier,
		Status:        string(sub.Status),
		BillingCycle:  string(sub.BillingCycle),
		CurrentPeriod: BillingPeriod{Start: sub.CurrentPeriodStart, End: sub.CurrentPeriodEnd},
		Usage:         usageSummary,
		PendingChange: pending,
		TrialEndsAt:   sub.TrialEndsAt,
		CancelledAt:   sub.CancelledAt,
		CreatedAt:     sub.CreatedAt,
	}
}

func normalizeBillingCycle(cycle string) (string, error) {
	if cycle == "" {
		return "", nil
	}
	switch cycle {
	case string(subscriptionmodel.BillingCycleMonthly), string(subscriptionmodel.BillingCycleYearly):
		return cycle, nil
	default:
		return "", ErrInvalidBillingCycle
	}
}

func isUpgrade(current, target *subscriptionmodel.SubscriptionTier, cycle string) bool {
	currentPrice := tierPriceForCycle(current, cycle)
	targetPrice := tierPriceForCycle(target, cycle)
	return targetPrice.Cmp(currentPrice) > 0
}

func isDowngrade(current, target *subscriptionmodel.SubscriptionTier, cycle string) bool {
	currentPrice := tierPriceForCycle(current, cycle)
	targetPrice := tierPriceForCycle(target, cycle)
	return targetPrice.Cmp(currentPrice) < 0
}

func tierPriceForCycle(tier *subscriptionmodel.SubscriptionTier, cycle string) decimal.Decimal {
	if tier == nil {
		return decimal.Zero
	}
	if cycle == string(subscriptionmodel.BillingCycleYearly) && tier.PriceYearly != nil {
		return *tier.PriceYearly
	}
	return tier.PriceMonthly
}

func dateStart(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func nextPeriodStart(end time.Time) time.Time {
	if end.IsZero() {
		return end
	}
	return dateStart(end).AddDate(0, 0, 1)
}

type pendingChangeMetadata struct {
	Type          string    `json:"type"`
	TierID        uuid.UUID `json:"tier_id"`
	EffectiveDate time.Time `json:"effective_date"`
	CreatedAt     time.Time `json:"created_at"`
	BillingCycle  string    `json:"billing_cycle,omitempty"`
}

func (s *ServiceImpl) pendingChangeFromMetadata(ctx context.Context, metadata datatypes.JSON) (*PendingChange, error) {
	pending, err := extractPendingChange(metadata)
	if err != nil || pending == nil {
		return nil, nil
	}

	tier, err := s.tiers.GetByID(ctx, pending.TierID)
	if err != nil {
		if errors.Is(err, repository.ErrTierNotFound) {
			return nil, nil
		}
		return nil, err
	}

	responseTier := toTier(tier, false)
	return &PendingChange{
		Type:          pending.Type,
		NewTier:       &responseTier,
		EffectiveDate: pending.EffectiveDate,
		CreatedAt:     pending.CreatedAt,
	}, nil
}

func extractPendingChange(metadata datatypes.JSON) (*pendingChangeMetadata, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var payload struct {
		Pending *pendingChangeMetadata `json:"pending_change"`
	}

	if err := json.Unmarshal(metadata, &payload); err != nil {
		return nil, err
	}

	return payload.Pending, nil
}

func updatePendingChange(metadata datatypes.JSON, pending *pendingChangeMetadata) (datatypes.JSON, error) {
	payload := map[string]interface{}{}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &payload); err != nil {
			payload = map[string]interface{}{}
		}
	}

	if pending == nil {
		delete(payload, "pending_change")
	} else {
		payload["pending_change"] = pending
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(data), nil
}

func toTier(tier *subscriptionmodel.SubscriptionTier, isCurrent bool) Tier {
	if tier == nil {
		return Tier{}
	}

	features := map[string]interface{}{}
	if len(tier.Features) > 0 {
		_ = json.Unmarshal(tier.Features, &features)
	}

	pricing := TierPricing{
		Monthly: moneyFromDecimal(tier.PriceMonthly, tier.Currency),
	}
	if tier.PriceYearly != nil {
		yearly := moneyFromDecimal(*tier.PriceYearly, tier.Currency)
		pricing.Yearly = &yearly
	}

	if tier.OverageEnabled {
		overage := &TierOveragePricing{}
		if tier.OverageRatePerRequest != nil {
			amount := moneyFromDecimal(*tier.OverageRatePerRequest, tier.Currency)
			overage.PerRequest = &amount
		}
		if tier.OverageRatePerMB != nil {
			amount := moneyFromDecimal(*tier.OverageRatePerMB, tier.Currency)
			overage.PerMB = &amount
		}
		if overage.PerRequest != nil || overage.PerMB != nil {
			pricing.Overage = overage
		}
	}

	var sla *float64
	if tier.SLAPercentage != nil {
		value, _ := tier.SLAPercentage.Float64()
		sla = &value
	}

	return Tier{
		ID:            tier.ID,
		Name:          tier.Name,
		Slug:          tier.Slug,
		Description:   tier.Description,
		Pricing:       pricing,
		Quotas:        TierQuotas{Requests: tier.QuotaRequests, BandwidthMB: tier.QuotaBandwidthMB, ComputeSeconds: tier.QuotaComputeSeconds},
		RateLimits:    TierRateLimits{PerSecond: tier.RateLimitPerSecond, PerMinute: tier.RateLimitPerMinute, Burst: tier.RateLimitBurst},
		Features:      features,
		SLAPercentage: sla,
		IsCurrent:     isCurrent,
	}
}

func moneyFromDecimal(amount decimal.Decimal, currency string) Money {
	value, _ := amount.Float64()
	return Money{
		Amount:   value,
		Currency: currency,
	}
}

func toUsageQuota(item quota.Item) UsageQuota {
	return UsageQuota{
		Used:       item.Used,
		Limit:      item.Limit,
		Percentage: item.Percentage,
	}
}
