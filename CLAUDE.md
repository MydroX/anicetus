# Anicetus

Lightweight, high-performance open-source IAM service (Keycloak alternative).

## Build & Run

```bash
make build           # Build binary
make up-dev          # Docker Compose dev (hot reload via Air)
make up-prod         # Docker Compose prod
make down-dev        # Stop dev
```

App runs on `:3000`, Postgres on `:5400`, Valkey on `:6379`.

## Test & Lint

```bash
make test            # go test -v ./...
make lint            # golangci-lint -v run
make generate-mocks  # Regenerate uber/mock mocks
make migrate-up      # Run goose migrations
make migrate-down    # Rollback one migration
```

## Architecture

Clean Architecture: `controller → usecases → repository`

```
internal/
  api.go              # Server setup, DI wiring, router
  config/             # Config structs
  common/             # Shared: errorsutil, jwt, context, response
  iam/                # IAM domain (login, audiences, sessions)
  users/              # Users domain (CRUD)
  middlewares/        # HTTP middleware
pkg/                  # Reusable packages (cache, db, logger, password, argon2id, uuid)
migrations/           # Goose SQL migrations
deploy/               # Docker configs (dev/prod)
.bruno/               # Bruno API collection
```

Each domain follows: `controller/ → usecases/ → repository/ → models/ + dto/ + mocks/`

## Conventions

- Error codes: `ErrorXxx` constants (10xxx common, 11xxx users, 12xxx IAM, 99xxx DB)
- Interfaces for all layers (testability + mocking)
- `AppContext` wraps gin.Context with trace IDs
- Standardized JSON error responses: `{message, code, trace_id}`
- SQL queries as methods on `Queries` struct (raw SQL, no ORM)
- Zap SugaredLogger throughout
- Constructor-based DI, no framework

## Stack

Go 1.26 | Gin | pgx/v5 (Postgres) | Ristretto (cache) | golang-jwt/v5 | Argon2id + bcrypt | Zap | Viper | Goose

## TODO

1. Remove prefix UUIDs — switch to native Postgres UUID columns, drop ValidateWithPrefix/NewWithPrefix/prefix constants
2. Introduce Valkey — replace Ristretto for distributed caching (horizontal scaling)
3. OpenAPI spec — add swaggo/swag annotations to generate API docs automatically
4. Fix security issues:
   - Single shared JWT secret for both token types
   - User enumeration (different errors for wrong password vs not found)
   - No rate limiting on login
   - No refresh token rotation
   - Empty validate.go — no config validation for session/hash params
