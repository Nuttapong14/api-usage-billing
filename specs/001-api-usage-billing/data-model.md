# Phase 1: Data Model Design

**Feature**: 001-api-usage-billing | **Date**: 2025-12-26 | **Status**: Complete

## Overview

This document defines the complete data model for the API Usage Analytics & Billing Platform. The design follows PostgreSQL 16 best practices with GORM ORM compatibility, partitioning for high-volume tables, and comprehensive indexing for performance.

---

## Entity Relationship Diagram

```
┌──────────────────┐       ┌───────────────────────┐       ┌──────────────────┐
│   ORGANIZATION   │───┐   │   SUBSCRIPTION_TIER   │       │    KEYCLOAK      │
│   (API Provider) │   │   │   (Pricing Plans)     │       │   (External)     │
└────────┬─────────┘   │   └───────────┬───────────┘       └────────┬─────────┘
         │             │               │                            │
         │ 1:N         │ 1:N           │                            │
         ▼             │               │                            │
┌──────────────────┐   │               │                            │
│     CUSTOMER     │◀──┘               │                            │
│  (API Consumer)  │◀──────────────────┤                            │
│                  │←──────────────────┼────────────────────────────┘
└────────┬─────────┘                   │         keycloak_user_id
         │                             │
         │ 1:N                         │ N:1
         ▼                             ▼
┌──────────────────┐       ┌───────────────────────┐
│     API_KEY      │       │     SUBSCRIPTION      │
│ (Auth Credential)│       │   (Customer → Tier)   │
└────────┬─────────┘       └───────────┬───────────┘
         │                             │
         │ 1:N                         │ 1:N
         ▼                             ▼
┌──────────────────┐       ┌───────────────────────┐
│  USAGE_RECORD    │       │       INVOICE         │
│   (Partitioned)  │       │   (Billing Document)  │
└──────────────────┘       └───────────┬───────────┘
                                       │
         ┌─────────────────────────────┤
         │ 1:N                         │ 1:N
         ▼                             ▼
┌──────────────────┐       ┌───────────────────────┐
│    USAGE_DAILY   │       │       PAYMENT         │
│   (Aggregated)   │       │   (Transaction)       │
└──────────────────┘       └───────────────────────┘

┌──────────────────┐       ┌───────────────────────┐
│ WEBHOOK_ENDPOINT │──────▶│  WEBHOOK_DELIVERY     │
│   (Customer URL) │ 1:N   │   (Audit Log)         │
└──────────────────┘       └───────────────────────┘

┌──────────────────┐       ┌───────────────────────┐
│   AUDIT_LOG      │       │   CREDIT_NOTE         │
│  (All Changes)   │       │   (Adjustments)       │
└──────────────────┘       └───────────────────────┘
```

---

## Core Entities

### 1. Organization (API Provider)

Multi-tenant root entity representing an API provider company.

```sql
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
    deleted_at TIMESTAMPTZ  -- Soft delete
);

CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_active ON organizations(is_active) WHERE deleted_at IS NULL;
```

**GORM Model**:
```go
type Organization struct {
    ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name           string          `gorm:"size:255;not null"`
    Slug           string          `gorm:"size:100;uniqueIndex;not null"`
    LogoURL        *string         `gorm:"size:500"`
    Settings       datatypes.JSON  `gorm:"default:'{}'"`
    BillingEmail   *string         `gorm:"size:255"`
    BillingAddress datatypes.JSON
    TaxID          *string         `gorm:"size:50"`
    DefaultCurrency string         `gorm:"size:3;default:'THB'"`
    Timezone       string          `gorm:"size:50;default:'Asia/Bangkok'"`
    IsActive       bool            `gorm:"default:true"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt  `gorm:"index"`

    // Relations
    Customers         []Customer          `gorm:"foreignKey:OrganizationID"`
    SubscriptionTiers []SubscriptionTier  `gorm:"foreignKey:OrganizationID"`
}
```

---

### 2. Customer (API Consumer)

Customers subscribing to an organization's API services.

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
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
    status VARCHAR(20) DEFAULT 'active',  -- active, suspended, closed
    suspended_reason TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_customers_org ON customers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_email ON customers(organization_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_keycloak ON customers(keycloak_user_id);
CREATE INDEX idx_customers_status ON customers(organization_id, status) WHERE deleted_at IS NULL;
```

**GORM Model**:
```go
type Customer struct {
    ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    OrganizationID   uuid.UUID       `gorm:"type:uuid;not null;index"`
    KeycloakUserID   uuid.UUID       `gorm:"type:uuid;uniqueIndex;not null"`
    CompanyName      *string         `gorm:"size:255"`
    DisplayName      string          `gorm:"size:255;not null"`
    Email            string          `gorm:"size:255;not null"`
    Phone            *string         `gorm:"size:50"`
    TaxID            *string         `gorm:"size:50"`
    BillingAddress   datatypes.JSON
    Settings         datatypes.JSON  `gorm:"default:'{}'"`
    Metadata         datatypes.JSON  `gorm:"default:'{}'"`
    PreferredLanguage string         `gorm:"size:10;default:'th'"`
    PreferredCurrency string         `gorm:"size:3;default:'THB'"`
    Status           string          `gorm:"size:20;default:'active'"`
    SuspendedReason  *string
    IsActive         bool            `gorm:"default:true"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt  `gorm:"index"`

    // Relations
    Organization  Organization   `gorm:"foreignKey:OrganizationID"`
    Subscriptions []Subscription `gorm:"foreignKey:CustomerID"`
    APIKeys       []APIKey       `gorm:"foreignKey:CustomerID"`
    Invoices      []Invoice      `gorm:"foreignKey:CustomerID"`
}
```

---

### 3. Subscription Tier (Pricing Plans)

Configurable pricing tiers per organization.

```sql
CREATE TABLE subscription_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
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

CREATE INDEX idx_tiers_org ON subscription_tiers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tiers_active ON subscription_tiers(organization_id, is_active) WHERE deleted_at IS NULL;
```

**GORM Model**:
```go
type SubscriptionTier struct {
    ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    OrganizationID  uuid.UUID       `gorm:"type:uuid;not null;index"`
    Name            string          `gorm:"size:100;not null"`
    Slug            string          `gorm:"size:100;not null;uniqueIndex:idx_tier_org_slug"`
    Description     *string
    DisplayOrder    int             `gorm:"default:0"`

    // Pricing
    PriceMonthly    decimal.Decimal `gorm:"type:decimal(12,2);not null"`
    PriceYearly     *decimal.Decimal `gorm:"type:decimal(12,2)"`
    Currency        string          `gorm:"size:3;default:'THB'"`

    // Quotas
    QuotaRequests       *int64
    QuotaBandwidthMB    *int64
    QuotaComputeSeconds *int64

    // Rate Limits
    RateLimitPerSecond *int
    RateLimitPerMinute *int
    RateLimitBurst     *int

    // Overage
    OverageEnabled         bool             `gorm:"default:false"`
    OverageRatePerRequest  *decimal.Decimal `gorm:"type:decimal(10,6)"`
    OverageRatePerMB       *decimal.Decimal `gorm:"type:decimal(10,6)"`

    // Config
    Features      datatypes.JSON  `gorm:"default:'{}'"`
    SLAPercentage *decimal.Decimal `gorm:"type:decimal(5,2)"`
    IsPublic      bool            `gorm:"default:true"`
    IsActive      bool            `gorm:"default:true"`

    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    // Relations
    Organization  Organization    `gorm:"foreignKey:OrganizationID"`
    Subscriptions []Subscription  `gorm:"foreignKey:TierID"`
}
```

---

### 4. Subscription (Customer → Tier Link)

Active subscriptions linking customers to tiers.

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    tier_id UUID NOT NULL REFERENCES subscription_tiers(id),

    status VARCHAR(20) DEFAULT 'active',  -- active, paused, cancelled, expired
    billing_cycle VARCHAR(20) DEFAULT 'monthly',  -- monthly, yearly

    -- Billing Period
    current_period_start DATE NOT NULL,
    current_period_end DATE NOT NULL,

    -- Lifecycle
    trial_ends_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,

    -- Usage Tracking
    usage_reset_at TIMESTAMPTZ,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(customer_id)  -- One active subscription per customer
);

CREATE INDEX idx_subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_tier ON subscriptions(tier_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period ON subscriptions(current_period_end);
```

**GORM Model**:
```go
type Subscription struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CustomerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
    TierID     uuid.UUID `gorm:"type:uuid;not null;index"`

    Status       string    `gorm:"size:20;default:'active'"`
    BillingCycle string    `gorm:"size:20;default:'monthly'"`

    CurrentPeriodStart time.Time `gorm:"type:date;not null"`
    CurrentPeriodEnd   time.Time `gorm:"type:date;not null"`

    TrialEndsAt        *time.Time
    PausedAt           *time.Time
    CancelledAt        *time.Time
    CancellationReason *string
    UsageResetAt       *time.Time

    Metadata  datatypes.JSON `gorm:"default:'{}'"`
    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    Customer Customer         `gorm:"foreignKey:CustomerID"`
    Tier     SubscriptionTier `gorm:"foreignKey:TierID"`
    Invoices []Invoice        `gorm:"foreignKey:SubscriptionID"`
}
```

---

### 5. API Key (Authentication Credential)

Secure API key storage with SHA-256 hashing.

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),

    -- Key Data (NEVER store plain key)
    key_hash VARCHAR(64) NOT NULL,  -- SHA-256 hash (64 hex chars)
    key_prefix VARCHAR(12) NOT NULL,  -- For display: "apik_live_a1"

    -- Metadata
    name VARCHAR(100),
    description TEXT,

    -- Permissions
    permissions JSONB DEFAULT '["read"]',  -- ["read", "write", "admin"]
    scopes JSONB DEFAULT '[]',  -- Specific API endpoint scopes

    -- Security
    ip_whitelist INET[],
    allowed_origins TEXT[],

    -- Lifecycle
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,

    -- Rotation
    rotated_from_id UUID REFERENCES api_keys(id),
    rotation_grace_until TIMESTAMPTZ,

    -- Status
    is_active BOOLEAN DEFAULT true,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);  -- Critical for auth
CREATE INDEX idx_api_keys_customer ON api_keys(customer_id);
CREATE INDEX idx_api_keys_active ON api_keys(customer_id, is_active);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);  -- For display lookup
```

**GORM Model**:
```go
type APIKey struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`

    KeyHash   string `gorm:"size:64;not null;index"`
    KeyPrefix string `gorm:"size:12;not null;index"`

    Name        *string
    Description *string

    Permissions     datatypes.JSON `gorm:"default:'[\"read\"]'"`
    Scopes          datatypes.JSON `gorm:"default:'[]'"`
    IPWhitelist     pq.StringArray `gorm:"type:inet[]"`
    AllowedOrigins  pq.StringArray `gorm:"type:text[]"`

    ExpiresAt    *time.Time
    LastUsedAt   *time.Time
    LastUsedIP   *string `gorm:"type:inet"`

    RotatedFromID      *uuid.UUID `gorm:"type:uuid"`
    RotationGraceUntil *time.Time

    IsActive      bool       `gorm:"default:true"`
    RevokedAt     *time.Time
    RevokedReason *string

    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    Customer    Customer     `gorm:"foreignKey:CustomerID"`
    RotatedFrom *APIKey      `gorm:"foreignKey:RotatedFromID"`
    UsageDaily  []UsageDaily `gorm:"foreignKey:APIKeyID"`
}
```

---

### 6. Usage Record (Partitioned by Month)

High-volume table for individual API requests.

```sql
CREATE TABLE usage_records (
    id BIGSERIAL,
    customer_id UUID NOT NULL,
    api_key_id UUID,
    organization_id UUID NOT NULL,

    -- Request Details
    request_id VARCHAR(64) NOT NULL,  -- Idempotency key
    endpoint VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code SMALLINT NOT NULL,

    -- Metrics
    request_size_bytes INT DEFAULT 0,
    response_size_bytes INT DEFAULT 0,
    latency_ms INT NOT NULL,

    -- Timing
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Additional Data
    user_agent VARCHAR(500),
    client_ip INET,
    metadata JSONB DEFAULT '{}',

    PRIMARY KEY (id, recorded_at)
) PARTITION BY RANGE (recorded_at);

-- Create partitions dynamically (example)
CREATE TABLE usage_records_2025_01 PARTITION OF usage_records
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE usage_records_2025_02 PARTITION OF usage_records
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Indexes on partitions (created automatically)
CREATE INDEX idx_usage_customer_time ON usage_records(customer_id, recorded_at);
CREATE INDEX idx_usage_org_time ON usage_records(organization_id, recorded_at);
CREATE INDEX idx_usage_request_id ON usage_records(request_id);  -- Idempotency
```

**GORM Model** (with partition handling):
```go
type UsageRecord struct {
    ID             int64      `gorm:"primaryKey;autoIncrement"`
    CustomerID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_usage_customer_time"`
    APIKeyID       *uuid.UUID `gorm:"type:uuid"`
    OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index:idx_usage_org_time"`

    RequestID         string `gorm:"size:64;not null;index"`
    Endpoint          string `gorm:"size:500;not null"`
    Method            string `gorm:"size:10;not null"`
    StatusCode        int16  `gorm:"not null"`
    RequestSizeBytes  int    `gorm:"default:0"`
    ResponseSizeBytes int    `gorm:"default:0"`
    LatencyMs         int    `gorm:"not null"`

    RecordedAt time.Time `gorm:"not null;default:now();index:idx_usage_customer_time"`

    UserAgent *string        `gorm:"size:500"`
    ClientIP  *string        `gorm:"type:inet"`
    Metadata  datatypes.JSON `gorm:"default:'{}'"`
}

func (UsageRecord) TableName() string {
    return "usage_records"
}
```

---

### 7. Usage Daily (Aggregated)

Pre-aggregated daily statistics for dashboard queries.

```sql
CREATE TABLE usage_daily (
    id BIGSERIAL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id),
    api_key_id UUID REFERENCES api_keys(id),
    organization_id UUID NOT NULL,
    date DATE NOT NULL,

    -- Counters
    total_requests BIGINT DEFAULT 0,
    successful_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    total_request_bytes BIGINT DEFAULT 0,
    total_response_bytes BIGINT DEFAULT 0,

    -- Latency Stats
    avg_latency_ms DECIMAL(10,2),
    min_latency_ms INT,
    max_latency_ms INT,
    p50_latency_ms DECIMAL(10,2),
    p95_latency_ms DECIMAL(10,2),
    p99_latency_ms DECIMAL(10,2),

    -- Breakdown
    endpoints_breakdown JSONB DEFAULT '{}',
    status_codes_breakdown JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(customer_id, api_key_id, date)
);

CREATE INDEX idx_usage_daily_customer ON usage_daily(customer_id, date);
CREATE INDEX idx_usage_daily_org ON usage_daily(organization_id, date);
```

**GORM Model**:
```go
type UsageDaily struct {
    ID             int64      `gorm:"primaryKey;autoIncrement"`
    CustomerID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_usage_daily_unique,priority:1"`
    APIKeyID       *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_usage_daily_unique,priority:2"`
    OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index"`
    Date           time.Time  `gorm:"type:date;not null;uniqueIndex:idx_usage_daily_unique,priority:3"`

    TotalRequests      int64 `gorm:"default:0"`
    SuccessfulRequests int64 `gorm:"default:0"`
    FailedRequests     int64 `gorm:"default:0"`
    TotalRequestBytes  int64 `gorm:"default:0"`
    TotalResponseBytes int64 `gorm:"default:0"`

    AvgLatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)"`
    MinLatencyMs *int
    MaxLatencyMs *int
    P50LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)"`
    P95LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)"`
    P99LatencyMs *decimal.Decimal `gorm:"type:decimal(10,2)"`

    EndpointsBreakdown   datatypes.JSON `gorm:"default:'{}'"`
    StatusCodesBreakdown datatypes.JSON `gorm:"default:'{}'"`

    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    Customer Customer `gorm:"foreignKey:CustomerID"`
    APIKey   *APIKey  `gorm:"foreignKey:APIKeyID"`
}
```

---

### 8. Invoice (Billing Document)

Monthly billing documents with line items.

```sql
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    subscription_id UUID REFERENCES subscriptions(id),
    organization_id UUID NOT NULL,

    -- Billing Period
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL,

    -- Amounts
    subtotal DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) DEFAULT 0,
    tax_rate DECIMAL(5,4) DEFAULT 0.07,
    tax_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    -- Status
    status VARCHAR(20) DEFAULT 'draft',  -- draft, pending, paid, overdue, cancelled, refunded

    -- Dates
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_at TIMESTAMPTZ,

    -- Documents
    pdf_url VARCHAR(500),
    pdf_generated_at TIMESTAMPTZ,

    -- Details
    line_items JSONB NOT NULL,
    notes TEXT,
    internal_notes TEXT,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_org ON invoices(organization_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due ON invoices(due_date) WHERE status IN ('pending', 'overdue');
CREATE INDEX idx_invoices_number ON invoices(invoice_number);
```

**Line Items Structure**:
```json
{
  "items": [
    {
      "type": "subscription",
      "description": "Pro Plan - Monthly",
      "quantity": 1,
      "unit_price": 2999.00,
      "amount": 2999.00
    },
    {
      "type": "overage",
      "description": "API Requests Overage (10,000 requests)",
      "quantity": 10000,
      "unit_price": 0.01,
      "amount": 100.00
    },
    {
      "type": "discount",
      "description": "Loyalty Discount",
      "quantity": 1,
      "unit_price": -100.00,
      "amount": -100.00
    }
  ]
}
```

---

### 9. Payment (Transaction Record)

Payment transactions against invoices.

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    organization_id UUID NOT NULL,

    -- Amount
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    -- Gateway
    payment_method VARCHAR(50),  -- credit_card, bank_transfer, promptpay
    payment_gateway VARCHAR(50),  -- omise, stripe, manual
    gateway_transaction_id VARCHAR(255),
    gateway_response JSONB,

    -- Status
    status VARCHAR(20) DEFAULT 'pending',  -- pending, completed, failed, refunded

    -- Dates
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    refund_reason TEXT,

    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_gateway_txn ON payments(gateway_transaction_id);
CREATE INDEX idx_payments_status ON payments(status);
```

---

### 10. Webhook Endpoint & Delivery

Customer-configured webhook endpoints with delivery tracking.

```sql
CREATE TABLE webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    organization_id UUID NOT NULL,

    url VARCHAR(500) NOT NULL,
    secret VARCHAR(100) NOT NULL,  -- HMAC-SHA256 signing secret
    events TEXT[] NOT NULL,  -- ['usage.quota_warning', 'invoice.created']

    description VARCHAR(255),
    is_active BOOLEAN DEFAULT true,

    -- Stats
    last_triggered_at TIMESTAMPTZ,
    consecutive_failures INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id UUID NOT NULL REFERENCES webhook_endpoints(id),
    organization_id UUID NOT NULL,

    event_type VARCHAR(100) NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL,

    -- Response
    response_status SMALLINT,
    response_body TEXT,
    response_time_ms INT,

    -- Retry
    attempts INT DEFAULT 1,
    max_attempts INT DEFAULT 5,
    next_retry_at TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) DEFAULT 'pending',  -- pending, delivered, failed
    delivered_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_endpoint ON webhook_deliveries(endpoint_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status) WHERE status = 'pending';
CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'pending';
```

---

### 11. Audit Log (Immutable)

Comprehensive audit trail for all sensitive operations.

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    organization_id UUID NOT NULL,

    -- Actor
    actor_type VARCHAR(50) NOT NULL,  -- user, api_key, system, admin
    actor_id UUID,
    actor_email VARCHAR(255),
    actor_ip INET,

    -- Action
    action VARCHAR(100) NOT NULL,  -- api_key.created, invoice.paid, subscription.upgraded
    resource_type VARCHAR(50) NOT NULL,  -- api_key, invoice, subscription
    resource_id UUID,

    -- Details
    old_values JSONB,
    new_values JSONB,
    metadata JSONB DEFAULT '{}',

    -- Timing
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partition by month for efficient retention
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE INDEX idx_audit_org ON audit_logs(organization_id, created_at);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_type, actor_id);
```

---

### 12. Credit Note (Adjustments)

Credit/debit notes for invoice adjustments.

```sql
CREATE TABLE credit_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_number VARCHAR(50) UNIQUE NOT NULL,
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    customer_id UUID NOT NULL,
    organization_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,  -- credit, debit
    amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',

    reason TEXT NOT NULL,
    line_items JSONB NOT NULL,

    status VARCHAR(20) DEFAULT 'issued',  -- issued, applied, cancelled
    applied_at TIMESTAMPTZ,

    pdf_url VARCHAR(500),
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_credit_notes_invoice ON credit_notes(invoice_id);
CREATE INDEX idx_credit_notes_customer ON credit_notes(customer_id);
```

---

## Redis Data Structures

### Real-Time Counters

```
# Monthly usage counter per customer
INCR usage:counter:{org_id}:{customer_id}:{YYYY-MM}:requests
INCRBY usage:counter:{org_id}:{customer_id}:{YYYY-MM}:bytes {size}

# Per-API-key counter
INCR usage:counter:{org_id}:{api_key_id}:{YYYY-MM}:requests

# TTL: 35 days (covers billing grace period)
```

### API Key Cache

```
# Key validation cache
HSET apikey:{hash} customer_id {uuid}
HSET apikey:{hash} permissions {json}
HSET apikey:{hash} ip_whitelist {json}
HSET apikey:{hash} expires_at {timestamp}
HSET apikey:{hash} is_active 1

# TTL: 5 minutes (short for security)
```

### Rate Limiting

```
# Sliding window rate limit
ZADD ratelimit:{customer_id}:{window} {timestamp} {request_id}
ZREMRANGEBYSCORE ratelimit:{customer_id}:{window} 0 {old_timestamp}
ZCARD ratelimit:{customer_id}:{window}

# TTL: window size (1 second or 1 minute)
```

### Pub/Sub Channels

```
# Key revocation broadcast
PUBLISH apikey:revoked {key_hash, customer_id, timestamp}

# Quota alert notifications
PUBLISH quota:alert:{org_id} {customer_id, percentage, type}
```

---

## Migration Strategy

### Initial Migration Order

1. `001_create_organizations.sql`
2. `002_create_customers.sql`
3. `003_create_subscription_tiers.sql`
4. `004_create_subscriptions.sql`
5. `005_create_api_keys.sql`
6. `006_create_usage_records_partitioned.sql`
7. `007_create_usage_daily.sql`
8. `008_create_invoices.sql`
9. `009_create_payments.sql`
10. `010_create_webhook_endpoints.sql`
11. `011_create_webhook_deliveries.sql`
12. `012_create_audit_logs_partitioned.sql`
13. `013_create_credit_notes.sql`
14. `014_create_indexes.sql`
15. `015_create_partition_maintenance.sql`

### Partition Maintenance Function

```sql
CREATE OR REPLACE FUNCTION create_monthly_partitions()
RETURNS void AS $$
DECLARE
    next_month DATE;
    partition_name TEXT;
BEGIN
    -- Create partitions for next 3 months
    FOR i IN 0..2 LOOP
        next_month := DATE_TRUNC('month', NOW() + (i || ' months')::INTERVAL);
        partition_name := 'usage_records_' || TO_CHAR(next_month, 'YYYY_MM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF usage_records
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            next_month,
            next_month + INTERVAL '1 month'
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Schedule monthly execution via pg_cron or external scheduler
```

---

## Data Retention Policy

| Table | Active Retention | Archive Strategy | Deletion |
|-------|-----------------|------------------|----------|
| usage_records | 12 months | Export to MinIO (Parquet) | 7 years |
| usage_daily | 24 months | Keep in DB | Never |
| invoices | Indefinite | None | Never |
| payments | Indefinite | None | Never |
| audit_logs | 12 months | Export to MinIO | 7 years |
| webhook_deliveries | 90 days | None | Delete |

---

**Data Model Complete** | Ready for API Contract Generation
