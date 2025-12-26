# Usage Analytics & Billing Platform

A production-ready platform for real-time API usage tracking, automated billing, and subscription management. Designed for API providers to monetize their services effectively.

## 🚀 Features

- **Real-time Usage Tracking**: High-performance tracking pipeline using Kong Gateway, Redis Streams, and Go workers.
- **Flexible Subscription Management**: Support for multiple tiers (Free, Pro, Enterprise) with custom quotas and rate limits.
- **Automated Billing & Invoicing**: Monthly invoice generation, PDF export, and multi-currency support.
- **Analytics Dashboard**: Comprehensive views for both providers and consumers to track usage, costs, and performance.
- **API Key Management**: Secure key generation, rotation, and lifecycle management with zero downtime.
- **Scalable Architecture**: Built on Kubernetes, auto-scalable components, and separate read/write paths.

## 🛠 Tech Stack

### Backend

- **Language**: Go (Golang) 1.22+
- **Framework**: Fiber v2
- **Database**: PostgreSQL 16 (Primary), Redis Cluster (Cache & Streams)
- **Storage**: MinIO (S3-compatible)

### Frontend

- **Framework**: Next.js 14
- **Styling**: Tailwind CSS + shadcn/ui
- **Build System**: Turborepo (Monorepo)

### Infrastructure

- **Gateway**: Kong Gateway
- **Authentication**: Keycloak (OIDC)
- **Orchestration**: Docker Compose (Dev), Kubernetes (Prod)
- **Observability**: Prometheus, Grafana, Loki, Jaeger

## 🏁 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.22+
- Node.js (Latest LTS) & bun
- Make

### Installation & Run

1.  **Clone the repository**

    ```bash
    git clone <repository-url>
    cd Usage-Analytics-and-Billing
    ```

2.  **Start Infrastructure**
    Spin up dependencies (Postgres, Redis, Kong, Keycloak, etc.).

    ```bash
    make infra-up
    ```

3.  **Run Migrations**
    Initialize the database schema.

    ```bash
    make db-migrate
    ```

4.  **Start Backend Services**
    Run the Go backend services in development mode with hot reload (Air).

    ```bash
    make backend-dev
    ```

5.  **Start Frontend Applications**
    Run the Next.js applications (Admin Dashboard & Customer Portal).
    ```bash
    make frontend-dev
    ```

### Useful Commands

- `make infra-logs` - View logs from infrastructure containers.
- `make infra-down` - Stop and remove infrastructure containers.
- `make test` - Run both backend and frontend tests.
- `make lint` - Run linters for both backend and frontend.

## 📂 Project Structure

```
├── backend/                # Go microservices and shared logic
│   ├── cmd/                # Service entry points
│   ├── internal/           # Core domain logic (Clean Architecture)
│   └── migrations/         # Database migrations
├── frontend/               # Next.js monorepo
│   ├── apps/               # Admin and Customer applications
│   └── packages/           # Shared UI and utility packages
├── infrastructure/         # Infrastructure as Code
│   ├── docker/             # Docker Compose configurations
│   └── kubernetes/         # K8s manifests
├── specs/                  # Detailed technical specifications
└── docs/                   # Project documentation
```

## 📖 Documentation

- [Product Requirements Document (PRD)](./prd-api-analytics-billing.md)
- [API Specifications](./specs/001-api-usage-billing/spec.md)
- [Data Model](./specs/001-api-usage-billing/data-model.md)

## 🤝 Contributing

Contributions are welcome! Please read our contribution guidelines and code of conduct before submitting pull requests.

## 📄 License

[MIT License](LICENSE)
