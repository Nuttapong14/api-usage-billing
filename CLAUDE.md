# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

API Usage Analytics & Billing Platform - A multi-tenant system for API providers to track usage, manage subscriptions, and automate billing. Built with Go microservices backend and Next.js frontend monorepo.

## Development Commands

### Infrastructure (Docker-based local services)

```bash
make infra-up          # Start PostgreSQL, Redis, Keycloak, Kong, MinIO
make infra-down        # Stop and remove containers with volumes
make infra-logs        # Tail logs from all services
```

### Database

```bash
make db-migrate        # Run all pending migrations
make db-migrate-down   # Rollback last migration
make db-reset          # Drop all and re-migrate

# Create new migration
cd backend && migrate create -ext sql -dir migrations -seq migration_name
```

### Backend (Go)

```bash
make backend-dev       # Run with Air hot-reload
make backend-build     # Compile all packages
make backend-test      # Run all tests
make backend-lint      # Run golangci-lint

# Run single test
cd backend && go test -v -run TestFunctionName ./path/to/package

# Run specific service
BACKEND_SERVICE=billing-service make backend-dev
```

### Frontend (Next.js 14 + bun + Turborepo)

```bash
make frontend-dev      # Start dev servers (admin:3000 + customer:3001)
make frontend-build    # Build all apps
make frontend-test     # Run tests
make frontend-lint     # ESLint + Prettier check

# Run from frontend directory
cd frontend && bun dev                      # Start all apps via Turborepo
cd frontend && bun run --filter @repo/admin dev   # Single app only

# Type checking
cd frontend && bun run type-check

# Clean and reinstall
cd frontend && bun run clean && bun install
```

### Full Stack

```bash
make dev               # Start infra, then run backend/frontend in separate terminals
make test              # Backend + frontend tests
make lint              # Backend + frontend linting
make build             # Build all
```

## Architecture

### Backend Services (Go/Fiber)

```
backend/
├── cmd/                          # Service entry points
│   ├── usage-service/            # Real-time usage tracking API
│   ├── subscription-service/     # Tier and quota management
│   ├── billing-service/          # Invoice generation and payments
│   ├── analytics-service/        # Provider dashboards
│   ├── notification-service/     # Alerts and webhooks
│   ├── api-gateway/              # Kong plugin for usage logging
│   └── worker/jobs/              # Background jobs (aggregation, billing cycle)
├── internal/
│   ├── domain/                   # Domain models (organization, customer, billing, usage)
│   ├── repository/               # Database access layer (GORM)
│   ├── service/                  # Business logic (usage, subscription, billing)
│   ├── handler/                  # HTTP handlers per domain
│   ├── middleware/               # Auth, tenant isolation, telemetry
│   ├── infrastructure/           # Redis client, events publisher/consumer
│   ├── event/                    # Domain events (usage events)
│   ├── api/                      # Shared API utilities (validation, pagination, errors)
│   └── pkg/                      # Shared packages (auth, currency, pdf, quota, storage)
└── migrations/                   # PostgreSQL migrations (golang-migrate format)
```

### Frontend Monorepo (bun + Turborepo)

```
frontend/
├── apps/
│   ├── admin/                    # Provider dashboard (Next.js 14)
│   └── customer/                 # Customer self-service portal (Next.js 14)
└── packages/
    ├── ui/                       # Shared shadcn/ui components
    ├── i18n/                     # Thai/English translations
    ├── api-client/               # Generated OpenAPI TypeScript client
    ├── auth/                     # Keycloak authentication
    └── query/                    # TanStack Query hooks
```

### Infrastructure

```
infrastructure/
├── docker/                       # Local dev compose + Kong config
├── kubernetes/                   # K8s manifests for Kong plugins
└── observability/                # Prometheus alerts, Grafana dashboards
```

### Key Data Flow

1. **Usage Tracking**: API Request → Kong Gateway → Usage Plugin → Redis Streams → Worker (aggregator) → PostgreSQL + Redis counters
2. **Quota Enforcement**: Kong authenticates API key (hash lookup) → checks tier limits → enforces rate limit → logs usage event
3. **Billing Cycle**: CronJob → collect usage from Redis/Postgres → calculate charges → generate invoice PDF → store in MinIO → notify customer

### Multi-Tenancy

- `Organization` is the root tenant (API Provider)
- All queries scoped by `organization_id` via middleware
- Tenant context extracted from JWT claims

## API Contracts

OpenAPI specs in `specs/001-api-usage-billing/contracts/`:

- `usage-api.yaml` - Real-time usage queries
- `subscription-api.yaml` - Tier and quota management
- `billing-api.yaml` - Invoices and payments
- `apikey-api.yaml` - API key lifecycle
- `webhook-api.yaml` - Customer webhook endpoints

Regenerate TypeScript client:

```bash
cd frontend/packages/api-client && bun run generate:openapi
```

## Key Technologies

- **Backend**: Go 1.22+, Fiber v2, GORM, Redis Streams, PostgreSQL 16
- **Frontend**: Next.js 14, bun, Turborepo, shadcn/ui, Tailwind CSS
- **Infra**: Kong Gateway (API gateway + rate limiting), Keycloak (auth), MinIO (S3-compatible storage)
- **Observability**: OpenTelemetry, Prometheus, Grafana, Loki

## Testing Patterns

- Backend: Standard Go testing, use `-short` flag for CI
- Frontend: Vitest/Jest for unit tests
- Integration: Use `make infra-up` to start dependencies before running integration tests

## Environment Variables

Copy `.env.example` to `.env` and configure before running:

```bash
cp .env.example .env
# Edit .env with your credentials
```

Required for local development:

- `DB_PASSWORD` - PostgreSQL password (REQUIRED, no default)
- `KEYCLOAK_ADMIN_PASSWORD` - Keycloak admin password (REQUIRED)
- `KEYCLOAK_DB_PASSWORD` - Keycloak database password (REQUIRED)
- `MINIO_ROOT_PASSWORD` - MinIO root password (REQUIRED)

Optional with sensible defaults:

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` - Database connection settings
- `REDIS_HOST`, `REDIS_PORT` - Redis connection settings
- `KEYCLOAK_HOST`, `KEYCLOAK_PORT` - Keycloak connection settings

See `.env.example` for complete list of configuration options.
