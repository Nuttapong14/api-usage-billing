-- Create payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    organization_id UUID NOT NULL,

    -- Amount
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    -- Gateway
    payment_method VARCHAR(50),
    payment_gateway VARCHAR(50),
    gateway_transaction_id VARCHAR(255),
    gateway_response JSONB,

    -- Status
    status VARCHAR(20) DEFAULT 'pending',

    -- Dates
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    refund_reason TEXT,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_gateway_txn ON payments(gateway_transaction_id);
CREATE INDEX idx_payments_status ON payments(status);
