package billing

import (
	"fmt"

	"github.com/shopspring/decimal"

	subscriptionmodel "github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/quota"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

// CalculateOverage creates line items for overage usage.
func CalculateOverage(metrics repository.UsageMetrics, tier *subscriptionmodel.SubscriptionTier) ([]LineItem, decimal.Decimal) {
	if tier == nil || !tier.OverageEnabled {
		return nil, decimal.Zero
	}

	lineItems := make([]LineItem, 0)
	total := decimal.Zero

	if tier.QuotaRequests != nil && tier.OverageRatePerRequest != nil {
		over := metrics.TotalRequests - *tier.QuotaRequests
		if over > 0 {
			amount := tier.OverageRatePerRequest.Mul(decimal.NewFromInt(over))
			total = total.Add(amount)
			lineItems = append(lineItems, LineItem{
				Type:        "overage",
				Description: fmt.Sprintf("API Requests Overage (%d requests)", over),
				Quantity:    float64(over),
				UnitPrice:   moneyFromDecimal(*tier.OverageRatePerRequest, tier.Currency),
				Amount:      moneyFromDecimal(amount, tier.Currency),
			})
		}
	}

	if tier.QuotaBandwidthMB != nil && tier.OverageRatePerMB != nil {
		usedMB := quota.BytesToMB(metrics.TotalBandwidthBytes)
		overMB := usedMB - *tier.QuotaBandwidthMB
		if overMB > 0 {
			amount := tier.OverageRatePerMB.Mul(decimal.NewFromInt(overMB))
			total = total.Add(amount)
			lineItems = append(lineItems, LineItem{
				Type:        "overage",
				Description: fmt.Sprintf("Bandwidth Overage (%d MB)", overMB),
				Quantity:    float64(overMB),
				UnitPrice:   moneyFromDecimal(*tier.OverageRatePerMB, tier.Currency),
				Amount:      moneyFromDecimal(amount, tier.Currency),
			})
		}
	}

	return lineItems, total
}
