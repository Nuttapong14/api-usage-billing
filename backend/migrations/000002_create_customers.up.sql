-- Create customers table (API consumers)
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    keycloak_user_id UUID UNIQUE NOT NULL,
    company_name VARCHAR(255),
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    tax_id VARCHAR(50),
    billing_address JSONB,
    settings JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    preferred_language VARCHAR(10) DEFAULT 'th',
    preferred_currency VARCHAR(3) DEFAULT 'THB',
    status VARCHAR(20) DEFAULT 'active',
    suspended_reason TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_customers_org ON customers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_email ON customers(organization_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_keycloak ON customers(keycloak_user_id);
CREATE INDEX idx_customers_status ON customers(organization_id, status) WHERE deleted_at IS NULL;

-- Constraints
ALTER TABLE customers ADD CONSTRAINT chk_customer_status
    CHECK (status IN ('active', 'suspended', 'closed'));
ALTER TABLE customers ADD CONSTRAINT chk_customer_language
    CHECK (preferred_language IN ('th', 'en'));
ALTER TABLE customers ADD CONSTRAINT chk_customer_currency
    CHECK (preferred_currency IN ('THB', 'USD', 'SGD'));

-- Comments
COMMENT ON TABLE customers IS 'API consumers subscribing to organization services';
COMMENT ON COLUMN customers.keycloak_user_id IS 'Link to Keycloak user identity (OIDC sub claim)';
COMMENT ON COLUMN customers.metadata IS 'Extensible custom fields for customer data';
COMMENT ON COLUMN customers.status IS 'Customer account status: active, suspended, closed';
