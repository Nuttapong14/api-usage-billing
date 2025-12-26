-- Create subscriptions table (customer to tier link)
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    tier_id UUID NOT NULL REFERENCES subscription_tiers(id) ON DELETE RESTRICT,

    status VARCHAR(20) DEFAULT 'active',
    billing_cycle VARCHAR(20) DEFAULT 'monthly',

    current_period_start DATE NOT NULL,
    current_period_end DATE NOT NULL,

    trial_ends_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,

    usage_reset_at TIMESTAMPTZ,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(customer_id)
);

-- Indexes for performance
CREATE INDEX idx_subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_tier ON subscriptions(tier_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- Constraints
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscription_status
    CHECK (status IN ('active', 'paused', 'cancelled', 'expired'));
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscription_billing_cycle
    CHECK (billing_cycle IN ('monthly', 'yearly'));
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscription_period
    CHECK (current_period_end >= current_period_start);

-- Comments
COMMENT ON TABLE subscriptions IS 'Active subscriptions linking customers to tiers';
COMMENT ON COLUMN subscriptions.usage_reset_at IS 'Timestamp when usage counters reset for billing period';
COMMENT ON COLUMN subscriptions.metadata IS 'Flexible JSON metadata for subscription changes';
