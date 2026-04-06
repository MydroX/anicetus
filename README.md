# Anicetus

![coverage](https://raw.githubusercontent.com/MydroX/anicetus/badges/.badges/main/coverage.svg)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/MydroX/anicetus?v=2)

Anicetus is a lightweight, security-first identity and access management (IAM) service built in Go. It provides authentication, session management, and multi-service token scoping — the core of what Keycloak does, without the weight.

> **Status:** Early development. The core auth flow works but the API is not stable yet. Contributions and feedback are welcome.

## Philosophy

A focused IAM service for modern backends — built for simplicity, security, and performance.

- Built in Go — no JVM, no heavy runtime. A small, readable codebase with predictable performance
- Security-first — designed with strict security principles and safe defaults from the ground up
- One binary, one config — no XML, no complex setup, no hidden magic
- Sane defaults — secure out of the box, configurable when needed
- No plugins — fewer moving parts, smaller attack surface
- Focused scope — authentication, session management, and audience scoping done right
- Simple operations — easy to run locally or in production (Docker-ready)

If you need SAML, LDAP federation, or identity brokering — Keycloak is the right choice.

## Features

- **JWT authentication** with access + refresh token flow
- **Per-user audience scoping** — control which services a user's tokens are valid for
- **Session management** — tracked sessions with device/browser info, hashed refresh tokens
- **User management** — registration, login, CRUD operations
- **Audience management** — register services, assign users, revoke access

## Security

As an IAM service, security is a core priority — not an afterthought.

### Implemented

- **JWT signing** — HS512 with enforced minimum 32-char secret
- **Token duration bounds** — access tokens max 1h, refresh tokens max 30d, validated at startup
- **Argon2id** for refresh token hashing (configurable iterations, memory, parallelism)
- **bcrypt** (cost 14) for password hashing with configurable complexity rules
- **HttpOnly + Secure cookies** — prevents XSS access to tokens
- **Audience validation** — tokens are rejected if their `aud` claim doesn't match expected audiences
- **Issuer validation** — tokens must match the configured issuer
- **Clock skew tolerance** — configurable tolerance for time-based claims (exp, iat, nbf)
- **Constant-time comparison** for token verification (prevents timing attacks)
- **Parameterized SQL** throughout (no injection risk)
- **Config validation at startup** — fails fast on weak secrets, invalid durations, or missing config

### Planned

- **TOTP** — time-based one-time password for two-factor authentication
- **Separate signing keys** for access and refresh tokens
- **Refresh token rotation** — issue new refresh token on use, invalidate the old one
- **Rate limiting** on authentication endpoints (login, register)
- **Uniform error responses** on login to prevent user enumeration
- **Token revocation / blacklisting** — invalidate tokens before expiry
- **RBAC / permissions** — role-based access control in token claims
- **CORS configuration**
- **Request signing** for admin endpoints

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

[Apache License 2.0](LICENSE)
