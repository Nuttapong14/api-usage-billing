SHELL := /bin/bash

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# DB_PASSWORD must be set in .env or environment - no default for security
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR ?= backend/migrations
BACKEND_SERVICE ?= usage-service

# Validation target to check required environment variables
.PHONY: check-env
check-env:
	@if [ -z "$(DB_PASSWORD)" ]; then \
		echo "ERROR: DB_PASSWORD is not set."; \
		echo "Please copy .env.example to .env and configure your credentials:"; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi

.PHONY: \
	infra-up infra-down infra-logs infra-wait \
	db-migrate db-migrate-down db-seed db-reset \
	backend-dev backend-build backend-test backend-lint \
	frontend-dev frontend-build frontend-test frontend-lint \
	dev test lint build clean clean-all pre-push

infra-up:
	cd infrastructure/docker && docker compose -f docker-compose.dev.yml up -d

infra-down:
	cd infrastructure/docker && docker compose -f docker-compose.dev.yml down -v

infra-logs:
	cd infrastructure/docker && docker compose -f docker-compose.dev.yml logs -f

infra-wait:
	@echo "Waiting for infrastructure services to be healthy..."


db-migrate: check-env
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

db-migrate-down: check-env
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

db-seed:
	@echo "Seed scripts not implemented yet."

db-reset: check-env
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down -all
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up


backend-dev:
	cd backend/cmd/$(BACKEND_SERVICE) && air

backend-build:
	cd backend && go build ./...

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && golangci-lint run


frontend-dev:
	cd frontend && bun dev

frontend-build:
	cd frontend && bun build

frontend-test:
	cd frontend && bun test

frontend-lint:
	cd frontend && bun lint


dev: infra-up
	@echo "Start backend and frontend in separate terminals:"
	@echo "- make backend-dev"
	@echo "- make frontend-dev"

test:
	make backend-test
	make frontend-test

lint:
	make backend-lint
	make frontend-lint

build:
	make backend-build
	make frontend-build

clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin backend/tmp
	rm -rf frontend/node_modules frontend/.turbo
	rm -rf frontend/apps/*/.next frontend/apps/*/node_modules
	rm -rf frontend/packages/*/node_modules frontend/packages/*/dist
	@echo "Clean complete!"

clean-all: clean
	@echo "Removing lock files..."
	rm -rf frontend/bun.lock
	@echo "Full clean complete! Run 'bun install' in frontend/ to reinstall."

# Pre-push check - verify nothing sensitive is being committed
pre-push:
	@echo "Checking for sensitive files..."
	@if git diff --cached --name-only | grep -E '\.env$$|\.env\.local|\.env\.production$$'; then \
		echo "ERROR: Attempting to commit sensitive .env files!"; \
		exit 1; \
	fi
	@echo "No sensitive files detected."
	@echo "Checking git status..."
	@git status --short
