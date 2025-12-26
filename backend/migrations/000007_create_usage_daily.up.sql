-- Create usage_daily table (aggregated daily stats)
CREATE TABLE usage_daily (
    id BIGSERIAL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    api_key_id UUID,
    organization_id UUID NOT NULL,
    date DATE NOT NULL,

    total_requests BIGINT DEFAULT 0,
    successful_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    total_request_bytes BIGINT DEFAULT 0,
    total_response_bytes BIGINT DEFAULT 0,

    avg_latency_ms DECIMAL(10,2),
    min_latency_ms INT,
    max_latency_ms INT,
    p50_latency_ms DECIMAL(10,2),
    p95_latency_ms DECIMAL(10,2),
    p99_latency_ms DECIMAL(10,2),

    endpoints_breakdown JSONB DEFAULT '{}',
    status_codes_breakdown JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(customer_id, api_key_id, date)
);

-- Indexes for performance
CREATE INDEX idx_usage_daily_customer ON usage_daily(customer_id, date);
CREATE INDEX idx_usage_daily_org ON usage_daily(organization_id, date);

-- Comments
COMMENT ON TABLE usage_daily IS 'Aggregated daily usage metrics per customer and API key';
COMMENT ON COLUMN usage_daily.endpoints_breakdown IS 'JSON map of endpoint => metrics';
COMMENT ON COLUMN usage_daily.status_codes_breakdown IS 'JSON map of status_code => metrics';
