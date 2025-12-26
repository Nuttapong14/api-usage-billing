-- Create webhook_endpoints table (customer webhooks)
CREATE TABLE webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,

    url VARCHAR(500) NOT NULL,
    secret VARCHAR(100) NOT NULL,
    events TEXT[] NOT NULL,

    description VARCHAR(255),
    is_active BOOLEAN DEFAULT true,

    last_triggered_at TIMESTAMPTZ,
    consecutive_failures INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_webhook_endpoints_customer ON webhook_endpoints(customer_id);
CREATE INDEX idx_webhook_endpoints_org ON webhook_endpoints(organization_id);
CREATE INDEX idx_webhook_endpoints_active ON webhook_endpoints(organization_id, is_active);

-- Comments
COMMENT ON TABLE webhook_endpoints IS 'Customer-configured webhook endpoints';
COMMENT ON COLUMN webhook_endpoints.secret IS 'HMAC-SHA256 signing secret';
