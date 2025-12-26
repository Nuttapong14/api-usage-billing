package billing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvoiceNotPayable    = errors.New("invoice not payable")
	ErrInvalidInvoiceFilter = errors.New("invalid invoice filter")
	ErrPDFNotAvailable      = errors.New("invoice pdf not available")
	ErrInvoiceGenerationUnavailable = errors.New("invoice generation unavailable")
)

// Money represents a monetary amount.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// LineItem represents a single invoice line item.
type LineItem struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   Money   `json:"unit_price"`
	Amount      Money   `json:"amount"`
}

// InvoiceAmounts represents invoice totals.
type InvoiceAmounts struct {
	Subtotal  Money  `json:"subtotal"`
	Discount  Money  `json:"discount"`
	TaxRate   float64 `json:"tax_rate"`
	TaxAmount Money  `json:"tax_amount"`
	Total     Money  `json:"total"`
}

// BillingPeriod represents the invoice billing period.
type BillingPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// InvoiceDates contains invoice dates.
type InvoiceDates struct {
	IssueDate time.Time  `json:"issue_date"`
	DueDate   time.Time  `json:"due_date"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
}

// Invoice represents a full invoice response.
type Invoice struct {
	ID             uuid.UUID     `json:"id"`
	InvoiceNumber  string        `json:"invoice_number"`
	CustomerID     uuid.UUID     `json:"customer_id"`
	SubscriptionID *uuid.UUID    `json:"subscription_id,omitempty"`
	BillingPeriod  BillingPeriod `json:"billing_period"`
	Amounts        InvoiceAmounts `json:"amounts"`
	LineItems      []LineItem    `json:"line_items"`
	Status         string        `json:"status"`
	Dates          InvoiceDates  `json:"dates"`
	PDFURL         *string       `json:"pdf_url,omitempty"`
	Notes          *string       `json:"notes,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// InvoiceSummary represents a summary invoice for lists.
type InvoiceSummary struct {
	ID            uuid.UUID `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	Total         Money     `json:"total"`
	Status        string    `json:"status"`
	IssueDate     time.Time `json:"issue_date"`
	DueDate       time.Time `json:"due_date"`
	PDFURL        *string   `json:"pdf_url,omitempty"`
}

// Pagination represents limit/offset pagination.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// InvoiceList represents a list response of invoices.
type InvoiceList struct {
	Data       []InvoiceSummary `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// PaymentRequest contains payment initiation fields.
type PaymentRequest struct {
	PaymentMethod string                 `json:"payment_method" validate:"required"`
	ReturnURL     string                 `json:"return_url,omitempty"`
	PaymentSource map[string]interface{} `json:"payment_source,omitempty"`
}

// PaymentResponse represents a payment initiation response.
type PaymentResponse struct {
	PaymentID  uuid.UUID `json:"payment_id"`
	Status     string    `json:"status"`
	RedirectURL *string   `json:"redirect_url,omitempty"`
	QRCode     *string   `json:"qr_code,omitempty"`
}

// ListInvoicesParams contains invoice list filters.
type ListInvoicesParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	Status         string
	StartDate      *time.Time
	EndDate        *time.Time
	Limit          int
	Offset         int
	SortBy         string
	SortDir        string
}

// GetInvoiceParams contains invoice lookup inputs.
type GetInvoiceParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	InvoiceID      uuid.UUID
}

// GetInvoicePDFParams contains PDF request inputs.
type GetInvoicePDFParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	InvoiceID      uuid.UUID
	Language       string
}

// PayInvoiceParams contains payment inputs.
type PayInvoiceParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	InvoiceID      uuid.UUID
	Request        PaymentRequest
}

// GenerateInvoiceParams contains data needed to generate an invoice.
type GenerateInvoiceParams struct {
	OrganizationID uuid.UUID
	CustomerID     uuid.UUID
	SubscriptionID *uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	IssueDate      time.Time
	DueDate        time.Time
}

// InvoicePDF represents a generated PDF.
type InvoicePDF struct {
	Data        []byte
	ContentType string
	URL         *string
}

// Service defines billing operations.
type Service interface {
	ListInvoices(ctx context.Context, params ListInvoicesParams) (*InvoiceList, error)
	GetInvoice(ctx context.Context, params GetInvoiceParams) (*Invoice, error)
	GetInvoicePDF(ctx context.Context, params GetInvoicePDFParams) (*InvoicePDF, error)
	PayInvoice(ctx context.Context, params PayInvoiceParams) (*PaymentResponse, error)
	GenerateInvoice(ctx context.Context, params GenerateInvoiceParams) (*Invoice, error)
}
