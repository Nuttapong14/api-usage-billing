# Implementation Plan: API Usage Analytics & Billing Platform

**Branch**: `001-api-usage-billing` | **Date**: 2025-12-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-api-usage-billing/spec.md`

## Summary

Build a multi-tenant platform for real-time API usage tracking and automated billing. The system enables API providers to monetize their APIs through configurable subscription tiers, quota enforcement, and automatic invoice generation. Key capabilities include sub-second usage tracking via Redis Streams, Kong Gateway integration for rate limiting, and customer self-service portals for usage visibility.

**Technical Approach**: Microservices architecture in Go with event-driven usage aggregation, Next.js frontends for both admin and customer portals, PostgreSQL for durable storage with Redis for real-time counters, and comprehensive observability via OpenTelemetry.

## Technical Context

**Language/Version**: Go 1.22+ (backend services), TypeScript 5.x (frontend)
**Primary Dependencies**:
- Backend: Fiber v2 (web framework), GORM (ORM), Asynq (background jobs)
- Frontend: Next.js 14, shadcn/ui, Tailwind CSS, Recharts/Apache ECharts
- Infrastructure: Kong Gateway (OSS), Keycloak (auth), Redis Streams (events)

**Storage**:
- PostgreSQL 16 (primary + read replicas) - ACID transactions, partitioned tables
- Redis Cluster (6+ nodes) - real-time counters, streams, pub/sub
- MinIO (S3-compatible) - PDF invoices, reports

**Testing**:
- Go: `go test` with testify, gomock for mocks
- Frontend: Jest + React Testing Library, Playwright for E2E
- Contract: OpenAPI schema validation
- Coverage targets: ≥95% billing services, ≥80% general code

**Target Platform**: Kubernetes cluster (self-hosted or cloud), Linux containers

**Project Type**: Web application (backend microservices + dual frontend portals)

**Performance Goals**:
- Gateway latency (added): <10ms
- Real-time counter update: <100ms P99
- Dashboard query response: <200ms
- Event processing throughput: 50,000 events/sec
- Invoice generation (10K customers): <5 minutes

**Constraints**:
- Billing accuracy: 99.99% (max 1 error per 10,000 transactions)
- System uptime: 99.9% (core functions)
- API key revocation: <1 second global propagation
- Quota alert delivery: <1 minute from threshold breach

**Scale/Scope**:
- 10,000+ customers per organization
- 50,000+ requests/second peak throughput
- 12 months data retention (active), then archive
- Multi-tenant isolation (data per organization)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence/Notes |
|-----------|--------|----------------|
| **I. Security-First Design** | ✅ PASS | SHA-256 API key hashing, TLS 1.3, AES-256 at rest, <1s key revocation via Redis pub/sub per PRD spec |
| **II. Test-Driven Accuracy** | ✅ PASS | TDD enforced for billing services, ≥95% coverage requirement, contract tests for all APIs |
| **III. Performance Budgets** | ✅ PASS | All budgets defined: <10ms gateway, <100ms P99 counters, <200ms dashboard, 50K events/sec |
| **IV. Consistent User Experience** | ✅ PASS | Unified shadcn/ui design system, Thai/English localization, <5s dashboard refresh |
| **V. Observability by Default** | ✅ PASS | OpenTelemetry, Prometheus, Grafana, Loki, Jaeger stack per PRD architecture |
| **VI. Data Integrity** | ✅ PASS | PostgreSQL ACID, Redis Streams exactly-once ACK, immutable audit logs |
| **VII. API Contract Stability** | ✅ PASS | Semantic versioning, 90-day deprecation, 24h key rotation grace period |
| **VIII. Scalable Architecture** | ✅ PASS | Kubernetes HPA, Redis Cluster (6 nodes), PostgreSQL read replicas, consumer groups |

**Gate Status**: ✅ ALL PRINCIPLES SATISFIED - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```text
specs/001-api-usage-billing/
├── plan.md              # This file
├── research.md          # Phase 0: Technology decisions and alternatives
├── data-model.md        # Phase 1: Database schema and entity relationships
├── quickstart.md        # Phase 1: Development environment setup
├── contracts/           # Phase 1: OpenAPI specifications
│   ├── auth-api.yaml
│   ├── usage-api.yaml
│   ├── billing-api.yaml
│   ├── subscription-api.yaml
│   └── webhook-api.yaml
└── tasks.md             # Phase 2: Implementation tasks (via /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── api-gateway/          # Kong plugin configurations
│   ├── auth-service/         # Authentication service entrypoint
│   ├── usage-service/        # Usage tracking service entrypoint
│   ├── billing-service/      # Billing and invoice service entrypoint
│   ├── subscription-service/ # Subscription management entrypoint
│   ├── analytics-service/    # Analytics aggregation entrypoint
│   ├── notification-service/ # Email, webhook, SMS notifications
│   └── worker/               # Background job workers (Asynq)
├── internal/
│   ├── config/               # Configuration loading
│   ├── domain/               # Domain models and business logic
│   │   ├── organization/
│   │   ├── customer/
│   │   ├── subscription/
│   │   ├── usage/
│   │   ├── billing/
│   │   ├── invoice/
│   │   ├── apikey/
│   │   └── webhook/
│   ├── repository/           # Data access layer (GORM)
│   ├── service/              # Business service implementations
│   ├── handler/              # HTTP handlers (Fiber)
│   ├── middleware/           # Auth, logging, rate limiting
│   ├── event/                # Redis Streams event handling
│   └── pkg/                  # Shared utilities
│       ├── auth/             # JWT/API key validation
│       ├── hash/             # SHA-256 hashing
│       ├── quota/            # Quota calculation
│       ├── currency/         # Multi-currency support
│       └── pdf/              # Invoice PDF generation
├── migrations/               # PostgreSQL migrations
├── tests/
│   ├── unit/
│   ├── integration/
│   └── contract/
├── go.mod
└── go.sum

frontend/
├── apps/
│   ├── admin/                # Provider admin dashboard (Next.js)
│   │   ├── src/
│   │   │   ├── app/          # App router pages
│   │   │   ├── components/   # Admin-specific components
│   │   │   ├── hooks/        # Custom React hooks
│   │   │   ├── lib/          # Utilities and API clients
│   │   │   └── types/        # TypeScript definitions
│   │   └── tests/
│   └── customer/             # Customer portal (Next.js)
│       ├── src/
│       │   ├── app/          # App router pages
│       │   ├── components/   # Customer-specific components
│       │   ├── hooks/
│       │   ├── lib/
│       │   └── types/
│       └── tests/
├── packages/
│   ├── ui/                   # Shared shadcn/ui components
│   ├── api-client/           # Generated API client
│   └── i18n/                 # Thai/English translations
├── package.json
└── turbo.json                # Turborepo configuration

infrastructure/
├── kubernetes/
│   ├── base/                 # Base Kustomize manifests
│   │   ├── namespaces/
│   │   ├── configmaps/
│   │   ├── secrets/
│   │   ├── deployments/
│   │   ├── services/
│   │   ├── cronjobs/
│   │   └── network-policies/
│   ├── overlays/
│   │   ├── development/
│   │   ├── staging/
│   │   └── production/
│   └── kong/                 # Kong Gateway configuration
│       ├── plugins/
│       └── routes/
├── terraform/
│   ├── modules/
│   │   ├── kubernetes/
│   │   ├── postgres/
│   │   ├── redis/
│   │   └── minio/
│   └── environments/
│       ├── dev/
│       ├── staging/
│       └── prod/
├── docker/
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   └── docker-compose.dev.yml
└── observability/
    ├── prometheus/
    ├── grafana/
    │   └── dashboards/
    ├── loki/
    └── jaeger/

.github/
├── workflows/
│   ├── ci.yml               # Build, test, lint
│   ├── cd-staging.yml       # Deploy to staging
│   ├── cd-production.yml    # Deploy to production (manual approval)
│   └── security-scan.yml    # Trivy, Snyk scanning
└── CODEOWNERS
```

**Structure Decision**: Option 2 (Web Application) with microservices backend. The platform requires:
1. **Backend microservices**: 7 Go services for domain separation and independent scaling
2. **Dual frontends**: Separate admin dashboard and customer portal for distinct user experiences
3. **Shared packages**: UI components and API client for consistency
4. **Infrastructure as Code**: Kubernetes manifests with Kustomize overlays, Terraform for cloud resources

## Complexity Tracking

> No constitution violations. All complexity is justified by explicit PRD requirements.

| Pattern | Justification | Alternative Considered |
|---------|---------------|------------------------|
| 7 microservices | PRD specifies distinct scaling needs (usage at 50K/s vs billing monthly batch) | Monolith rejected: cannot scale usage tracking independently |
| Redis Cluster (6 nodes) | Constitution requires scalable architecture, no SPOF | Single Redis rejected: violates SPOF principle |
| Dual frontends | PRD specifies distinct Admin Dashboard and Customer Portal | Single app rejected: different auth flows, feature sets, and user journeys |
| GORM over raw SQL | Development velocity for 20+ tables with relationships | Raw SQL rejected: migration complexity, less maintainable |

---

## Phase 0: Research Summary

*To be generated in `research.md`*

## Phase 1: Design Artifacts

*To be generated:*
- `data-model.md` - Entity relationships and PostgreSQL schema
- `contracts/` - OpenAPI specifications for all service APIs
- `quickstart.md` - Local development environment setup
