package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// InvoiceStatus represents the lifecycle status of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

// Invoice represents a billing invoice.
type Invoice struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceNumber  string    `gorm:"size:50;uniqueIndex;not null" json:"invoice_number"`
	CustomerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	SubscriptionID *uuid.UUID `gorm:"type:uuid;index" json:"subscription_id,omitempty"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`

	BillingPeriodStart time.Time `gorm:"type:date;not null" json:"billing_period_start"`
	BillingPeriodEnd   time.Time `gorm:"type:date;not null" json:"billing_period_end"`

	Subtotal       decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	DiscountAmount decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"discount_amount"`
	TaxRate        decimal.Decimal `gorm:"type:decimal(5,4);default:0.07" json:"tax_rate"`
	TaxAmount      decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"tax_amount"`
	Total          decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"total"`
	Currency       string          `gorm:"size:3;default:'THB'" json:"currency"`

	Status InvoiceStatus `gorm:"size:20;default:'draft'" json:"status"`

	IssueDate time.Time  `gorm:"type:date;not null" json:"issue_date"`
	DueDate   time.Time  `gorm:"type:date;not null" json:"due_date"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`

	PDFURL         *string    `gorm:"size:500" json:"pdf_url,omitempty"`
	PDFGeneratedAt *time.Time `json:"pdf_generated_at,omitempty"`

	LineItems     datatypes.JSON `gorm:"type:jsonb;not null" json:"line_items"`
	Notes         *string        `json:"notes,omitempty"`
	InternalNotes *string        `json:"internal_notes,omitempty"`
	Metadata      datatypes.JSON `gorm:"default:'{}'" json:"metadata"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (Invoice) TableName() string {
	return "invoices"
}
