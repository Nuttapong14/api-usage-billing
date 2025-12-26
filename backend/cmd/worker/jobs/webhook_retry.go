package jobs

import (
	"context"
	"time"

	notificationsvc "github.com/Nuttapong14/api-usage-billing/internal/service/notification"
)

// RunWebhookRetry dispatches webhook retry deliveries.
func RunWebhookRetry(ctx context.Context, svc notificationsvc.Service, params WebhookDispatchParams) ([]notificationsvc.DeliveryResult, error) {
	asOf := params.AsOf
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	return svc.DispatchPendingDeliveries(ctx, notificationsvc.DispatchPendingParams{
		AsOf:      asOf,
		Limit:     params.Limit,
		RetryOnly: true,
	})
}
