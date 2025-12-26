# Quickstart Guide: API Usage Analytics & Billing Platform

**Feature**: 001-api-usage-billing | **Date**: 2025-12-26 | **Status**: Complete

This guide provides step-by-step instructions for setting up the local development environment.

---

## Prerequisites

### Required Software

| Software       | Version | Purpose                      |
| -------------- | ------- | ---------------------------- |
| Go             | 1.22+   | Backend services             |
| Node.js        | 20 LTS  | Frontend applications        |
| pnpm           | 8.x     | Package management           |
| Docker         | 24+     | Container runtime            |
| Docker Compose | 2.20+   | Local orchestration          |
| kubectl        | 1.28+   | Kubernetes CLI (optional)    |
| k3d            | 5.x     | Local K8S cluster (optional) |
| Make           | 4.x     | Build automation             |

### Verify Installation

```bash
# Check all prerequisites
go version          # go version go1.22+
node --version      # v20.x.x
bun --version      # 8.x.x
docker --version    # Docker version 24+
docker compose version  # Docker Compose version v2.20+
```

---

## Quick Start (5 minutes)

### 1. Clone Repository

```bash
git clone https://github.com/your-org/api-usage-billing.git
cd api-usage-billing
```

### 2. Configure Environment

```bash
# Copy environment template and configure credentials
cp .env.example .env

# Edit .env with your local development credentials
# IMPORTANT: Never commit .env files with real credentials
```

### 3. Start Infrastructure

```bash
# Start all dependencies (PostgreSQL, Redis, Keycloak, Kong, MinIO)
make infra-up

# Wait for services to be healthy (~60 seconds)
make infra-wait
```

### 4. Initialize Database

```bash
# Run migrations
make db-migrate

# Seed development data
make db-seed
```

### 5. Start Backend Services

```bash
# Start all Go services in development mode
make backend-dev
```

### 6. Start Frontend

```bash
# In a new terminal
make frontend-dev
```

### 7. Access Applications

| Application     | URL                   | Credentials                           |
| --------------- | --------------------- | ------------------------------------- |
| Customer Portal | http://localhost:3000 | demo@example.com / demo123            |
| Admin Dashboard | http://localhost:3001 | admin@example.com / admin123          |
| Kong Admin API  | http://localhost:8001 | -                                     |
| Keycloak Admin  | http://localhost:8080 | Set via KEYCLOAK_ADMIN_PASSWORD       |
| Grafana         | http://localhost:3030 | admin / admin                         |
| MinIO Console   | http://localhost:9001 | Set via MINIO_ROOT_USER/PASSWORD      |

---

## Detailed Setup

### Infrastructure Services

#### Docker Compose Services

```yaml
# infrastructure/docker/docker-compose.dev.yml
# NOTE: All credentials are loaded from environment variables (.env file)
# See .env.example for required configuration
services:
  postgres:
    image: postgres:16-alpine
    ports:
      - "${DB_PORT:-5432}:5432"
    environment:
      POSTGRES_DB: ${DB_NAME:-api_billing}
      POSTGRES_USER: ${DB_USER:-billing}
      POSTGRES_PASSWORD: ${DB_PASSWORD:?Required}  # From .env
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "${REDIS_PORT:-6379}:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  keycloak:
    image: quay.io/keycloak/keycloak:23.0
    ports:
      - "${KEYCLOAK_PORT:-8080}:8080"
    environment:
      KEYCLOAK_ADMIN: ${KEYCLOAK_ADMIN:-admin}
      KEYCLOAK_ADMIN_PASSWORD: ${KEYCLOAK_ADMIN_PASSWORD:?Required}  # From .env
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
      KC_DB_USERNAME: ${KEYCLOAK_DB_USER:-keycloak}
      KC_DB_PASSWORD: ${KEYCLOAK_DB_PASSWORD:?Required}  # From .env
    command: start-dev

  kong:
    image: kong:3.5
    ports:
      - "${KONG_PROXY_PORT:-8000}:8000"  # Proxy
      - "${KONG_ADMIN_PORT:-8001}:8001"  # Admin API
      - "${KONG_PROXY_SSL_PORT:-8443}:8443"  # Proxy SSL
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /etc/kong/kong.yml
      # ... logging configuration ...
    volumes:
      - ./kong/kong.yml:/etc/kong/kong.yml

  minio:
    image: minio/minio:latest
    ports:
      - "${MINIO_PORT:-9000}:9000"
      - "${MINIO_CONSOLE_PORT:-9001}:9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:?Required}  # From .env
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
```

#### Start Infrastructure

```bash
# Start all infrastructure
cd infrastructure/docker
docker compose -f docker-compose.dev.yml up -d

# View logs
docker compose -f docker-compose.dev.yml logs -f

# Stop infrastructure
docker compose -f docker-compose.dev.yml down

# Stop and remove volumes (fresh start)
docker compose -f docker-compose.dev.yml down -v
```

### Backend Development

#### Project Structure

```
backend/
├── cmd/                    # Service entrypoints
│   ├── usage-service/
│   ├── billing-service/
│   ├── subscription-service/
│   └── ...
├── internal/               # Private packages
│   ├── config/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   └── handler/
├── migrations/             # Database migrations
├── go.mod
└── go.sum
```

#### Install Dependencies

```bash
cd backend
go mod download
go mod verify
```

#### Environment Configuration

```bash
# Copy example environment file
cp .env.example .env.local

# Edit configuration
vim .env.local
```

```env
# .env.local - Copy from .env.example and configure
# SECURITY: Never commit this file with real credentials

# Database - REQUIRED: Set your password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=api_billing
DB_USER=billing
DB_PASSWORD=<your-secure-password>
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Keycloak - REQUIRED: Set your credentials
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=api-billing
KEYCLOAK_CLIENT_ID=api-billing-backend
KEYCLOAK_CLIENT_SECRET=<your-client-secret>

# MinIO - REQUIRED: Set your credentials
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=<your-access-key>
MINIO_SECRET_KEY=<your-secret-key>
MINIO_BUCKET=invoices
MINIO_USE_SSL=false

# Service Ports
USAGE_SERVICE_PORT=8081
BILLING_SERVICE_PORT=8082
SUBSCRIPTION_SERVICE_PORT=8083
ANALYTICS_SERVICE_PORT=8084
NOTIFICATION_SERVICE_PORT=8085
```

#### Run Migrations

```bash
# Install migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations using Makefile (recommended - uses .env automatically)
make db-migrate

# Rollback last migration
make db-migrate-down

# Or run manually with environment variables:
# migrate -path migrations -database "${DB_URL}" up
```

#### Start Services

```bash
# Run all services with hot reload (using air)
go install github.com/cosmtrek/air@latest

# Start individual service
cd cmd/usage-service
air

# Or run directly
go run cmd/usage-service/main.go
```

#### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/service/billing/...

# Run integration tests
go test -tags=integration ./tests/integration/...
```

### Frontend Development

#### Project Structure

```
frontend/
├── apps/
│   ├── admin/              # Admin dashboard
│   └── customer/           # Customer portal
├── packages/
│   ├── ui/                 # Shared components
│   ├── api-client/         # Generated API client
│   └── i18n/               # Translations
├── pnpm-workspace.yaml
├── turbo.json
└── package.json
```

#### Install Dependencies

```bash
cd frontend
bun install
```

#### Environment Configuration

```bash
# Admin dashboard
cp apps/admin/.env.example apps/admin/.env.local

# Customer portal
cp apps/customer/.env.example apps/customer/.env.local
```

```env
# apps/admin/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8000/v1
NEXT_PUBLIC_KEYCLOAK_URL=http://localhost:8080
NEXT_PUBLIC_KEYCLOAK_REALM=api-billing
NEXT_PUBLIC_KEYCLOAK_CLIENT_ID=api-billing-admin
```

#### Start Development Server

```bash
# Start all apps
bun dev

# Start specific app
bun --filter admin dev
bun --filter customer dev
```

#### Build for Production

```bash
# Build all apps
bun build

# Build specific app
bun --filter admin build
```

#### Run Tests

```bash
# Run all tests
bun test

# Run with watch mode
bun test:watch

# Run E2E tests
bun test:e2e
```

### Generate API Client

```bash
# Install openapi-generator
npm install -g @openapitools/openapi-generator-cli

# Generate TypeScript client from contracts
cd frontend
bun generate:api
```

---

## Development Workflows

### Creating a New API Endpoint

1. **Define Contract** (OpenAPI)

   ```yaml
   # specs/001-api-usage-billing/contracts/usage-api.yaml
   /usage/new-endpoint:
     get:
       operationId: getNewEndpoint
       # ...
   ```

2. **Generate Types** (Optional)

   ```bash
   bun generate:api
   ```

3. **Implement Handler** (Go)

   ```go
   // internal/handler/usage/new_endpoint.go
   func (h *Handler) GetNewEndpoint(c *fiber.Ctx) error {
       // Implementation
   }
   ```

4. **Add Route**

   ```go
   // cmd/usage-service/routes.go
   api.Get("/new-endpoint", handler.GetNewEndpoint)
   ```

5. **Write Tests**
   ```go
   // internal/handler/usage/new_endpoint_test.go
   func TestGetNewEndpoint(t *testing.T) {
       // Test implementation
   }
   ```

### Database Changes

1. **Create Migration**

   ```bash
   migrate create -ext sql -dir migrations -seq add_new_table
   ```

2. **Write Migration**

   ```sql
   -- migrations/000001_add_new_table.up.sql
   CREATE TABLE new_table (...);

   -- migrations/000001_add_new_table.down.sql
   DROP TABLE new_table;
   ```

3. **Run Migration**

   ```bash
   make db-migrate
   ```

4. **Update GORM Model**
   ```go
   // internal/domain/new_table.go
   type NewTable struct {
       gorm.Model
       // ...
   }
   ```

### Adding UI Components

1. **Check shadcn/ui**

   ```bash
   cd frontend/packages/ui
   bun dlx shadcn-ui@latest add button
   ```

2. **Create Custom Component**

   ```tsx
   // packages/ui/src/components/custom-component.tsx
   export function CustomComponent() {
     return <div>...</div>;
   }
   ```

3. **Export Component**

   ```tsx
   // packages/ui/src/index.ts
   export * from "./components/custom-component";
   ```

4. **Use in App**
   ```tsx
   import { CustomComponent } from "@repo/ui";
   ```

---

## Makefile Commands

```makefile
# Infrastructure
make infra-up          # Start all infrastructure
make infra-down        # Stop infrastructure
make infra-logs        # View infrastructure logs
make infra-wait        # Wait for services to be healthy

# Database
make db-migrate        # Run migrations
make db-migrate-down   # Rollback last migration
make db-seed           # Seed development data
make db-reset          # Reset database (drop + migrate + seed)

# Backend
make backend-dev       # Start all services with hot reload
make backend-build     # Build all services
make backend-test      # Run all tests
make backend-lint      # Run linter

# Frontend
make frontend-dev      # Start all frontend apps
make frontend-build    # Build all frontend apps
make frontend-test     # Run frontend tests
make frontend-lint     # Run linter

# All-in-one
make dev               # Start everything for development
make test              # Run all tests
make lint              # Run all linters
make build             # Build everything
make clean             # Clean all build artifacts
```

---

## Troubleshooting

### Common Issues

#### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

#### Database Connection Failed

```bash
# Check PostgreSQL is running
docker compose ps postgres

# Check connection
psql -h localhost -U billing -d api_billing
```

#### Redis Connection Failed

```bash
# Check Redis is running
docker compose ps redis

# Test connection
redis-cli ping
```

#### Keycloak Not Starting

```bash
# Check logs
docker compose logs keycloak

# Common fix: wait for PostgreSQL
docker compose restart keycloak
```

### Reset Development Environment

```bash
# Nuclear option - reset everything
make clean
docker compose down -v
docker system prune -f
make infra-up
make db-migrate
make db-seed
```

---

## IDE Setup

### VS Code Extensions

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "42crunch.vscode-openapi"
  ]
}
```

### Go Settings

```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### TypeScript Settings

```json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[typescript]": {
    "editor.formatOnSave": true
  },
  "[typescriptreact]": {
    "editor.formatOnSave": true
  }
}
```

---

## Next Steps

1. Review [data-model.md](./data-model.md) for database schema
2. Review API contracts in [contracts/](./contracts/)
3. Check [plan.md](./plan.md) for implementation phases
4. Run `/speckit.tasks` to generate implementation tasks

---

**Quickstart Complete** | Ready for Development
