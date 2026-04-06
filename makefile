# Database Configuration
DB_STRING = postgres://postgres:postgres@localhost:5400/postgres?sslmode=disable

# Coverage configuration
COVERAGE_DIR = ./coverage/unit

.PHONY: help up down build test lint generate-mocks create-migration migrate-up migrate-down migrate-reset show-test-coverage clean

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  up                - Start dev environment"
	@echo "  down              - Stop dev environment"
	@echo "  build             - Build application binary"
	@echo "  test              - Run tests"
	@echo "  lint              - Run linter"
	@echo "  generate-mocks    - Generate mock code"
	@echo "  create-migration  - Create new migration (NAME=migration_name)"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback last migration"
	@echo "  migrate-reset     - Reset database"
	@echo "  show-test-coverage - Show test coverage"
	@echo "  clean             - Clean build artifacts"

up:
	docker compose up -d

down:
	docker compose down --remove-orphans

build:
	mkdir -p bin
	go build -o bin/app cmd/main.go

test:
	go test -v -race ./...

lint:
	golangci-lint run -v

generate-mocks:
	go generate ./...

create-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration NAME=your_migration_name"; \
		exit 1; \
	fi
	GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="$(DB_STRING)" goose create $(NAME) sql

migrate-up:
	GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="$(DB_STRING)" goose up

migrate-down:
	GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="$(DB_STRING)" goose down

migrate-reset:
	GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="$(DB_STRING)" goose reset

show-test-coverage:
	mkdir -p $(COVERAGE_DIR)
	go test -cover ./... -args -test.gocoverdir="$(shell pwd)/$(COVERAGE_DIR)"
	go tool covdata percent -i=$(COVERAGE_DIR)

clean:
	rm -rf bin/
	rm -rf $(COVERAGE_DIR)
