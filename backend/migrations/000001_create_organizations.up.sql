-- Create organizations table (root multi-tenant entity)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    logo_url VARCHAR(500),
    settings JSONB DEFAULT '{}',
    billing_email VARCHAR(255),
    billing_address JSONB,
    tax_id VARCHAR(50),
    default_currency VARCHAR(3) DEFAULT 'THB',
    timezone VARCHAR(50) DEFAULT 'Asia/Bangkok',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_active ON organizations(is_active) WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE organizations IS 'API provider companies (multi-tenant root entity)';
COMMENT ON COLUMN organizations.slug IS 'URL-safe identifier for organization';
COMMENT ON COLUMN organizations.settings IS 'Flexible JSON storage for organization-specific configuration';
COMMENT ON COLUMN organizations.billing_address IS 'JSON object with street, city, postal_code, country fields';
COMMENT ON COLUMN organizations.default_currency IS 'ISO 4217 currency code (THB, USD, SGD)';
COMMENT ON COLUMN organizations.deleted_at IS 'Soft delete timestamp for data retention';
