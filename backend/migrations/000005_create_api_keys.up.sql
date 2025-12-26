-- Create api_keys table (customer API key credentials)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,

    -- Key Data (NEVER store plain key)
    key_hash VARCHAR(64) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,

    -- Metadata
    name VARCHAR(100),
    description TEXT,

    -- Permissions
    permissions JSONB DEFAULT '["read"]',
    scopes JSONB DEFAULT '[]',

    -- Security
    ip_whitelist TEXT[],
    allowed_origins TEXT[],

    -- Lifecycle
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,

    -- Rotation
    rotated_from_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    rotation_grace_until TIMESTAMPTZ,

    -- Status
    is_active BOOLEAN DEFAULT true,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE (key_hash)
);

-- Indexes for performance
CREATE INDEX idx_api_keys_customer ON api_keys(customer_id);
CREATE INDEX idx_api_keys_active ON api_keys(customer_id, is_active);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_rotation_grace ON api_keys(rotation_grace_until);

-- Comments
COMMENT ON TABLE api_keys IS 'Hashed API keys for customer authentication';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the API key';
COMMENT ON COLUMN api_keys.key_prefix IS 'First 12 characters for display';
COMMENT ON COLUMN api_keys.rotation_grace_until IS 'Old key validity during rotation grace period';
