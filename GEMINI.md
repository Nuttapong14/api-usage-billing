# Usage Analytics & Billing Platform

## Project Overview

This project is a comprehensive platform for real-time API usage tracking and automated billing. It is designed to help API providers monetize their services by offering features like subscription management, usage analytics, and automated invoicing.

**Key Features:**

- **Real-time Usage Tracking:** Tracks API requests via Kong Gateway and Redis Streams.
- **Subscription Management:** Supports multiple tiers (Free, Pro, Enterprise) with quota enforcement.
- **Billing & Invoicing:** Automated monthly invoicing, PDF generation, and payment gateway integration.
- **Analytics Dashboard:** Visualizes usage trends, revenue, and top consumers.
- **API Key Management:** Secure key generation, rotation, and lifecycle management.

## Architecture

The system follows a microservices architecture with a separate frontend monorepo.

### Backend (`/backend`)

- **Language:** Go (Golang)
- **Framework:** Fiber v2
- **Database:** PostgreSQL (Primary), Redis (Cache/Streams)
- **Structure:**
  - `cmd/`: Entry points for services (analytics, billing, usage, etc.).
  - `internal/`: Core logic (Clean Architecture: Domain, Service, Repository, Handler).
  - `migrations/`: Database schema migrations.

### Frontend (`/frontend`)

- **Framework:** Next.js (React)
- **Build Tool:** Turborepo (Monorepo)
- **Package Manager:** bun
- **Apps:**
  - `apps/admin`: Dashboard for API providers.
  - `apps/customer`: Portal for API consumers.
- **Packages:** Shared UI components, API clients, etc.

### Infrastructure (`/infrastructure`)

- **Gateway:** Kong Gateway
- **Auth:** Keycloak
- **Orchestration:** Docker Compose (Local), Kubernetes (Prod)
- **Observability:** Prometheus, Grafana

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.22+
- Node.js & bun
- Make

### Quick Start

1.  **Start Infrastructure:**

    ```bash
    make infra-up
    ```

    This spins up Postgres, Redis, Kong, Keycloak, etc.

2.  **Run Migrations:**

    ```bash
    make db-migrate
    ```

3.  **Start Backend (in a new terminal):**

    ```bash
    make backend-dev
    ```

4.  **Start Frontend (in a new terminal):**
    ```bash
    make frontend-dev
    ```

## Development Workflow

- **Makefile:** Use the `Makefile` in the root for common tasks.
  - `make infra-logs`: View infrastructure logs.
  - `make backend-test`: Run backend tests.
  - `make frontend-lint`: Lint frontend code.
  - `make clean`: Clean up build artifacts.

- **Documentation:**
  - `prd-api-analytics-billing.md`: Product Requirements Document (Source of Truth).
  - `specs/`: Detailed technical specifications.

## Conventions

- **Backend:** Follows Clean Architecture. Logic should reside in `internal/service`, data access in `internal/repository`.
- **Frontend:** Uses Turborepo. Shared code should go into `packages/`.
- **Commits:** Follow conventional commits (e.g., `feat: add billing service`, `fix: calculation error`).
