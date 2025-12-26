-- Create subscription_tiers table (pricing plans)
CREATE TABLE subscription_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    display_order INT DEFAULT 0,

    -- Pricing
    price_monthly DECIMAL(12,2) NOT NULL,
    price_yearly DECIMAL(12,2),
    currency VARCHAR(3) DEFAULT 'THB',

    -- Quotas (NULL = unlimited)
    quota_requests BIGINT,
    quota_bandwidth_mb BIGINT,
    quota_compute_seconds BIGINT,

    -- Rate Limits
    rate_limit_per_second INT,
    rate_limit_per_minute INT,
    rate_limit_burst INT,

    -- Overage
    overage_enabled BOOLEAN DEFAULT false,
    overage_rate_per_request DECIMAL(10,6),
    overage_rate_per_mb DECIMAL(10,6),

    -- Config
    features JSONB DEFAULT '{}',
    sla_percentage DECIMAL(5,2),
    is_public BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(organization_id, slug)
);

-- Indexes for performance
CREATE INDEX idx_tiers_org ON subscription_tiers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tiers_active ON subscription_tiers(organization_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_tiers_public ON subscription_tiers(organization_id, is_public) WHERE deleted_at IS NULL AND is_active = true;
CREATE INDEX idx_tiers_order ON subscription_tiers(organization_id, display_order) WHERE deleted_at IS NULL;

-- Constraints
ALTER TABLE subscription_tiers ADD CONSTRAINT chk_tier_currency
    CHECK (currency IN ('THB', 'USD', 'SGD'));
ALTER TABLE subscription_tiers ADD CONSTRAINT chk_tier_price_positive
    CHECK (price_monthly > 0);
ALTER TABLE subscription_tiers ADD CONSTRAINT chk_tier_sla
    CHECK (sla_percentage IS NULL OR (sla_percentage >= 0 AND sla_percentage <= 100));

-- Comments
COMMENT ON TABLE subscription_tiers IS 'Configurable pricing tiers per organization';
COMMENT ON COLUMN subscription_tiers.quota_requests IS 'Monthly request quota (NULL = unlimited)';
COMMENT ON COLUMN subscription_tiers.quota_bandwidth_mb IS 'Monthly bandwidth quota in MB (NULL = unlimited)';
COMMENT ON COLUMN subscription_tiers.quota_compute_seconds IS 'Monthly compute seconds quota (NULL = unlimited)';
COMMENT ON COLUMN subscription_tiers.features IS 'JSON object with tier-specific features enabled';
COMMENT ON COLUMN subscription_tiers.overage_rate_per_request IS 'Cost per request beyond quota (if overage enabled)';
COMMENT ON COLUMN subscription_tiers.is_public IS 'Whether tier is shown on public pricing page';
