package pdf

import "context"

// Generator renders PDF documents.
type Generator interface {
	GenerateInvoice(ctx context.Context, document InvoiceDocument) ([]byte, error)
}

// InvoiceDocument holds invoice data for PDF generation.
type InvoiceDocument struct {
	Invoice  Invoice
	Language string
}

// Invoice represents invoice data for templates.
type Invoice struct {
	InvoiceNumber  string
	CustomerID     string
	BillingPeriodStart string
	BillingPeriodEnd   string
	Status         string
	IssueDate      string
	DueDate        string
	PaidAt         string
	Amounts        InvoiceAmounts
	LineItems      []LineItem
}

// InvoiceAmounts represents invoice totals.
type InvoiceAmounts struct {
	Subtotal  Money
	Discount  Money
	TaxRate   float64
	TaxAmount Money
	Total     Money
}

// LineItem represents a line item in the invoice.
type LineItem struct {
	Type        string
	Description string
	Quantity    float64
	UnitPrice   Money
	Amount      Money
}

// Money represents an amount in a currency.
type Money struct {
	Amount   float64
	Currency string
}
