package usage

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	usageevent "github.com/your-org/api-usage-billing/backend/internal/event/usage"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
	"github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/domain/usage"
)

var (
	ErrInvalidDateRange = errors.New("invalid date range")
)

// ServiceImpl implements usage operations.
type ServiceImpl struct {
	repo    repository.UsageRepository
	counter *CounterStore
	clock   func() time.Time
}

// NewService creates a new usage service.
func NewService(repo repository.UsageRepository, counter *CounterStore) *ServiceImpl {
	return &ServiceImpl{
		repo:    repo,
		counter: counter,
		clock:   time.Now,
	}
}

// RecordUsage persists a usage event and updates counters.
func (s *ServiceImpl) RecordUsage(ctx context.Context, event *usageevent.Event) error {
	if event == nil {
		return errors.New("nil usage event")
	}

	recordedAt := event.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = s.clock().UTC()
	}

	metadata := datatypes.JSON([]byte("{}"))
	if event.Metadata != nil {
		if data, err := json.Marshal(event.Metadata); err == nil {
			metadata = datatypes.JSON(data)
		}
	}

	var userAgent *string
	if event.UserAgent != "" {
		ua := event.UserAgent
		userAgent = &ua
	}

	var clientIP *string
	if event.ClientIP != "" {
		ip := event.ClientIP
		clientIP = &ip
	}

	record := usage.UsageRecord{
		CustomerID:        event.CustomerID,
		APIKeyID:          event.APIKeyID,
		OrganizationID:    event.OrganizationID,
		RequestID:         event.RequestID,
		Endpoint:          event.Endpoint,
		Method:            event.Method,
		StatusCode:        int16(event.StatusCode),
		RequestSizeBytes:  event.RequestSizeBytes,
		ResponseSizeBytes: event.ResponseSizeBytes,
		LatencyMs:         event.LatencyMs,
		RecordedAt:        recordedAt,
		UserAgent:         userAgent,
		ClientIP:          clientIP,
		Metadata:          metadata,
	}

	if err := s.repo.InsertUsageRecords(ctx, []usage.UsageRecord{record}); err != nil {
		return err
	}

	if s.counter == nil {
		return nil
	}

	return s.counter.Increment(
		ctx,
		event.OrganizationID,
		event.CustomerID,
		event.APIKeyID,
		recordedAt,
		event.StatusCode,
		event.RequestSizeBytes,
		event.ResponseSizeBytes,
		event.LatencyMs,
	)
}

// GetCurrentUsage returns current usage for the active billing period.
func (s *ServiceImpl) GetCurrentUsage(ctx context.Context, params CurrentUsageParams) (*CurrentUsage, error) {
	periodStart, periodEnd := s.currentBillingPeriod(ctx, params.CustomerID)

	metrics := UsageMetrics{}
	if s.counter != nil {
		counterMetrics, err := s.counter.GetMetrics(ctx, params.OrganizationID, params.CustomerID, params.APIKeyID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}
		metrics = counterMetrics
	}

	quota := QuotaStatus{}
	if tier, err := s.currentTier(ctx, params.CustomerID); err == nil && tier != nil {
		quota = buildQuotaStatus(tier.QuotaRequests, tier.QuotaBandwidthMB, metrics.TotalRequests, metrics.TotalBandwidthBytes)
	}

	return &CurrentUsage{
		CustomerID: params.CustomerID,
		BillingPeriod: BillingPeriod{
			Start: periodStart,
			End:   periodEnd,
		},
		Usage:       metrics,
		Quota:       quota,
		LastUpdated: s.clock().UTC(),
	}, nil
}

// GetUsageHistory returns usage history for a date range.
func (s *ServiceImpl) GetUsageHistory(ctx context.Context, params UsageHistoryParams) (*UsageHistory, error) {
	if params.End.Before(params.Start) {
		return nil, ErrInvalidDateRange
	}

	granularity := params.Granularity
	if granularity == "" {
		granularity = "daily"
	}

	start := params.Start
	end := params.End
	if granularity == "hourly" {
		end = end.Add(24 * time.Hour)
	}

	points, total, err := s.repo.GetUsageHistory(ctx, params.OrganizationID, params.CustomerID, params.APIKeyID, start, end, granularity, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	data := make([]UsageDataPoint, 0, len(points))
	for _, point := range points {
		data = append(data, UsageDataPoint{
			Timestamp: point.Period,
			Period:    formatPeriodLabel(point.Period, granularity),
			Metrics:   toServiceMetrics(point.Metrics),
		})
	}

	hasMore := int64(params.Offset+params.Limit) < total

	return &UsageHistory{
		Data: data,
		Pagination: Pagination{
			Total:   total,
			Limit:   params.Limit,
			Offset:  params.Offset,
			HasMore: hasMore,
		},
	}, nil
}

// GetUsageBreakdown returns usage breakdown for a date range.
func (s *ServiceImpl) GetUsageBreakdown(ctx context.Context, params UsageBreakdownParams) (*UsageBreakdown, error) {
	if params.End.Before(params.Start) {
		return nil, ErrInvalidDateRange
	}

	endExclusive := params.End.Add(24 * time.Hour)
	totalMetrics, rows, err := s.repo.GetUsageBreakdown(ctx, params.OrganizationID, params.CustomerID, params.APIKeyID, params.Start, endExclusive, params.GroupBy)
	if err != nil {
		return nil, err
	}

	breakdown := make([]UsageBreakdownItem, 0, len(rows))
	for _, row := range rows {
		percentage := 0.0
		if totalMetrics.TotalRequests > 0 {
			percentage = (float64(row.Metrics.TotalRequests) / float64(totalMetrics.TotalRequests)) * 100
		}
		breakdown = append(breakdown, UsageBreakdownItem{
			Key:        row.Key,
			Metrics:    toServiceMetrics(row.Metrics),
			Percentage: percentage,
		})
	}

	return &UsageBreakdown{
		Total:     toServiceMetrics(totalMetrics),
		Breakdown: breakdown,
	}, nil
}

func (s *ServiceImpl) currentBillingPeriod(ctx context.Context, customerID uuid.UUID) (time.Time, time.Time) {
	if sub, err := s.repo.GetCurrentSubscription(ctx, customerID); err == nil && sub != nil {
		start := time.Date(sub.CurrentPeriodStart.Year(), sub.CurrentPeriodStart.Month(), sub.CurrentPeriodStart.Day(), 0, 0, 0, 0, time.UTC)
		end := time.Date(sub.CurrentPeriodEnd.Year(), sub.CurrentPeriodEnd.Month(), sub.CurrentPeriodEnd.Day(), 23, 59, 59, 0, time.UTC)
		return start, end
	}

	now := s.clock().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	return start, end
}

func (s *ServiceImpl) currentTier(ctx context.Context, customerID uuid.UUID) (*subscription.SubscriptionTier, error) {
	sub, err := s.repo.GetCurrentSubscription(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, repository.ErrSubscriptionNotFound
	}
	return s.repo.GetTierByID(ctx, sub.TierID)
}

func buildQuotaStatus(quotaRequests, quotaBandwidthMB *int64, usedRequests int64, usedBytes int64) QuotaStatus {
	return QuotaStatus{
		Requests:    buildQuotaItem(quotaRequests, usedRequests),
		BandwidthMB: buildQuotaItem(quotaBandwidthMB, bytesToMB(usedBytes)),
	}
}

func buildQuotaItem(limit *int64, used int64) QuotaItem {
	if limit == nil {
		return QuotaItem{Limit: nil, Used: used, Remaining: nil, PercentageUsed: 0}
	}
	remaining := *limit - used
	if remaining < 0 {
		remaining = 0
	}
	percentage := 0.0
	if *limit > 0 {
		percentage = (float64(used) / float64(*limit)) * 100
	}
	return QuotaItem{Limit: limit, Used: used, Remaining: &remaining, PercentageUsed: percentage}
}

func bytesToMB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return int64(math.Floor(float64(bytes) / (1024 * 1024)))
}

func toServiceMetrics(metrics repository.UsageMetrics) UsageMetrics {
	return UsageMetrics{
		TotalRequests:       metrics.TotalRequests,
		SuccessfulRequests:  metrics.SuccessfulRequests,
		FailedRequests:      metrics.FailedRequests,
		TotalBandwidthBytes: metrics.TotalBandwidthBytes,
		AvgLatencyMs:        metrics.AvgLatencyMs,
		P95LatencyMs:        metrics.P95LatencyMs,
		P99LatencyMs:        metrics.P99LatencyMs,
	}
}

func formatPeriodLabel(period time.Time, granularity string) string {
	switch granularity {
	case "hourly":
		return period.Format(time.RFC3339)
	case "weekly":
		return period.Format("2006-01-02")
	case "monthly":
		return period.Format("2006-01")
	default:
		return period.Format("2006-01-02")
	}
}
