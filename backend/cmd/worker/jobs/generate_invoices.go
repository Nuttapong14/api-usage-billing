package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"

	billingsvc "github.com/Nuttapong14/api-usage-billing/internal/service/billing"
)

// GenerateInvoicesParams defines inputs for invoice generation.
type GenerateInvoicesParams struct {
	OrganizationID uuid.UUID
	CustomerIDs    []uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	IssueDate      time.Time
	DueDate        time.Time
}

// GenerateInvoices creates invoices for a list of customers.
func GenerateInvoices(ctx context.Context, svc billingsvc.Service, params GenerateInvoicesParams) ([]*billingsvc.Invoice, error) {
	return RunBillingCycle(ctx, svc, BillingCycleParams{
		OrganizationID: params.OrganizationID,
		CustomerIDs:    params.CustomerIDs,
		PeriodStart:    params.PeriodStart,
		PeriodEnd:      params.PeriodEnd,
		IssueDate:      params.IssueDate,
		DueDate:        params.DueDate,
	})
}

// GenerateInvoicesForPreviousMonth generates invoices for the previous month.
func GenerateInvoicesForPreviousMonth(ctx context.Context, svc billingsvc.Service, orgID uuid.UUID, customerIDs []uuid.UUID, now time.Time) ([]*billingsvc.Invoice, error) {
	previousMonth := now.AddDate(0, 0, -1)
	periodStart, periodEnd := BillingPeriodForMonth(previousMonth)
	issueDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return GenerateInvoices(ctx, svc, GenerateInvoicesParams{
		OrganizationID: orgID,
		CustomerIDs:    customerIDs,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		IssueDate:      issueDate,
		DueDate:        issueDate.AddDate(0, 0, 30),
	})
}
