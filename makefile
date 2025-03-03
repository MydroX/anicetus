init:
	@echo "Welcome! Enjoy the ride!"

up-dev:
	@echo "Starting..."
	@docker compose -f deploy/docker-compose.yml -f deploy/dev/docker-compose.yml up --build

up-prod:
	@echo "Starting..."
	@docker compose -f deploy/docker-compose.yml -f deploy/prod/docker-compose.yml up --build

down-dev:
	@echo "Stopping..."
	@docker compose -f deploy/docker-compose.yml -f deploy/dev/docker-compose.yml down
	
build:
	@echo "Building..."
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

create-migration:
	@echo "Creating migration..."
	@GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" goose create $(NAME) sql

migrate-up:
	@echo "Migrating up..."
	@GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" goose up

migrate-reset:
	@echo "Reset database..."
	@GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations	 GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" goose reset

lint:
	@echo "Linting..."
	@golangci-lint -v run

generate-mocks:
	@echo "Generating code..."
	@go generate ./...