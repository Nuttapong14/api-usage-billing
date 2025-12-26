-- Create credit_notes table
CREATE TABLE credit_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_number VARCHAR(50) UNIQUE NOT NULL,
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    customer_id UUID NOT NULL,
    organization_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    reason TEXT NOT NULL,
    line_items JSONB NOT NULL,

    status VARCHAR(20) DEFAULT 'issued',
    applied_at TIMESTAMPTZ,

    pdf_url VARCHAR(500),
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_credit_notes_invoice ON credit_notes(invoice_id);
CREATE INDEX idx_credit_notes_customer ON credit_notes(customer_id);
