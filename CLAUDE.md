# Anicetus

Lightweight, high-performance open-source IAM service (Keycloak alternative).

## Build & Run

```bash
make build           # Build binary
make up              # Start dev environment (Postgres + Valkey + app with hot reload)
make down            # Stop environment
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
  api.go                # Server setup, DI wiring, router
  config/               # Config structs
  identity/             # Identity domain (user CRUD, registration, profile)
  authentication/       # Authentication domain (login, logout, tokens, sessions, auth middleware)
  authorization/        # Authorization domain (RBAC: roles, permissions, middleware)
  services/             # Services domain (service/audience registration, assignment, Valkey caching)
  middlewares/          # Shared HTTP middleware (trace ID)
pkg/                    # Reusable packages (cache, db, logger, password, argon2id, uuid, jwt, errs)
migrations/             # Goose SQL migrations
deploy/                 # Docker configs (dev/prod)
.bruno/                 # Bruno API collection
```

Each domain follows: `controller/ → usecases/ → repository/ → models/ + dto/ + mocks/`

### Domain Dependencies

```
identity         (standalone)
services         (standalone, provides JWT audiences via ServiceManager)
authorization    (standalone, provides permissions via AuthorizationRepository)
authentication   (depends on: identity repo, services ServiceManager, jwt pkg)
```

## Conventions

- Error codes: `ErrorXxx` constants (10xxx common, 11xxx users, 12xxx IAM, 13xxx authorization, 14xxx sessions, 99xxx DB)
- Interfaces for all layers (testability + mocking)
- Standardized JSON error responses: `{message, code, trace_id}`
- SQL queries as methods on `Queries` struct (raw SQL, no ORM)
- Zap SugaredLogger throughout
- Constructor-based DI, no framework

## Stack

Go 1.26 | Gin | pgx/v5 (Postgres) | Valkey (cache) | golang-jwt/v5 | Argon2id + bcrypt | Zap | Viper | Goose

## TODO

1. OpenAPI spec — add swaggo/swag annotations to generate API docs automatically
2. Fix security issues:
  - No rate limiting on login (needs a rate limiter library/middleware)                                                                                                                                                                                                                
  - Authorization middleware never applied / RequireRole is a no-op stub                                                                                                                                                                                                               
  - Missing SameSite cookie attribute                                                                                                                                                                                                                                                  
  - Access tokens remain valid after logout (needs Valkey blacklist)                                                                                                                                                                                                                   
  - No session invalidation on password change        
  - Empty validate.go — no config validation for session/hash params
