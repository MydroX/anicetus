init:
	@echo "Welcome! Enjoy the ride!"

up:
	@echo "Starting..."
	@docker-compose -f deploy/docker-compose.yml up --build

down:
	@echo "Stopping..."
	@docker-compose -f deploy/docker-compose.yml down
	
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
	@GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" goose reset

lint:
	@echo "Linting..."
	@golangci-lint -v run

go-generate:
	@echo "Generating code..."
	@go generate ./...