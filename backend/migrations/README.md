# Database Migrations

This directory contains PostgreSQL migrations for the API Usage Analytics & Billing Platform.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Installation

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Prerequisites

Before running migrations, ensure your environment is configured:

```bash
# Copy and configure environment file
cp .env.example .env
# Edit .env with your database credentials
```

### Usage

**Recommended**: Use the Makefile targets which handle environment variables automatically:

```bash
# Run all pending migrations
make db-migrate

# Rollback last migration
make db-migrate-down

# Reset database (drop + migrate)
make db-reset
```

#### Manual Commands (Advanced)

If you need to run migrations manually, use environment variables:

```bash
# Set database URL (never hardcode credentials)
export DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Run all pending migrations
migrate -path migrations -database "${DB_URL}" up

# Rollback last migration
migrate -path migrations -database "${DB_URL}" down 1

# Check migration status
migrate -path migrations -database "${DB_URL}" version
```

#### Create new migration
```bash
migrate create -ext sql -dir migrations -seq migration_name
```

## Migration Naming Convention

Files are named with a sequence number and descriptive name:
- `000001_create_organizations.up.sql`
- `000001_create_organizations.down.sql`
- `000002_create_customers.up.sql`
- `000002_create_customers.down.sql`

## Migration Order

Migrations must be run in sequence as they have dependencies:

1. `001_create_organizations` - Root multi-tenant entity
2. `002_create_customers` - API consumers (depends on organizations)
3. `003_create_subscription_tiers` - Pricing plans (depends on organizations)
4. `004_create_subscriptions` - Customer tier assignments
5. `005_create_api_keys` - Authentication credentials
6. `006_create_usage_records_partitioned` - API usage tracking (partitioned)
7. `007_create_usage_daily` - Aggregated daily stats
8. `008_create_invoices` - Billing documents
9. `009_create_payments` - Payment transactions
10. `010_create_webhook_endpoints` - Customer webhooks
11. `011_create_webhook_deliveries` - Webhook delivery log
12. `012_create_audit_logs_partitioned` - Audit trail (partitioned)
13. `013_create_credit_notes` - Invoice adjustments
14. `014_create_indexes` - Performance indexes
15. `015_create_partition_maintenance` - Automated partition management

## Partition Management

Tables `usage_records` and `audit_logs` use PostgreSQL declarative partitioning by month.

A maintenance function automatically creates partitions for the next 3 months. This should be scheduled via:
- Kubernetes CronJob (recommended)
- pg_cron extension
- External scheduler

## Best Practices

1. **Atomic Migrations**: Each migration should be a single logical change
2. **Reversible**: Always provide both `up` and `down` migrations
3. **Test Rollbacks**: Verify `down` migrations work before deploying
4. **No Data Loss**: Down migrations should preserve data where possible
5. **Idempotent**: Migrations should be safe to run multiple times
6. **No Hardcoded Credentials**: Always use environment variables for database connections
