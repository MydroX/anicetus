# Application Configuration
APP_NAME ?= myapp

# Database Configuration (Development only)
DB_STRING = postgres://postgres:postgres@localhost:5400/postgres?sslmode=disable

# Docker Compose configuration
COMPOSE_FILES = -f deploy/docker-compose.yml
DEV_COMPOSE = $(COMPOSE_FILES) -f deploy/dev/docker-compose.yml
PROD_COMPOSE = $(COMPOSE_FILES) -f deploy/prod/docker-compose.yml

# Coverage configuration
COVERAGE_DIR = ./coverage/unit

# Declare phony targets for better performance
.PHONY: init help up build down test dev prod
.PHONY: create-migration migrate-up migrate-down migrate-reset lint generate-mocks
.PHONY: show-test-coverage clean

# Default target
.DEFAULT_GOAL := help

## Display help information
help:
	@echo "Available commands:"
	@echo "  init              - Initialize the project"
	@echo "  up [env]          - Start environment (dev/prod)"
	@echo "  build [env]       - Build environment (dev/prod) or application binary"
	@echo "  down [env]        - Stop environment (dev/prod)"
	@echo "  tests              - Run tests"
	@echo "  create-migration  - Create new migration (requires NAME=migration_name)"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback last migration"
	@echo "  migrate-reset     - Reset database"
	@echo "  lint              - Run linter"
	@echo "  generate-mocks    - Generate mock code"
	@echo "  show-test-coverage - Show test coverage"
	@echo "  clean             - Clean build artifacts"

## Initialize project
init:
	@echo "🚀 Welcome! Enjoy the ride!"

## Environment Management
up:
	@if [ "$(filter dev,$(MAKECMDGOALS))" ]; then \
		echo "🏗️  Starting development environment..."; \
		docker compose $(DEV_COMPOSE) up -d; \
	elif [ "$(filter prod,$(MAKECMDGOALS))" ]; then \
		echo "🚀 Starting production environment..."; \
		docker compose $(PROD_COMPOSE) up -d; \
	else \
		echo "❌ Usage: make up dev  OR  make up prod"; \
		exit 1; \
	fi

build:
	@if [ "$(filter dev,$(MAKECMDGOALS))" ]; then \
		echo "🔨 Building development environment..."; \
		docker compose $(DEV_COMPOSE) build --parallel; \
	elif [ "$(filter prod,$(MAKECMDGOALS))" ]; then \
		echo "🔨 Building production environment..."; \
		docker compose $(PROD_COMPOSE) build --parallel; \
	elif [ -z "$(filter dev prod,$(MAKECMDGOALS))" ]; then \
		echo "🔨 Building application..."; \
		mkdir -p bin; \
		go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go; \
	else \
		echo "❌ Usage: make build dev  OR  make build prod  OR  make build"; \
		exit 1; \
	fi

down:
	@if [ "$(filter dev,$(MAKECMDGOALS))" ]; then \
		echo "🛑 Stopping development environment..."; \
		docker compose $(DEV_COMPOSE) down --remove-orphans; \
	elif [ "$(filter prod,$(MAKECMDGOALS))" ]; then \
		echo "🛑 Stopping production environment..."; \
		docker compose $(PROD_COMPOSE) down --remove-orphans; \
	else \
		echo "❌ Usage: make down dev  OR  make down prod"; \
		exit 1; \
	fi

# Dummy targets to avoid "No rule to make target" errors
dev:
	@:

prod:
	@:

## Database Migrations
tests:
	@echo "🧪 Running tests..."
	@go test -v -race ./...
create-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ ERROR: Migration name required. Usage: make create-migration NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)"
	@GOOSE_DRIVER=postgres \
	 GOOSE_MIGRATION_DIR=migrations \
	 GOOSE_DBSTRING="$(DB_STRING)" \
	 goose create $(NAME) sql

migrate-up:
	@echo "⬆️  Running database migrations..."
	@GOOSE_DRIVER=postgres \
	 GOOSE_MIGRATION_DIR=migrations \
	 GOOSE_DBSTRING="$(DB_STRING)" \
	 goose up

migrate-down:
	@echo "⬇️  Rolling back last migration..."
	@GOOSE_DRIVER=postgres \
	 GOOSE_MIGRATION_DIR=migrations \
	 GOOSE_DBSTRING="$(DB_STRING)" \
	 goose down

migrate-reset:
	@echo "🔄 Resetting database..."
	@GOOSE_DRIVER=postgres \
	 GOOSE_MIGRATION_DIR=migrations \
	 GOOSE_DBSTRING="$(DB_STRING)" \
	 goose reset

## Code Quality
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run -v

generate-mocks:
	@echo "🎭 Generating mock code..."
	@go generate ./...
	@echo "🥸 Mocks generated! Enjoy writing tests!"

show-test-coverage:
	@echo "📊 Generating test coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -cover ./... -args -test.gocoverdir="$(shell pwd)/$(COVERAGE_DIR)"
	@go tool covdata percent -i=$(COVERAGE_DIR)

## Cleanup
clean:
	@echo "🧹 Cleaning up build artifacts..."
	@rm -rf bin/
	@rm -rf $(COVERAGE_DIR)
	@echo "✅ Cleanup complete"