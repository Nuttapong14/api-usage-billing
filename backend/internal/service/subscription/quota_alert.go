package subscription

import (
	"sort"

	"github.com/your-org/api-usage-billing/backend/internal/pkg/quota"
)

// QuotaAlertLevel indicates alert severity.
type QuotaAlertLevel string

const (
	QuotaAlertWarning  QuotaAlertLevel = "warning"
	QuotaAlertCritical QuotaAlertLevel = "critical"
	QuotaAlertExceeded QuotaAlertLevel = "exceeded"
)

// QuotaAlert represents a triggered quota alert.
type QuotaAlert struct {
	Resource   string          `json:"resource"`
	Percentage float64         `json:"percentage"`
	Threshold  float64         `json:"threshold"`
	Level      QuotaAlertLevel `json:"level"`
}

// QuotaAlertChecker evaluates quota thresholds.
type QuotaAlertChecker struct {
	thresholds []float64
}

// NewQuotaAlertChecker creates a checker with default thresholds when none provided.
func NewQuotaAlertChecker(thresholds ...float64) *QuotaAlertChecker {
	values := thresholds
	if len(values) == 0 {
		values = []float64{80, 95, 100}
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	return &QuotaAlertChecker{thresholds: sorted}
}

// Check returns alerts for a quota item.
func (c *QuotaAlertChecker) Check(resource string, item quota.Item) []QuotaAlert {
	if item.Limit == nil || *item.Limit == 0 {
		return nil
	}

	alerts := make([]QuotaAlert, 0)
	for _, threshold := range c.thresholds {
		if item.Percentage >= threshold {
			alerts = append(alerts, QuotaAlert{
				Resource:   resource,
				Percentage: item.Percentage,
				Threshold:  threshold,
				Level:      alertLevel(threshold),
			})
		}
	}
	return alerts
}

func alertLevel(threshold float64) QuotaAlertLevel {
	switch {
	case threshold >= 100:
		return QuotaAlertExceeded
	case threshold >= 95:
		return QuotaAlertCritical
	default:
		return QuotaAlertWarning
	}
}
