-- Create invoices table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    subscription_id UUID REFERENCES subscriptions(id),
    organization_id UUID NOT NULL,

    -- Billing Period
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL,

    -- Amounts
    subtotal DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) DEFAULT 0,
    tax_rate DECIMAL(5,4) DEFAULT 0.07,
    tax_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    -- Status
    status VARCHAR(20) DEFAULT 'draft',

    -- Dates
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_at TIMESTAMPTZ,

    -- Documents
    pdf_url VARCHAR(500),
    pdf_generated_at TIMESTAMPTZ,

    -- Details
    line_items JSONB NOT NULL,
    notes TEXT,
    internal_notes TEXT,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_org ON invoices(organization_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due ON invoices(due_date) WHERE status IN ('pending', 'overdue');
CREATE INDEX idx_invoices_number ON invoices(invoice_number);
