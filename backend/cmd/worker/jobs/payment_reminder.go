package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/api-usage-billing/backend/internal/domain/billing"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

// PaymentReminder represents an overdue invoice reminder.
type PaymentReminder struct {
	InvoiceID  uuid.UUID
	CustomerID uuid.UUID
	DueDate    time.Time
	Total      float64
	Currency   string
}

// PaymentReminderParams defines reminder filters.
type PaymentReminderParams struct {
	OrganizationID uuid.UUID
	CustomerID     *uuid.UUID
	Limit          int
	Offset         int
	AsOf           time.Time
	OverdueDays    int
}

// BuildPaymentReminders fetches overdue invoices for reminders.
func BuildPaymentReminders(ctx context.Context, repo repository.InvoiceRepository, params PaymentReminderParams) ([]PaymentReminder, int64, error) {
	status := billing.InvoiceStatusOverdue
	filters := repository.InvoiceFilters{
		OrganizationID: params.OrganizationID,
		Status:         &status,
		Limit:          params.Limit,
		Offset:         params.Offset,
	}

	if params.CustomerID != nil {
		filters.CustomerID = *params.CustomerID
	}

	cutoff := time.Time{}
	if !params.AsOf.IsZero() {
		cutoff = params.AsOf
		if params.OverdueDays > 0 {
			cutoff = cutoff.AddDate(0, 0, -params.OverdueDays)
		}
	}

	invoices, total, err := repo.ListForCustomer(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	reminders := make([]PaymentReminder, 0, len(invoices))
	for _, invoice := range invoices {
		if !cutoff.IsZero() && invoice.DueDate.After(cutoff) {
			continue
		}
		amount, _ := invoice.Total.Float64()
		reminders = append(reminders, PaymentReminder{
			InvoiceID:  invoice.ID,
			CustomerID: invoice.CustomerID,
			DueDate:    invoice.DueDate,
			Total:      amount,
			Currency:   invoice.Currency,
		})
	}

	return reminders, total, nil
}
