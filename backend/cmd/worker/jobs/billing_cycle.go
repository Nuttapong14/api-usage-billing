package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"

	billingsvc "github.com/Nuttapong14/api-usage-billing/internal/service/billing"
)

// BillingCycleParams defines inputs for running a billing cycle.
type BillingCycleParams struct {
	OrganizationID uuid.UUID
	CustomerIDs    []uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	IssueDate      time.Time
	DueDate        time.Time
}

// RunBillingCycle generates invoices for the given customers and period.
func RunBillingCycle(ctx context.Context, svc billingsvc.Service, params BillingCycleParams) ([]*billingsvc.Invoice, error) {
	invoices := make([]*billingsvc.Invoice, 0, len(params.CustomerIDs))
	for _, customerID := range params.CustomerIDs {
		invoice, err := svc.GenerateInvoice(ctx, billingsvc.GenerateInvoiceParams{
			OrganizationID: params.OrganizationID,
			CustomerID:     customerID,
			PeriodStart:    params.PeriodStart,
			PeriodEnd:      params.PeriodEnd,
			IssueDate:      params.IssueDate,
			DueDate:        params.DueDate,
		})
		if err != nil {
			return invoices, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// BillingPeriodForMonth returns the start and end date for the month containing the date.
func BillingPeriodForMonth(date time.Time) (time.Time, time.Time) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).AddDate(0, 0, -1)
	return start, end
}
