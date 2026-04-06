# Anicetus

![coverage](https://raw.githubusercontent.com/MydroX/anicetus/badges/.badges/main/coverage.svg)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/MydroX/anicetus)

Anicetus is a lightweight identity and access management (IAM) service built in Go. It provides authentication, session management, and multi-service token scoping — the core of what Keycloak does, without the weight.

> **Status:** Early development. The core auth flow works but the API is not stable yet. Contributions and feedback are welcome.

## Features

- **JWT authentication** with access + refresh token flow
- **Per-user audience scoping** — control which services a user's tokens are valid for
- **Session management** — tracked sessions with device/browser info, hashed refresh tokens
- **User management** — registration, login, CRUD operations
- **Audience management** — register services, assign users, revoke access

### Security

- HS512 JWT signing with configurable secret, issuer, and expiration bounds
- Argon2id hashing for refresh tokens
- bcrypt for passwords with configurable complexity rules
- HttpOnly + Secure cookies
- Audience validation on token parse
- Config validation at startup (secret length, token duration bounds)

## Quick Start

### Requirements

- Docker & Docker Compose

### Run

```bash
git clone https://github.com/MydroX/anicetus.git
cd anicetus
make up-dev
make migrate-up
```

The API is available at `http://localhost:3000`. A [Bruno](https://www.usebruno.com/) collection is included in `.bruno/` to test all endpoints.

### Configuration

Configuration is in `cmd/config.yaml`:

```yaml
jwt:
  secret: "your-secret-key-min-32-chars"
  issuer: "anicetus"
  skew: 60
  access_token:
    expiration: 900      # seconds
  refresh_token:
    expiration: 86400    # seconds

session:
  hash:
    iterations: 3
    memory: 131072       # 128MB
    parallelism: 4
    key_length: 32
    salt_length: 16
```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/register` | Register a new user |
| POST | `/api/v1/login` | Login (returns access + refresh tokens) |
| GET | `/api/v1/refresh` | Refresh access token |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/` | List all users |
| GET | `/api/v1/users/:uuid` | Get user |
| PUT | `/api/v1/users/:uuid` | Update user |
| PATCH | `/api/v1/users/:uuid/email` | Update email |
| PATCH | `/api/v1/users/:uuid/password` | Update password |
| DELETE | `/api/v1/users/:uuid` | Delete user |

### Audiences

Audiences define which services a user's tokens are valid for. Assign audiences to users to control cross-service access.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/audiences` | Register a new audience |
| GET | `/api/v1/audiences` | List all audiences |
| DELETE | `/api/v1/audiences/:audience` | Revoke an audience |
| GET | `/api/v1/audiences/users/:uuid` | Get user's audiences |
| POST | `/api/v1/audiences/users/:uuid` | Assign audience to user |

## Contributing

### Setup

```bash
# Prerequisites: Go 1.26+, Docker
make up-dev          # Start Postgres + Valkey + app (hot reload)
make migrate-up      # Run database migrations
make test            # Run tests
make lint            # Run linter
make generate-mocks  # Regenerate mocks after interface changes
```

### Project Structure

```
internal/
  common/jwt/       JWT service, validation, token types
  iam/              Login, sessions, audience management
  users/            User CRUD
pkg/                Reusable packages (cache, db, crypto, logging)
migrations/         SQL migrations (Goose)
.bruno/             API collection for testing
```

Each domain follows clean architecture: `controller → usecases → repository`

### What Needs Help

- **Open source** I am new to open source so feel free to suggest any rules or processes for contributing, code style, etc.
- **Features** The core auth flow is implemented but there are many features that would be great to add.
- **Dashboard** A simple admin dashboard to manage users and audiences would be a great addition.
- **More tests** The core auth flow is tested but more coverage is needed, especially for

## License

[MIT](LICENSE)
