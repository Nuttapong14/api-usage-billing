package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// CreditNoteType represents credit or debit note type.
type CreditNoteType string

const (
	CreditNoteTypeCredit CreditNoteType = "credit"
	CreditNoteTypeDebit  CreditNoteType = "debit"
)

// CreditNoteStatus represents credit note status.
type CreditNoteStatus string

const (
	CreditNoteStatusIssued    CreditNoteStatus = "issued"
	CreditNoteStatusApplied   CreditNoteStatus = "applied"
	CreditNoteStatusCancelled CreditNoteStatus = "cancelled"
)

// CreditNote represents a credit/debit note for an invoice.
type CreditNote struct {
	ID               uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreditNoteNumber string        `gorm:"size:50;uniqueIndex;not null" json:"credit_note_number"`
	InvoiceID        uuid.UUID     `gorm:"type:uuid;not null;index" json:"invoice_id"`
	CustomerID       uuid.UUID     `gorm:"type:uuid;not null;index" json:"customer_id"`
	OrganizationID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"organization_id"`

	Type   CreditNoteType `gorm:"size:20;not null" json:"type"`
	Amount decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	TaxAmount decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"tax_amount"`
	Total  decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"total"`
	Currency string        `gorm:"size:3;default:'THB'" json:"currency"`

	Reason    string         `gorm:"not null" json:"reason"`
	LineItems datatypes.JSON `gorm:"type:jsonb;not null" json:"line_items"`

	Status    CreditNoteStatus `gorm:"size:20;default:'issued'" json:"status"`
	AppliedAt *time.Time       `json:"applied_at,omitempty"`

	PDFURL   *string    `gorm:"size:500" json:"pdf_url,omitempty"`
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM.
func (CreditNote) TableName() string {
	return "credit_notes"
}
