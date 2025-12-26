-- Create webhook_deliveries table (delivery log)
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id UUID NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,

    event_type VARCHAR(100) NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL,

    response_status SMALLINT,
    response_body TEXT,
    response_time_ms INT,

    attempts INT DEFAULT 1,
    max_attempts INT DEFAULT 5,
    next_retry_at TIMESTAMPTZ,

    status VARCHAR(20) DEFAULT 'pending',
    delivered_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_webhook_deliveries_endpoint ON webhook_deliveries(endpoint_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status) WHERE status = 'pending';
CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_webhook_deliveries_org ON webhook_deliveries(organization_id);

-- Comments
COMMENT ON TABLE webhook_deliveries IS 'Webhook delivery log for event notifications';
