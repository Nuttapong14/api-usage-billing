-- Create usage_records table (partitioned by recorded_at)
CREATE TABLE usage_records (
    id BIGSERIAL,
    customer_id UUID NOT NULL,
    api_key_id UUID,
    organization_id UUID NOT NULL,

    request_id VARCHAR(64) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code SMALLINT NOT NULL,

    request_size_bytes INT DEFAULT 0,
    response_size_bytes INT DEFAULT 0,
    latency_ms INT NOT NULL,

    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_agent VARCHAR(500),
    client_ip INET,
    metadata JSONB DEFAULT '{}',

    PRIMARY KEY (id, recorded_at)
) PARTITION BY RANGE (recorded_at);

-- Indexes for performance
CREATE INDEX idx_usage_customer_time ON usage_records(customer_id, recorded_at);
CREATE INDEX idx_usage_org_time ON usage_records(organization_id, recorded_at);
CREATE INDEX idx_usage_request_id ON usage_records(request_id);

-- Default partition to avoid insert failures when partitions are missing
CREATE TABLE IF NOT EXISTS usage_records_default
    PARTITION OF usage_records DEFAULT;

-- Comments
COMMENT ON TABLE usage_records IS 'High-volume API usage records partitioned by recorded_at';
COMMENT ON COLUMN usage_records.request_id IS 'Idempotency key for API requests';
COMMENT ON COLUMN usage_records.metadata IS 'Additional request metadata';
