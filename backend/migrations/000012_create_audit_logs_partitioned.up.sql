-- Create audit_logs table (partitioned by created_at)
CREATE TABLE audit_logs (
    id BIGSERIAL NOT NULL,
    organization_id UUID NOT NULL,

    -- Actor
    actor_type VARCHAR(50) NOT NULL,
    actor_id UUID,
    actor_email VARCHAR(255),
    actor_ip INET,

    -- Action
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,

    -- Details
    old_values JSONB,
    new_values JSONB,
    metadata JSONB DEFAULT '{}',

    -- Timing
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Indexes for performance
CREATE INDEX idx_audit_org ON audit_logs(organization_id, created_at);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_type, actor_id);

-- Default partition to avoid insert failures when partitions are missing
CREATE TABLE IF NOT EXISTS audit_logs_default
    PARTITION OF audit_logs DEFAULT;

-- Comments
COMMENT ON TABLE audit_logs IS 'Immutable audit trail for sensitive operations';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (e.g., api_key.created)';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context for audit entries';
