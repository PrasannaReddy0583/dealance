# Dealance Backend

> **Instagram + WhatsApp + LinkedIn + PhonePe + NetBanking** — for the Indian startup investment ecosystem.

## Architecture

Dealance is a **Go microservices monorepo** with Clean Architecture and polyglot database design:

| Service | Port | Database | Purpose |
|---------|------|----------|---------|
| `auth` | 8080 | PostgreSQL + ScyllaDB | Identity, authentication, KYC |
| `user` | 8081 | PostgreSQL + Neo4j | Profiles, social graph |
| `content` | 8082 | PostgreSQL + ScyllaDB | Posts, shorts, videos |
| `startup` | 8083 | PostgreSQL + Typesense | Startup profiles, search |
| `deal` | 8084 | PostgreSQL + Neo4j | Deal rooms, investments |
| `wallet` | 8085 | PostgreSQL | Wallets, escrow, ledger |
| `chat` | 8086 | PostgreSQL + ScyllaDB | Real-time messaging |
| `media` | 8087 | PostgreSQL + S3 | Upload pipeline |
| `feed` | 8088 | Redis | Feed computation |
| `notify` | 8089 | ScyllaDB | Push/email/SMS |
| `admin` | 8090 | PostgreSQL | Ops dashboard |

## Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- Make
- OpenSSL (for key generation)

### 1. Start Infrastructure

```bash
make infra-up
```

This starts PostgreSQL, Redis, ScyllaDB, Neo4j, Typesense, MailHog, and LocalStack.

### 2. Generate JWT Keys

```bash
make keys
```

### 3. Run Migrations

```bash
make migrate-sql service=auth
```

### 4. Copy Environment File

```bash
cp .env.example services/auth/.env
```

### 5. Start Auth Service

```bash
make dev service=auth
```

### 6. Verify

```bash
curl http://localhost:8080/health
# → {"success":true,"data":{"service":"dealance-auth","status":"healthy"}}
```

## Testing

```bash
# Run all tests
make test

# Auth service only
make test-auth

# With coverage report
make test-coverage
```

## Auth Service API

### Signup Flow (5 stages, must advance in order)

```bash
# 1. Initiate signup
curl -X POST http://localhost:8080/auth/signup/initiate \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# 2. Verify email OTP (check MailHog at http://localhost:8025)
curl -X POST http://localhost:8080/auth/signup/verify-email \
  -H "Content-Type: application/json" \
  -d '{"session_id":"<session_id>","otp":"<otp>"}'

# 3. Confirm auth method
curl -X POST http://localhost:8080/auth/signup/confirm-auth \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<user_id>","provider_type":"PASSKEY","external_id":"cred_123"}'

# 4. Set country
curl -X POST http://localhost:8080/auth/signup/country \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<user_id>","country_code":"IN"}'

# 5. Set role
curl -X POST http://localhost:8080/auth/signup/role \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<user_id>","roles":["ENTREPRENEUR"]}'
```

### Login

```bash
# Email OTP login (always returns 200 — anti-enumeration)
curl -X POST http://localhost:8080/auth/login/email/begin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Verify OTP
curl -X POST http://localhost:8080/auth/login/email/finish \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"<otp>"}'
```

### Token Management

```bash
# Refresh token
curl -X POST http://localhost:8080/auth/token/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'

# Logout (requires JWT)
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

## Security Layers

Every request passes through 4 security layers (skippable in dev):

1. **App Attestation** — Apple App Attest / Google Play Integrity
2. **Device Signing** — HMAC-SHA256 request signing with nonce replay prevention
3. **JWT RS256** — Access tokens with JTI blacklist support
4. **RBAC + KYC Gates** — Role and verification checks per route

## Project Structure

```
dealance/
├── go.work
├── Makefile
├── .env.example
├── shared/                   # Shared packages and middleware
│   ├── pkg/                  # response, logger, crypto, jwt, pagination
│   ├── middleware/            # All 4 security layers + helpers
│   └── domain/               # Shared entity types and errors
├── services/
│   └── auth/                 # Auth microservice
│       ├── cmd/server/        # Entry point
│       ├── config/            # Viper configuration
│       ├── migrations/        # PostgreSQL + ScyllaDB schemas
│       └── internal/
│           ├── domain/        # Entities, DTOs, repository interfaces
│           ├── application/   # Business logic use cases
│           ├── infrastructure/# PostgreSQL, Redis, ScyllaDB repos
│           └── transport/http/# Gin handlers and routes
└── infrastructure/
    ├── docker-compose.yml
    ├── init-databases.sql
    └── keys/                  # RSA keys (dev only, gitignored)
```

## Dev Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SKIP_ATTEST` | `true` | Skip device attestation in dev |
| `SKIP_SIGNING` | `true` | Skip request signing in dev |
| `KYC_MOCK` | `true` | Use mock KYC responses |
| `APP_ENV` | `development` | Environment mode |

## MailHog (Dev Email)

All emails in development are captured by MailHog. View them at:
**http://localhost:8025**
