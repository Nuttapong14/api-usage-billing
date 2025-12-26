package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// PaymentStatus represents payment state.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Payment represents a payment record for an invoice.
type Payment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceID      uuid.UUID `gorm:"type:uuid;not null;index" json:"invoice_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`

	Amount   decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	Currency string          `gorm:"size:3;default:'THB'" json:"currency"`

	PaymentMethod       *string        `gorm:"size:50" json:"payment_method,omitempty"`
	PaymentGateway      *string        `gorm:"size:50" json:"payment_gateway,omitempty"`
	GatewayTransactionID *string        `gorm:"size:255" json:"gateway_transaction_id,omitempty"`
	GatewayResponse     datatypes.JSON `gorm:"type:jsonb" json:"gateway_response,omitempty"`

	Status PaymentStatus `gorm:"size:20;default:'pending'" json:"status"`

	PaidAt      *time.Time `json:"paid_at,omitempty"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
	RefundReason *string    `json:"refund_reason,omitempty"`

	Metadata datatypes.JSON `gorm:"default:'{}'" json:"metadata"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (Payment) TableName() string {
	return "payments"
}
