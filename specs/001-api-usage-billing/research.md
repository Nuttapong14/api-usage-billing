# Phase 0 Research: API Usage Analytics & Billing Platform

**Feature**: 001-api-usage-billing | **Date**: 2025-12-26 | **Status**: Complete

## Executive Summary

This research validates the technology stack defined in the PRD against the project's constitution principles and industry best practices. All selected technologies align with performance, security, and scalability requirements. Key architectural decisions are documented with alternatives considered and rationale for final choices.

---

## 1. Backend Framework Selection

### Decision: Go 1.22+ with Fiber v2

| Criteria | Go + Fiber | Node.js + Express | Rust + Actix | Java + Spring |
|----------|------------|-------------------|--------------|---------------|
| Performance | ✅ Excellent | ⚠️ Good | ✅ Excellent | ⚠️ Good |
| Memory footprint | ✅ Low (~20MB) | ⚠️ Medium (~100MB) | ✅ Low (~15MB) | ❌ High (~300MB) |
| Concurrency model | ✅ Goroutines | ⚠️ Event loop | ✅ Async/await | ⚠️ Thread pools |
| Learning curve | ✅ Moderate | ✅ Easy | ❌ Steep | ⚠️ Moderate |
| Ecosystem maturity | ✅ Mature | ✅ Mature | ⚠️ Growing | ✅ Mature |
| Cold start time | ✅ <100ms | ⚠️ ~500ms | ✅ <100ms | ❌ 2-5s |

**Rationale**: Go's goroutine-based concurrency model is ideal for processing 50,000 events/sec. Fiber v2 provides Express-like ergonomics while delivering superior performance. Small binary sizes and fast cold starts optimize Kubernetes pod scaling.

**Constitution Alignment**:
- Principle III (Performance): <10ms gateway latency achievable with Fiber
- Principle VIII (Scalability): Horizontal scaling via Kubernetes HPA

---

## 2. Database Technology

### Decision: PostgreSQL 16 + Redis Cluster

| Criteria | PostgreSQL + Redis | MySQL + Redis | MongoDB | CockroachDB |
|----------|-------------------|---------------|---------|-------------|
| ACID compliance | ✅ Full | ✅ Full | ⚠️ Limited | ✅ Full |
| JSONB support | ✅ Native | ⚠️ JSON type | ✅ Native | ✅ Native |
| Partitioning | ✅ Declarative | ⚠️ Manual | ✅ Auto-sharding | ✅ Auto |
| Read replicas | ✅ Native | ✅ Native | ✅ Native | ✅ Native |
| Real-time counters | ❌ N/A | ❌ N/A | ❌ N/A | ❌ N/A |
| Redis Streams | ✅ Required | ✅ Required | ✅ Required | ✅ Required |

**Rationale**: PostgreSQL 16 provides ACID transactions for billing accuracy (99.99% requirement), JSONB for flexible settings storage, and declarative partitioning for usage_records tables. Redis Cluster provides real-time counters, event streaming via Redis Streams, and sub-second pub/sub for API key revocation.

**Constitution Alignment**:
- Principle VI (Data Integrity): PostgreSQL ACID transactions
- Principle III (Performance): Redis for <100ms P99 counter updates
- Principle VIII (Scalability): Redis Cluster (6 nodes), PostgreSQL read replicas

### Schema Partitioning Strategy

```sql
-- Usage records partitioned by month for efficient queries and archival
CREATE TABLE usage_records (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    api_key_id UUID NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    ...
) PARTITION BY RANGE (recorded_at);

-- Create monthly partitions
CREATE TABLE usage_records_2025_01 PARTITION OF usage_records
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

**Retention**: 12 months active → archive to MinIO → delete after 7 years

---

## 3. API Gateway Selection

### Decision: Kong Gateway (OSS)

| Criteria | Kong OSS | AWS API Gateway | Traefik | Envoy |
|----------|----------|-----------------|---------|-------|
| Self-hosted | ✅ Yes | ❌ AWS-only | ✅ Yes | ✅ Yes |
| Plugin ecosystem | ✅ Rich | ⚠️ Limited | ⚠️ Middleware | ⚠️ Filters |
| Rate limiting | ✅ Native | ✅ Native | ⚠️ Plugin | ⚠️ Filter |
| Custom plugins | ✅ Lua/Go | ❌ Lambda | ✅ Go | ✅ WASM |
| Declarative config | ✅ Kong Deck | ⚠️ CloudFormation | ✅ YAML | ✅ YAML |
| Latency overhead | ⚠️ ~3-5ms | ⚠️ ~5-10ms | ✅ ~1-2ms | ✅ ~1-2ms |

**Rationale**: Kong's rich plugin ecosystem enables rate limiting, authentication, and usage tracking without custom development. Self-hosted deployment aligns with PRD requirement for customer infrastructure deployment. The ~3-5ms latency overhead is acceptable within <10ms budget.

**Custom Kong Plugin Requirements**:
1. **Usage Tracking Plugin**: Capture request metadata → publish to Redis Stream
2. **API Key Validation Plugin**: Hash lookup → Redis cache → quota check
3. **Rate Limit Enhancement**: Per-tier rate limits with burst handling

---

## 4. Authentication & Authorization

### Decision: Keycloak with OIDC/OAuth2

| Criteria | Keycloak | Auth0 | Okta | Custom JWT |
|----------|----------|-------|------|------------|
| Self-hosted | ✅ Yes | ❌ SaaS | ❌ SaaS | ✅ Yes |
| Multi-tenant | ✅ Realms | ✅ Orgs | ✅ Orgs | ⚠️ Custom |
| MFA support | ✅ Native | ✅ Native | ✅ Native | ❌ Custom |
| Social login | ✅ Native | ✅ Native | ✅ Native | ❌ Custom |
| RBAC | ✅ Native | ✅ Native | ✅ Native | ⚠️ Custom |
| Cost | ✅ Free | ⚠️ $$ | ⚠️ $$$ | ✅ Free |

**Rationale**: Keycloak provides enterprise-grade authentication with multi-tenant support (realms per organization), MFA, and RBAC without SaaS dependencies. Aligns with self-hosted deployment model.

**API Key Authentication Flow**:
```
Request → Kong → Hash API Key (SHA-256) → Redis Lookup → Validate:
  ├── Permissions match request scope
  ├── IP in whitelist (if configured)
  ├── Key not expired
  └── Key not revoked
→ Pass/Reject
```

**Constitution Alignment**:
- Principle I (Security): SHA-256 hashing, zero-trust validation
- Principle VII (API Stability): 24-hour key rotation grace period

---

## 5. Event Streaming Architecture

### Decision: Redis Streams (not Kafka)

| Criteria | Redis Streams | Apache Kafka | RabbitMQ | NATS |
|----------|---------------|--------------|----------|------|
| Throughput | ✅ 100K+ msg/s | ✅ 1M+ msg/s | ⚠️ 50K msg/s | ✅ 200K+ msg/s |
| Exactly-once | ✅ ACK-based | ✅ Transactions | ⚠️ At-least-once | ⚠️ At-least-once |
| Persistence | ✅ AOF/RDB | ✅ Disk-based | ✅ Disk-based | ⚠️ Optional |
| Consumer groups | ✅ Native | ✅ Native | ⚠️ Competing | ⚠️ Queue groups |
| Operational overhead | ✅ Low | ❌ High | ⚠️ Medium | ✅ Low |
| Already in stack | ✅ Yes (counters) | ❌ Additional | ❌ Additional | ❌ Additional |

**Rationale**: Redis Streams provides exactly-once semantics via XACK, consumer groups for parallel processing, and is already required for real-time counters. Avoids operational overhead of running a separate Kafka cluster. 50,000 events/sec is well within Redis Streams capacity.

**Event Processing Architecture**:
```
Kong Plugin → XADD usage_events → Consumer Group (N workers) →
  ├── INCR redis_counter:{customer}:{month}
  ├── INSERT usage_records (batch write)
  └── Check quota thresholds → Trigger alerts
→ XACK (exactly-once)
```

---

## 6. Frontend Framework

### Decision: Next.js 14 with shadcn/ui

| Criteria | Next.js 14 | Remix | SvelteKit | Nuxt 3 |
|----------|------------|-------|-----------|--------|
| SSR/SSG | ✅ Both | ✅ Both | ✅ Both | ✅ Both |
| React ecosystem | ✅ Full | ✅ Full | ❌ Svelte | ❌ Vue |
| App Router | ✅ Stable | ⚠️ Different model | ✅ File-based | ✅ File-based |
| Vercel edge | ✅ Native | ⚠️ Limited | ⚠️ Limited | ⚠️ Limited |
| Component libraries | ✅ Rich | ✅ Rich | ⚠️ Growing | ⚠️ Growing |
| TypeScript | ✅ Native | ✅ Native | ✅ Native | ✅ Native |

**Rationale**: Next.js 14 with App Router provides the most mature React ecosystem, excellent TypeScript support, and production-ready SSR/SSG. shadcn/ui provides accessible, customizable components that align with Tailwind CSS. Turborepo enables efficient monorepo management for shared packages.

**Constitution Alignment**:
- Principle IV (UX Consistency): Unified shadcn/ui design system across portals
- Principle IV (Localization): i18n package with Thai/English support

---

## 7. Observability Stack

### Decision: OpenTelemetry + Prometheus + Grafana + Loki + Jaeger

| Component | Selected | Alternative | Rationale |
|-----------|----------|-------------|-----------|
| Instrumentation | OpenTelemetry | Proprietary SDKs | Vendor-neutral, future-proof |
| Metrics | Prometheus | InfluxDB | De facto K8s standard, PromQL |
| Visualization | Grafana | Datadog | Self-hosted, rich ecosystem |
| Logging | Loki | Elasticsearch | Lightweight, Grafana-native |
| Tracing | Jaeger | Zipkin | OpenTelemetry native, rich UI |

**Key Dashboards Required**:
1. **Provider Analytics**: Revenue MTD, API calls, top customers
2. **Customer Usage**: Real-time usage, quota progress, billing history
3. **System Health**: Request rates, error rates, latencies, queue depths
4. **Alerts**: Quota thresholds (80%, 95%, 100%), payment overdue, error spikes

**Constitution Alignment**:
- Principle V (Observability): All 4 pillars covered (metrics, logs, traces, dashboards)
- Principle V (Alerting): Prometheus Alertmanager integration

---

## 8. PDF Invoice Generation

### Decision: Go-based PDF generation with go-pdf

| Criteria | go-pdf | wkhtmltopdf | Puppeteer | LaTeX |
|----------|--------|-------------|-----------|-------|
| Performance | ✅ Fast | ⚠️ Medium | ❌ Slow | ⚠️ Medium |
| Dependencies | ✅ None | ❌ System binary | ❌ Chrome | ❌ TeX distribution |
| Thai font support | ✅ TTF embed | ✅ CSS fonts | ✅ CSS fonts | ✅ XeLaTeX |
| Template flexibility | ⚠️ Code-based | ✅ HTML/CSS | ✅ HTML/CSS | ⚠️ LaTeX |
| Container-friendly | ✅ Yes | ⚠️ Binary needed | ❌ Chrome needed | ⚠️ Large image |

**Rationale**: Pure Go solution eliminates external dependencies, reducing container size and security surface. TTF font embedding supports Thai language requirements. Template-based approach with reusable components for invoice structure.

**Invoice Template Structure**:
- Header: Organization branding, invoice number, dates
- Customer: Company name, tax ID, address
- Line Items: Subscription fee, overage charges, discounts
- Tax: VAT 7% calculation with breakdown
- Footer: Payment terms, bank details, legal text

---

## 9. Background Job Processing

### Decision: Asynq (Redis-backed)

| Criteria | Asynq | Temporal | Celery | Custom |
|----------|-------|----------|--------|--------|
| Language | ✅ Go-native | ⚠️ SDK-based | ❌ Python | ✅ Go |
| Backend | ✅ Redis | ⚠️ PostgreSQL | ⚠️ Redis/RabbitMQ | ⚠️ Any |
| Scheduled jobs | ✅ Native | ✅ Native | ✅ Native | ⚠️ Custom |
| Retry logic | ✅ Built-in | ✅ Built-in | ✅ Built-in | ⚠️ Custom |
| Complexity | ✅ Simple | ❌ Complex | ⚠️ Medium | ⚠️ Variable |

**Rationale**: Asynq is Go-native, uses existing Redis infrastructure, and provides scheduled job support for billing CronJobs. Simple API with built-in retry logic and dead letter queue.

**Scheduled Jobs**:
1. **Monthly Invoice Generation**: 1st of month, 00:01 UTC
2. **Daily Usage Aggregation**: Nightly rollup to usage_daily table
3. **Payment Reminders**: 7 days before due date
4. **Quota Threshold Checks**: Every minute for real-time alerts
5. **Report Generation**: Weekly/monthly analytics reports

---

## 10. Security Considerations

### API Key Security Implementation

```go
// Key generation (shown once)
func GenerateAPIKey() (plainKey string, hashedKey string) {
    // 32 bytes = 256 bits of entropy
    randomBytes := make([]byte, 32)
    crypto.Read(randomBytes)

    // Prefix for identification: apik_live_ or apik_test_
    plainKey = "apik_live_" + base64.URLEncoding.EncodeToString(randomBytes)

    // SHA-256 hash for storage
    hash := sha256.Sum256([]byte(plainKey))
    hashedKey = hex.EncodeToString(hash[:])

    return plainKey, hashedKey
}

// Key validation (every request)
func ValidateAPIKey(providedKey string) (*APIKey, error) {
    hash := sha256.Sum256([]byte(providedKey))
    hashedKey := hex.EncodeToString(hash[:])

    // Redis lookup first (cache)
    cached, err := redis.Get("apikey:" + hashedKey)
    if err == nil {
        return cached, nil
    }

    // Database fallback
    return db.FindAPIKeyByHash(hashedKey)
}
```

### Key Revocation Propagation

```
Revoke Request →
  ├── Update PostgreSQL (revoked_at = NOW())
  ├── PUBLISH apikey:revoked {hash, customer_id}
  └── All Kong instances subscribe → Update local cache
→ Effective in <1 second globally
```

**Constitution Alignment**:
- Principle I (Security): SHA-256 hashing, never store plain keys
- Principle I (Revocation): Redis pub/sub for <1s propagation

---

## 11. Multi-Currency Implementation

### Decision: THB primary, USD/SGD secondary

| Currency | Use Case | Exchange Rate Source |
|----------|----------|---------------------|
| THB | Domestic customers | Base currency |
| USD | International customers | Bank of Thailand API (daily) |
| SGD | Singapore customers | Bank of Thailand API (daily) |

**Implementation**:
- Store prices in THB as base currency
- Convert at invoice generation time using locked exchange rate
- Display customer-preferred currency in portal
- All internal calculations in THB for consistency

---

## 12. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Redis Cluster failure | Low | Critical | 6-node cluster, automatic failover, PostgreSQL fallback for counters |
| Billing calculation error | Low | Critical | TDD with 95% coverage, reconciliation jobs, double-entry audit |
| Kong plugin performance | Medium | High | Benchmark testing, circuit breakers, fallback to basic auth |
| PDF generation bottleneck | Medium | Medium | Async generation, queue-based processing, pre-generation for large batches |
| Thai font rendering issues | Low | Medium | Embedded TTF fonts, fallback to English template |

---

## 13. Research Conclusions

### Validated Technology Choices

All technologies from the PRD have been validated against:
- ✅ Constitution principles (8/8 principles satisfied)
- ✅ Performance requirements (budgets achievable)
- ✅ Security requirements (zero-trust, encryption, audit trails)
- ✅ Scalability requirements (horizontal scaling, no SPOF)

### Recommended Adjustments

1. **None required**: PRD technology choices are sound and well-justified
2. **Enhancement**: Consider adding circuit breakers (go-resiliency) for inter-service communication
3. **Enhancement**: Consider adding OpenAPI schema validation middleware for contract testing

### Next Steps

1. Proceed to Phase 1: Data Model design
2. Generate OpenAPI contracts for all services
3. Create development environment quickstart guide

---

**Research Complete** | Ready for Phase 1 Design
