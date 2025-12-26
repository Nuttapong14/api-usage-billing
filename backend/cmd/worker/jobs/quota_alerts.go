package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	notificationsvc "github.com/Nuttapong14/api-usage-billing/internal/service/notification"
	subscriptionsvc "github.com/Nuttapong14/api-usage-billing/internal/service/subscription"
)

// QuotaAlertJobParams defines inputs for quota alerts.
type QuotaAlertJobParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	CustomerEmail  string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Alerts         []subscriptionsvc.QuotaAlert
}

// RunQuotaAlerts dispatches quota alerts through webhooks and email.
func RunQuotaAlerts(ctx context.Context, svc notificationsvc.Service, params QuotaAlertJobParams) ([]notificationsvc.DeliveryResult, error) {
	results := make([]notificationsvc.DeliveryResult, 0)
	var firstErr error

	for _, alert := range params.Alerts {
		eventType := quotaEventType(alert)
		payload := map[string]any{
			"resource":     alert.Resource,
			"percentage":   alert.Percentage,
			"threshold":    alert.Threshold,
			"level":        string(alert.Level),
			"period_start": params.PeriodStart.Format(time.RFC3339),
			"period_end":   params.PeriodEnd.Format(time.RFC3339),
		}

		deliveries, err := svc.TriggerWebhook(ctx, notificationsvc.TriggerWebhookParams{
			OrganizationID: params.OrganizationID,
			CustomerID:     params.CustomerID,
			EventType:      eventType,
			Data:           payload,
		})
		results = append(results, deliveries...)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if params.CustomerEmail != "" && len(params.Alerts) > 0 {
		subject := "Usage quota alert"
		body := buildQuotaEmailBody(params.Alerts, params.PeriodStart, params.PeriodEnd)
		if err := svc.SendEmail(ctx, notificationsvc.SendEmailParams{
			To:      []string{params.CustomerEmail},
			Subject: subject,
			Body:    body,
		}); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}

func quotaEventType(alert subscriptionsvc.QuotaAlert) string {
	if alert.Level == subscriptionsvc.QuotaAlertExceeded {
		return string(notificationsvc.EventUsageQuotaExceeded)
	}
	return string(notificationsvc.EventUsageQuotaWarning)
}

func buildQuotaEmailBody(alerts []subscriptionsvc.QuotaAlert, start, end time.Time) string {
	var builder strings.Builder
	builder.WriteString("Usage quota alerts detected.\n\n")
	builder.WriteString(fmt.Sprintf("Billing period: %s to %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02")))
	for _, alert := range alerts {
		builder.WriteString(fmt.Sprintf("- %s: %.1f%% (threshold %.1f%%, %s)\n",
			alert.Resource,
			alert.Percentage,
			alert.Threshold,
			alert.Level,
		))
	}
	builder.WriteString("\nPlease review your usage dashboard for details.\n")
	return builder.String()
}
