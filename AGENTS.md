# Repository Guidelines

## Project Structure & Module Organization

This repo is a Go + Next.js monorepo. Backend services live under `backend/` with domain logic in `backend/internal/`, migrations in `backend/migrations/`, and tests in `backend/tests/` (`unit/`, `integration/`, `contract/`). Frontend apps are in `frontend/apps/` (admin and customer) with shared packages in `frontend/packages/` (UI, API client, i18n). Infrastructure and observability configs are in `infrastructure/`, and feature specs/contracts are in `specs/001-api-usage-billing/`. Generated OpenAPI client code is in `frontend/packages/api-client/src/generated/`.

## Build, Test, and Development Commands

- `make infra-up` / `make infra-wait`: start and wait for local dependencies.
- `make db-migrate`: apply database migrations.
- `make backend-dev` / `make frontend-dev`: run backend services and frontend apps.
- `make lint`, `make test`, `make build`: run lint, tests, and builds for both stacks.
- `bun run generate:openapi` (from `frontend/packages/api-client/`): regenerate the API client from contracts.

## Coding Style & Naming Conventions

Go code follows `gofmt` and `golangci-lint`; use lowercase package names, PascalCase for exported types, and keep file names short and descriptive (e.g., `service.go`, `handler.go`). TypeScript uses ESLint + Prettier (single quotes, semicolons, print width 100). Avoid manual edits in `frontend/packages/api-client/src/generated/`; regenerate instead.

## Testing Guidelines

Backend uses `go test` (testify/gomock as needed) with unit, integration, and contract tests (`*_test.go`). Frontend uses Jest/React Testing Library and Playwright; E2E specs follow `*.e2e.ts` in app directories. Coverage targets: ≥95% for billing services and ≥80% overall. Run `make test` or `go test ./...` / `bun test` for focused runs.

## Commit & Pull Request Guidelines

Git history currently only has an initial template commit, so no established convention. Use concise, imperative messages; adding a scope (e.g., `backend:`, `frontend:`) is encouraged. PRs should include a brief summary, linked issue (if any), test evidence, and screenshots for UI changes. Note any migration impacts explicitly.

## Security & Configuration Tips

Do not commit secrets; use `.env` files or environment variables. For migrations, keep schemas compatible with PostgreSQL partitioning rules (primary keys must include partition columns).
