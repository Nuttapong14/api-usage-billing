# API Documentation

Welcome to the API Usage Analytics & Billing Platform API documentation. This site links to the OpenAPI contracts and explains how to integrate with the platform.

## Base URL

Production API base URL pattern:

```
https://api.{organization}.example.com/v1
```

## Authentication

- **BearerAuth**: OAuth2/JWT access tokens from Keycloak.
- **ApiKeyAuth**: API keys for usage endpoints (Kong gateway).

## OpenAPI Contracts

- API Key Management: `specs/001-api-usage-billing/contracts/apikey-api.yaml`
- Billing: `specs/001-api-usage-billing/contracts/billing-api.yaml`
- Subscription: `specs/001-api-usage-billing/contracts/subscription-api.yaml`
- Usage: `specs/001-api-usage-billing/contracts/usage-api.yaml`
- Webhooks: `specs/001-api-usage-billing/contracts/webhook-api.yaml`

## Generated Client

The generated TypeScript client is available in `frontend/packages/api-client/src/generated`.

Regenerate after contract changes:

```
cd frontend/packages/api-client
bun run generate:openapi
```

## Request/Response Conventions

- All responses follow a consistent envelope with `success`, `data`, and `error` fields.
- Pagination uses `limit` and `offset` query parameters, returning a `pagination` object.
- Rate limits return standard `X-RateLimit-*` headers.
