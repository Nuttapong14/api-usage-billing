package jobs

import (
	"context"
	"time"

	notificationsvc "github.com/Nuttapong14/api-usage-billing/internal/service/notification"
)

// WebhookDispatchParams defines delivery dispatch inputs.
type WebhookDispatchParams struct {
	AsOf  time.Time
	Limit int
}

// RunWebhookDelivery dispatches pending webhook deliveries.
func RunWebhookDelivery(ctx context.Context, svc notificationsvc.Service, params WebhookDispatchParams) ([]notificationsvc.DeliveryResult, error) {
	asOf := params.AsOf
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	return svc.DispatchPendingDeliveries(ctx, notificationsvc.DispatchPendingParams{
		AsOf:      asOf,
		Limit:     params.Limit,
		RetryOnly: false,
	})
}
