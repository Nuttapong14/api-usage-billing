# Tasks: API Usage Analytics & Billing Platform

**Input**: Design documents from `/specs/001-api-usage-billing/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅
**Generated**: 2025-12-26 | **Total Tasks**: 127

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7)
- All paths relative to repository root

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize project structure, dependencies, and development infrastructure

### Backend Setup

- [x] T001 Create Go workspace and module structure in `backend/go.mod` with Go 1.22+
- [x] T002 [P] Initialize backend service directories per plan.md structure in `backend/cmd/`
- [x] T003 [P] Configure golangci-lint with strict settings in `backend/.golangci.yml`
- [x] T004 [P] Create shared types package in `backend/internal/pkg/types/types.go`
- [x] T005 [P] Create error definitions package in `backend/internal/pkg/errors/errors.go`
- [x] T006 [P] Setup configuration loader with Viper in `backend/internal/config/config.go`
- [x] T007 [P] Create logger wrapper (zerolog) in `backend/internal/pkg/logger/logger.go`

### Frontend Setup

- [x] T008 Initialize Turborepo monorepo in `frontend/package.json` with bun workspace
- [x] T009 [P] Create admin app scaffold (Next.js 14) in `frontend/apps/admin/`
- [x] T010 [P] Create customer app scaffold (Next.js 14) in `frontend/apps/customer/`
- [x] T011 [P] Initialize shared UI package (shadcn/ui) in `frontend/packages/ui/`
- [x] T012 [P] Create API client package in `frontend/packages/api-client/`
- [x] T013 [P] Setup i18n package (Thai/English) in `frontend/packages/i18n/`
- [x] T014 [P] Configure ESLint and Prettier in `frontend/.eslintrc.js` and `.prettierrc`

### Infrastructure Setup

- [x] T015 Create docker-compose.dev.yml in `infrastructure/docker/` per quickstart.md
- [x] T016 [P] Create Kong Gateway configuration in `infrastructure/kubernetes/kong/`
- [x] T017 [P] Create Keycloak realm configuration in `infrastructure/docker/keycloak/`
- [x] T018 [P] Setup Prometheus/Grafana configs in `infrastructure/observability/`
- [x] T019 Create Makefile with all commands from quickstart.md in root `Makefile`

**Checkpoint**: Project structure ready - `make infra-up` should start all dependencies

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Foundation

- [x] T020 Create PostgreSQL migration framework in `backend/migrations/`
- [x] T021 Create `001_create_organizations.sql` migration in `backend/migrations/`
- [x] T022 [P] Create `002_create_customers.sql` migration in `backend/migrations/`
- [x] T023 [P] Create `003_create_subscription_tiers.sql` migration in `backend/migrations/`
- [x] T024 Create Organization GORM model in `backend/internal/domain/organization/model.go`
- [x] T025 [P] Create Customer GORM model in `backend/internal/domain/customer/model.go`
- [x] T026 [P] Create SubscriptionTier GORM model in `backend/internal/domain/subscription/tier_model.go`
- [x] T027 Create base repository interface in `backend/internal/repository/base.go`
- [x] T028 [P] Create Organization repository in `backend/internal/repository/organization.go`
- [x] T029 [P] Create Customer repository in `backend/internal/repository/customer.go`

### Authentication & Authorization

- [x] T030 Implement JWT middleware (Keycloak) in `backend/internal/middleware/auth.go`
- [x] T031 [P] Implement API key validator in `backend/internal/middleware/apikey.go`
- [x] T032 [P] Create auth context helper in `backend/internal/pkg/auth/context.go`
- [x] T033 Implement multi-tenant middleware in `backend/internal/middleware/tenant.go`
- [x] T034 Create permission checker in `backend/internal/pkg/auth/permissions.go`

### Event Infrastructure

- [x] T035 Create Redis Streams client wrapper in `backend/internal/infrastructure/redis/client.go`
- [x] T036 [P] Create Redis pub/sub wrapper in `backend/internal/infrastructure/redis/pubsub.go`
- [x] T037 [P] Create Redis counter helper in `backend/internal/infrastructure/redis/counter.go`
- [x] T038 Create event publisher interface in `backend/internal/infrastructure/events/publisher.go`
- [x] T039 Create event consumer base in `backend/internal/infrastructure/events/consumer.go`

### API Framework

- [x] T040 Create Fiber app factory in `backend/internal/api/server.go`
- [x] T041 [P] Create request validation helper in `backend/internal/api/validation.go`
- [x] T042 [P] Create response builder in `backend/internal/api/response.go`
- [x] T043 [P] Create pagination helper in `backend/internal/api/pagination.go`
- [x] T044 Create error handler middleware in `backend/internal/api/errors.go`
- [x] T045 Create request ID middleware in `backend/internal/middleware/requestid.go`
- [x] T046 Create OpenTelemetry middleware in `backend/internal/middleware/telemetry.go`

### Shared Frontend Foundation

- [x] T047 Create auth provider (Keycloak) in `frontend/packages/auth/src/AuthProvider.tsx`
- [x] T048 [P] Create API client base in `frontend/packages/api-client/src/client.ts`
- [x] T049 [P] Create error boundary component in `frontend/packages/ui/src/ErrorBoundary.tsx`
- [x] T050 [P] Create loading spinner in `frontend/packages/ui/src/Loading.tsx`
- [x] T051 Create theme provider in `frontend/packages/ui/src/ThemeProvider.tsx`
- [x] T052 Setup React Query provider in `frontend/packages/query/src/provider.tsx`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Real-time Usage Tracking (Priority: P1) 🎯 MVP

**Goal**: Track every API request in real-time with <5s counter updates

**Independent Test**: Make API calls through gateway, verify counters update within 5 seconds

### Database Layer (US1)

- [x] T053 [P] [US1] Create `004_create_subscriptions.sql` migration in `backend/migrations/`
- [x] T054 [P] [US1] Create `006_create_usage_records_partitioned.sql` in `backend/migrations/`
- [x] T055 [P] [US1] Create `007_create_usage_daily.sql` migration in `backend/migrations/`
- [x] T056 [US1] Create Subscription GORM model in `backend/internal/domain/subscription/model.go`
- [x] T057 [P] [US1] Create UsageRecord GORM model in `backend/internal/domain/usage/record_model.go`
- [x] T058 [P] [US1] Create UsageDaily GORM model in `backend/internal/domain/usage/daily_model.go`

### Event Processing (US1)

- [x] T059 [US1] Implement usage event schema in `backend/internal/event/usage/event.go`
- [x] T060 [US1] Create usage event publisher in `backend/internal/event/usage/publisher.go`
- [x] T061 [US1] Create usage event consumer in `backend/internal/event/usage/consumer.go`
- [x] T062 [US1] Implement batch insert for usage records in `backend/internal/repository/usage.go`
- [x] T063 [US1] Create daily aggregation job in `backend/cmd/worker/jobs/aggregate_usage.go`

### Service Layer (US1)

- [x] T064 [US1] Create usage service interface in `backend/internal/service/usage/interface.go`
- [x] T065 [US1] Implement usage service in `backend/internal/service/usage/service.go`
- [x] T066 [US1] Implement real-time counter sync in `backend/internal/service/usage/counter.go`
- [x] T067 [US1] Create usage repository in `backend/internal/repository/usage.go`

### API Handlers (US1)

- [x] T068 [US1] Implement GET /usage/current handler in `backend/internal/handler/usage/current.go`
- [x] T069 [P] [US1] Implement GET /usage/history handler in `backend/internal/handler/usage/history.go`
- [x] T070 [P] [US1] Implement GET /usage/breakdown handler in `backend/internal/handler/usage/breakdown.go`
- [x] T071 [US1] Create usage service main entrypoint in `backend/cmd/usage-service/main.go`
- [x] T072 [US1] Configure usage service routes in `backend/cmd/usage-service/routes.go`

### Kong Integration (US1)

- [x] T073 [US1] Create Kong logging plugin config in `infrastructure/kubernetes/kong/plugins/usage-log.yaml`
- [x] T074 [US1] Implement Kong plugin handler in `backend/cmd/api-gateway/usage_plugin.go`

**Checkpoint**: User Story 1 complete - usage tracking functional with <5s updates

---

## Phase 4: User Story 2 - Subscription & Quota Management (Priority: P1)

**Goal**: Define tiers with quotas/rate limits, enforce limits, send quota alerts

**Independent Test**: Create tier, assign customer, verify quota enforcement at 80%/95%/100%

### Service Layer (US2)

- [x] T075 [US2] Create subscription service interface in `backend/internal/service/subscription/interface.go`
- [x] T076 [US2] Implement subscription service in `backend/internal/service/subscription/service.go`
- [x] T077 [US2] Create tier service in `backend/internal/service/subscription/tier_service.go`
- [x] T078 [US2] Implement quota calculator in `backend/internal/pkg/quota/calculator.go`
- [x] T079 [US2] Create quota alert checker in `backend/internal/service/subscription/quota_alert.go`
- [x] T080 [US2] Create subscription repository in `backend/internal/repository/subscription.go`
- [x] T081 [P] [US2] Create tier repository in `backend/internal/repository/tier.go`

### API Handlers (US2)

- [x] T082 [US2] Implement GET /subscription handler in `backend/internal/handler/subscription/get.go`
- [x] T083 [P] [US2] Implement GET /subscription/tiers handler in `backend/internal/handler/subscription/tiers.go`
- [x] T084 [P] [US2] Implement POST /subscription/upgrade handler in `backend/internal/handler/subscription/upgrade.go`
- [x] T085 [P] [US2] Implement POST /subscription/downgrade handler in `backend/internal/handler/subscription/downgrade.go`
- [x] T086 [US2] Create subscription service main in `backend/cmd/subscription-service/main.go`
- [x] T087 [US2] Configure subscription service routes in `backend/cmd/subscription-service/routes.go`

### Rate Limiting (US2)

- [x] T088 [US2] Implement rate limit middleware in `backend/internal/middleware/ratelimit.go`
- [x] T089 [US2] Create Kong rate limit plugin config in `infrastructure/kubernetes/kong/plugins/rate-limit.yaml`
- [x] T090 [US2] Implement tier-based rate limit resolver in `backend/internal/service/subscription/ratelimit.go`

**Checkpoint**: User Story 2 complete - subscriptions and quota enforcement working

---

## Phase 5: User Story 3 - Automatic Invoice Generation (Priority: P1)

**Goal**: Auto-generate monthly invoices with accurate calculations and PDF generation

**Independent Test**: Simulate billing cycle close, verify invoice accuracy against usage

### Database Layer (US3)

- [x] T091 [P] [US3] Create `008_create_invoices.sql` migration in `backend/migrations/`
- [x] T092 [P] [US3] Create `009_create_payments.sql` migration in `backend/migrations/`
- [x] T093 [P] [US3] Create `013_create_credit_notes.sql` migration in `backend/migrations/`
- [x] T094 [US3] Create Invoice GORM model in `backend/internal/domain/billing/invoice_model.go`
- [x] T095 [P] [US3] Create Payment GORM model in `backend/internal/domain/billing/payment_model.go`
- [x] T096 [P] [US3] Create CreditNote GORM model in `backend/internal/domain/billing/credit_note_model.go`

### Service Layer (US3)

- [x] T097 [US3] Create billing service interface in `backend/internal/service/billing/interface.go`
- [x] T098 [US3] Implement billing service in `backend/internal/service/billing/service.go`
- [x] T099 [US3] Implement invoice calculator in `backend/internal/service/billing/calculator.go`
- [x] T100 [US3] Implement overage calculator in `backend/internal/service/billing/overage.go`
- [x] T101 [US3] Create VAT calculator (7%) in `backend/internal/pkg/currency/vat.go`
- [x] T102 [US3] Create invoice repository in `backend/internal/repository/invoice.go`
- [x] T103 [P] [US3] Create payment repository in `backend/internal/repository/payment.go`

### PDF Generation (US3)

- [x] T104 [US3] Create PDF generator interface in `backend/internal/pkg/pdf/interface.go`
- [x] T105 [US3] Implement invoice PDF generator in `backend/internal/pkg/pdf/invoice.go`
- [x] T106 [US3] Create Thai/English invoice templates in `backend/internal/pkg/pdf/templates/`
- [x] T107 [US3] Implement MinIO uploader in `backend/internal/pkg/storage/minio.go`

### Background Jobs (US3)

- [x] T108 [US3] Create billing cycle job in `backend/cmd/worker/jobs/billing_cycle.go`
- [x] T109 [US3] Create invoice generation job in `backend/cmd/worker/jobs/generate_invoices.go`
- [x] T110 [US3] Create payment reminder job in `backend/cmd/worker/jobs/payment_reminder.go`

### API Handlers (US3)

- [x] T111 [US3] Implement GET /invoices handler in `backend/internal/handler/billing/invoices.go`
- [x] T112 [P] [US3] Implement GET /invoices/{id}/pdf handler in `backend/internal/handler/billing/invoice_pdf.go`
- [x] T113 [P] [US3] Implement POST /invoices/{id}/pay handler in `backend/internal/handler/billing/pay.go`
- [x] T114 [US3] Create billing service main in `backend/cmd/billing-service/main.go`

**Checkpoint**: User Story 3 complete - automatic invoicing with PDF generation working

---

## Phase 6: User Story 4 - Customer Usage Dashboard (Priority: P2)

**Goal**: Self-service dashboard for usage monitoring, quota status, and invoice access

**Independent Test**: Login as customer, verify dashboard shows real-time data

### Customer Portal Frontend (US4)

- [x] T115 [US4] Create dashboard layout in `frontend/apps/customer/src/app/(dashboard)/layout.tsx`
- [x] T116 [US4] Implement usage overview page in `frontend/apps/customer/src/app/(dashboard)/page.tsx`
- [x] T117 [P] [US4] Create usage chart component in `frontend/apps/customer/src/components/usage-chart.tsx`
- [x] T118 [P] [US4] Create quota progress component in `frontend/apps/customer/src/components/quota-progress.tsx`
- [x] T119 [P] [US4] Create billing period card in `frontend/apps/customer/src/components/billing-period-card.tsx`
- [x] T120 [US4] Implement billing history page in `frontend/apps/customer/src/app/(dashboard)/billing/page.tsx`
- [x] T121 [P] [US4] Create invoice list component in `frontend/apps/customer/src/components/invoice-list.tsx`
- [x] T122 [P] [US4] Create invoice download button in `frontend/apps/customer/src/components/invoice-download.tsx`
- [x] T123 [US4] Create usage hooks in `frontend/apps/customer/src/hooks/use-usage.ts`
- [x] T124 [P] [US4] Create billing hooks in `frontend/apps/customer/src/hooks/use-billing.ts`

**Checkpoint**: User Story 4 complete - customer self-service dashboard functional

---

## Phase 7: User Story 5 - API Key Management (Priority: P2)

**Goal**: Secure API key lifecycle with rotation and instant revocation

**Independent Test**: Create key, rotate it, verify 24h grace period, revoke and verify <1s propagation

### Database Layer (US5)

- [x] T125 [US5] Create `005_create_api_keys.sql` migration in `backend/migrations/`
- [x] T126 [US5] Create APIKey GORM model in `backend/internal/domain/apikey/model.go`

### Service Layer (US5)

- [x] T127 [US5] Implement SHA-256 key hasher in `backend/internal/pkg/hash/apikey.go`
- [x] T128 [US5] Create API key service interface in `backend/internal/service/apikey/interface.go`
- [x] T129 [US5] Implement API key service in `backend/internal/service/apikey/service.go`
- [x] T130 [US5] Implement key rotation logic in `backend/internal/service/apikey/rotation.go`
- [x] T131 [US5] Implement instant revocation via pub/sub in `backend/internal/service/apikey/revocation.go`
- [x] T132 [US5] Create API key repository in `backend/internal/repository/apikey.go`
- [x] T133 [US5] Create API key cache in `backend/internal/service/apikey/cache.go`

### API Handlers (US5)

- [x] T134 [US5] Implement GET /api-keys handler in `backend/internal/handler/apikey/list.go`
- [x] T135 [P] [US5] Implement POST /api-keys handler in `backend/internal/handler/apikey/create.go`
- [x] T136 [P] [US5] Implement POST /api-keys/{id}/rotate handler in `backend/internal/handler/apikey/rotate.go`
- [x] T137 [P] [US5] Implement DELETE /api-keys/{id} handler in `backend/internal/handler/apikey/revoke.go`
- [x] T138 [US5] Configure API key routes in `backend/cmd/usage-service/routes.go`

### Customer Portal (US5)

- [x] T139 [US5] Create API keys page in `frontend/apps/customer/src/app/(dashboard)/api-keys/page.tsx`
- [x] T140 [P] [US5] Create API key list component in `frontend/apps/customer/src/components/api-key-list.tsx`
- [x] T141 [P] [US5] Create key creation modal in `frontend/apps/customer/src/components/create-key-modal.tsx`
- [x] T142 [P] [US5] Create key rotation dialog in `frontend/apps/customer/src/components/rotate-key-dialog.tsx`

**Checkpoint**: User Story 5 complete - secure API key management with instant revocation

---

## Phase 8: User Story 6 - Provider Analytics Dashboard (Priority: P3)

**Goal**: Business analytics for revenue trends, top customers, and API performance

**Independent Test**: Generate sample data, verify analytics reflect accurate aggregations

### Database Layer (US6)

- [x] T143 [US6] Create `012_create_audit_logs_partitioned.sql` migration in `backend/migrations/`
- [x] T144 [US6] Create AuditLog GORM model in `backend/internal/domain/audit/model.go`

### Service Layer (US6)

- [x] T145 [US6] Create analytics service interface in `backend/internal/service/analytics/interface.go`
- [x] T146 [US6] Implement analytics service in `backend/internal/service/analytics/service.go`
- [x] T147 [US6] Create revenue aggregator in `backend/internal/service/analytics/revenue.go`
- [x] T148 [P] [US6] Create customer ranking in `backend/internal/service/analytics/ranking.go`
- [x] T149 [P] [US6] Create API endpoint stats in `backend/internal/service/analytics/endpoints.go`

### API Handlers (US6)

- [x] T150 [US6] Implement GET /analytics/overview handler in `backend/internal/handler/analytics/overview.go`
- [x] T151 [P] [US6] Implement GET /analytics/revenue handler in `backend/internal/handler/analytics/revenue.go`
- [x] T152 [P] [US6] Implement GET /analytics/customers handler in `backend/internal/handler/analytics/customers.go`
- [x] T153 [US6] Create analytics service main in `backend/cmd/analytics-service/main.go`

### Admin Dashboard Frontend (US6)

- [x] T154 [US6] Create admin dashboard layout in `frontend/apps/admin/src/app/(dashboard)/layout.tsx`
- [x] T155 [US6] Implement overview page in `frontend/apps/admin/src/app/(dashboard)/page.tsx`
- [x] T156 [P] [US6] Create revenue chart component in `frontend/apps/admin/src/components/revenue-chart.tsx`
- [x] T157 [P] [US6] Create top customers table in `frontend/apps/admin/src/components/top-customers.tsx`
- [x] T158 [P] [US6] Create top APIs chart in `frontend/apps/admin/src/components/top-apis.tsx`
- [x] T159 [US6] Implement customer management page in `frontend/apps/admin/src/app/(dashboard)/customers/page.tsx`
- [x] T160 [P] [US6] Implement tier management page in `frontend/apps/admin/src/app/(dashboard)/tiers/page.tsx`

**Checkpoint**: User Story 6 complete - provider analytics dashboard functional

---

## Phase 9: User Story 7 - Alerting & Notifications (Priority: P3)

**Goal**: Proactive notifications for quota warnings, payments, and system events

**Independent Test**: Trigger quota threshold, verify email/webhook delivery within 1 minute

### Database Layer (US7)

- [x] T161 [P] [US7] Create `010_create_webhook_endpoints.sql` migration in `backend/migrations/`
- [x] T162 [P] [US7] Create `011_create_webhook_deliveries.sql` migration in `backend/migrations/`
- [x] T163 [US7] Create WebhookEndpoint GORM model in `backend/internal/domain/webhook/endpoint_model.go`
- [x] T164 [P] [US7] Create WebhookDelivery GORM model in `backend/internal/domain/webhook/delivery_model.go`

### Service Layer (US7)

- [x] T165 [US7] Create notification service interface in `backend/internal/service/notification/interface.go`
- [x] T166 [US7] Implement notification service in `backend/internal/service/notification/service.go`
- [x] T167 [US7] Create email sender (SMTP) in `backend/internal/pkg/email/sender.go`
- [x] T168 [US7] Create webhook dispatcher in `backend/internal/service/notification/webhook.go`
- [x] T169 [US7] Implement HMAC-SHA256 signing in `backend/internal/pkg/crypto/hmac.go`
- [x] T170 [US7] Create webhook repository in `backend/internal/repository/webhook.go`

### Background Jobs (US7)

- [x] T171 [US7] Create quota alert job in `backend/cmd/worker/jobs/quota_alerts.go`
- [x] T172 [P] [US7] Create webhook delivery job in `backend/cmd/worker/jobs/webhook_delivery.go`
- [x] T173 [P] [US7] Create webhook retry job in `backend/cmd/worker/jobs/webhook_retry.go`

### API Handlers (US7)

- [x] T174 [US7] Implement GET /webhooks handler in `backend/internal/handler/webhook/list.go`
- [x] T175 [P] [US7] Implement POST /webhooks handler in `backend/internal/handler/webhook/create.go`
- [x] T176 [P] [US7] Implement POST /webhooks/{id}/test handler in `backend/internal/handler/webhook/test.go`
- [x] T177 [US7] Create notification service main in `backend/cmd/notification-service/main.go`

### Customer Portal (US7)

- [x] T178 [US7] Create webhooks page in `frontend/apps/customer/src/app/(dashboard)/webhooks/page.tsx`
- [x] T179 [P] [US7] Create webhook form component in `frontend/apps/customer/src/components/webhook-form.tsx`

**Checkpoint**: User Story 7 complete - alerting and webhook notifications working

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Quality improvements affecting multiple user stories

### Testing

- [x] T180 [P] Create unit test helper in `backend/tests/unit/helper.go`
- [x] T181 [P] Create integration test setup in `backend/tests/integration/setup.go`
- [x] T182 [P] Create contract test runner in `backend/tests/contract/runner.go`
- [x] T183 [P] Create frontend E2E test setup in `frontend/playwright.config.ts`

### Security Hardening

- [x] T184 [P] Implement rate limit bypass protection in `backend/internal/middleware/ratelimit.go`
- [x] T185 [P] Add CORS configuration in `backend/internal/middleware/cors.go`
- [x] T186 [P] Implement request sanitization in `backend/internal/middleware/sanitize.go`

### Observability

- [x] T187 [P] Create health check endpoint in `backend/internal/handler/health.go`
- [x] T188 [P] Create Grafana dashboards in `infrastructure/observability/grafana/dashboards/`
- [x] T189 [P] Create alert rules in `infrastructure/observability/prometheus/alerts/`

### CI/CD

- [x] T190 [P] Create GitHub Actions CI workflow in `.github/workflows/ci.yml`
- [x] T191 [P] Create staging deployment workflow in `.github/workflows/cd-staging.yml`
- [x] T192 [P] Create production deployment workflow in `.github/workflows/cd-production.yml`
- [x] T193 [P] Create security scanning workflow in `.github/workflows/security-scan.yml`

### Documentation

- [x] T194 [P] Generate OpenAPI client from contracts in `frontend/packages/api-client/`
- [x] T195 [P] Create API documentation site in `docs/api/`
- [ ] T196 Run quickstart.md validation end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup) ──────────────────────────────────────────────────────────────►
                                                                               │
Phase 2 (Foundational) ◄───────────────────────────────────────────────────────┘
         │
         │ BLOCKS ALL USER STORIES
         ▼
    ┌────┴────┬────────────┬────────────┐
    │         │            │            │
Phase 3    Phase 4     Phase 5      Phase 6/7
(US1-P1)   (US2-P1)    (US3-P1)     (Lower Priority)
    │         │            │            │
    ▼         ▼            ▼            ▼
Phase 4    Phase 5     Phase 6      Phase 7
(US4-P2 depends on US1 for usage data)
    │
    ▼
Phase 10 (Polish) ◄─────────────────────────────────────────── All Stories Done
```

### User Story Dependencies

| Story                    | Priority | Depends On                    | Can Start After    |
| ------------------------ | -------- | ----------------------------- | ------------------ |
| US1 - Real-time Usage    | P1       | Foundational                  | Phase 2 complete   |
| US2 - Subscriptions      | P1       | Foundational                  | Phase 2 complete   |
| US3 - Invoicing          | P1       | US1 (usage data), US2 (tiers) | US1 + US2 complete |
| US4 - Customer Dashboard | P2       | US1 (data to display)         | US1 complete       |
| US5 - API Key Management | P2       | Foundational                  | Phase 2 complete   |
| US6 - Provider Analytics | P3       | US1, US2, US3                 | Core P1 complete   |
| US7 - Notifications      | P3       | US2 (quota alerts)            | US2 complete       |

### Within Each User Story

1. Database migrations first
2. GORM models second
3. Repository layer third
4. Service layer fourth
5. API handlers fifth
6. Frontend components last

---

## Parallel Opportunities

### Setup Phase (All Parallel)

```bash
# Backend setup in parallel:
T002 + T003 + T004 + T005 + T006 + T007

# Frontend setup in parallel:
T009 + T010 + T011 + T012 + T013 + T014

# Infrastructure in parallel:
T016 + T017 + T018
```

### Foundational Phase

```bash
# Database models (after migrations):
T024 + T025 + T026  # All entity models in parallel

# Auth components:
T030 + T031 + T032  # Different middleware files

# Event infrastructure:
T035 + T036 + T037  # Different Redis wrappers
```

### User Story 1 (After Foundational)

```bash
# Database layer:
T053 + T054 + T055  # Different migration files

# API handlers:
T068 + T069 + T070  # Different endpoint files
```

### P1 Stories (US1 + US2 + US3 in Parallel)

```bash
# Team A: User Story 1 (Usage Tracking)
# Team B: User Story 2 (Subscriptions)
# Team C: User Story 3 (Invoicing) - starts after US1 data layer

# Within each team, tasks marked [P] run in parallel
```

### P2/P3 Stories (After P1 Core)

```bash
# US4, US5, US6, US7 can all proceed in parallel
# Each has independent functionality
```

---

## Implementation Strategy

### MVP First (User Stories 1, 2, 3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: US1 - Usage Tracking
4. **VALIDATE**: Test usage tracking independently
5. Complete Phase 4: US2 - Subscriptions
6. Complete Phase 5: US3 - Invoicing
7. **DEPLOY MVP**: Core billing functional

### Incremental Delivery

| Increment     | Stories         | Business Value                                               |
| ------------- | --------------- | ------------------------------------------------------------ |
| MVP           | US1 + US2 + US3 | Core billing: track usage, enforce quotas, generate invoices |
| +Self-Service | US4 + US5       | Customer autonomy: dashboard, API key management             |
| +Analytics    | US6 + US7       | Business intelligence: analytics, proactive alerts           |

### Parallel Team Strategy

With 3 developers after Foundational phase:

- **Developer A**: US1 (Usage) → US4 (Dashboard)
- **Developer B**: US2 (Subscriptions) → US7 (Notifications)
- **Developer C**: US3 (Billing) → US5 (API Keys) → US6 (Analytics)

---

## Summary

| Metric                     | Count                      |
| -------------------------- | -------------------------- |
| **Total Tasks**            | 196                        |
| **Setup Tasks**            | 19                         |
| **Foundational Tasks**     | 33                         |
| **US1 Tasks**              | 22                         |
| **US2 Tasks**              | 16                         |
| **US3 Tasks**              | 24                         |
| **US4 Tasks**              | 10                         |
| **US5 Tasks**              | 18                         |
| **US6 Tasks**              | 18                         |
| **US7 Tasks**              | 19                         |
| **Polish Tasks**           | 17                         |
| **Parallel Opportunities** | 67 tasks (34%)             |
| **MVP Scope**              | US1 + US2 + US3 (62 tasks) |

---

## Notes

- [P] tasks = different files, no dependencies on other tasks in same phase
- [US#] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `make lint && make test` before marking phase complete
