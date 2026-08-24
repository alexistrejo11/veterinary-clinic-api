# Veterinary Clinic API

REST API for veterinary clinic operations—customers, pets, appointments, clinical sessions, payments, and notifications.

[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.10-00ADD8)](https://github.com/gin-gonic/gin)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12+-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Status](https://img.shields.io/badge/status-stable-brightgreen)]()

---

## Table of contents

- [About](#about)
- [Features](#features)
- [Documentation](#documentation)
- [Tech stack](#tech-stack)
- [Architecture at a glance](#architecture-at-a-glance)
- [Prerequisites](#prerequisites)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [API overview](#api-overview)
- [Project structure](#project-structure)
- [Deployment](#deployment)
- [Testing](#testing)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)
- [Links](#links)

---

## About

Veterinary practices often juggle reception, clinical records, pet owner communication, and billing across disconnected tools. This project provides a **single Go API** with hexagonal architecture, type-safe PostgreSQL access via **sqlc**, and **Redis-backed JWT** sessions for a future client portal and staff apps.

Production data plane targets **Amazon RDS (PostgreSQL)** and **Upstash Redis**; the API ships as a **Docker** image ready for AWS ECS or EC2 behind an ALB.

| | |
|---|---|
| **Version** | 2.0.0 |
| **Status** | stable |
| **Primary API prefix** | `/api/v2` (most routes) |
| **Health check** | [GET /health](https://api.vet-clinic.example.com/health) (local: http://localhost:8000/health) |
| **OpenAPI (Swagger)** | http://localhost:8000/swagger/index.html when `ENABLE_SWAGGER=true` and not production |

---

## Features

- **JWT authentication & RBAC** — Register, activate, login, refresh, logout, 2FA hooks, password reset ([details](https://alexis-trejo.com/projects/veterinary-clinic-api/features))
- **Appointment scheduling** — Customer requests, employee assignments, admin search; overlap and hourly capacity rules ([details](https://alexis-trejo.com/projects/veterinary-clinic-api/features))
- **Customers & employees** — Clinic CRUD with soft-delete and restore ([details](https://alexis-trejo.com/projects/veterinary-clinic-api/features))
- **Docker & migrations** — One-command local stack with isolated PostgreSQL and Redis ([details](https://alexis-trejo.com/projects/veterinary-clinic-api/infrastructure))

See [Project Features](https://alexis-trejo.com/projects/veterinary-clinic-api/features) for the full breakdown (pets, medical, payments status included).

---

## Documentation

Project documentation lives on the portfolio site: [veterinary-clinic-api](https://alexis-trejo.com/projects/veterinary-clinic-api).

### Documentation index

| Document | What you will find |
|----------|-------------------|
| [Overview](https://alexis-trejo.com/projects/veterinary-clinic-api/overview) | Problem, solution, metrics, stack, links |
| [Architecture](https://alexis-trejo.com/projects/veterinary-clinic-api/architecture) | Layers, patterns, diagram, data flows |
| [Features](https://alexis-trejo.com/projects/veterinary-clinic-api/features) | Feature cards, snippets, status per area |
| [Infrastructure](https://alexis-trejo.com/projects/veterinary-clinic-api/infrastructure) | Docker, RDS, Upstash, AWS deploy target |
| [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) | Endpoints, auth, rate limits, examples |
| [Code](https://alexis-trejo.com/projects/veterinary-clinic-api/code) | Curated code examples from the repo |

Local OpenAPI: `docs/swagger.json` / Swagger UI when enabled.

---

## Tech stack

- **Go 1.26.5** — monolith binary (`clinic-vet-api`)
- **Gin** — HTTP router and middleware
- **PostgreSQL** — pgx pool + **sqlc** generated queries
- **Redis** — go-redis (JWT / sessions)
- **golang-migrate** — `database/migrations/`
- **Swagger (swaggo)** — handler annotations → `docs/docs.go`
- **Twilio** & **SMTP** — notifications (configurable)
- **Docker** — multi-stage Alpine image (`Dockerfile`)

---

## Architecture at a glance

Hexagonal layout: `internal/core` (domain), `internal/infrastructure` (HTTP, persistence), `internal/config` (bootstrap). Stateless API tasks can scale behind a load balancer; JWT validation does not require sticky sessions.

```mermaid
flowchart LR
  Client[Clients] --> ALB[ALB / TLS]
  ALB --> API[Go API in Docker]
  API --> RDS[(RDS PostgreSQL)]
  API --> Redis[(Upstash Redis)]
  API --> Notify[SMTP / Twilio]
```

Full diagram, layers, and decisions: [Architecture](https://alexis-trejo.com/projects/veterinary-clinic-api/architecture).

---

## Prerequisites

- **Go 1.26.5+**
- **Docker** & Docker Compose (recommended for local stack)
- **PostgreSQL** and **Redis** (created automatically by Docker Compose for local development)

---

## Quick start

### 1. Clone and configure

```bash
git clone https://github.com/alexisTrejo11/veterinary-clinic-api.git
cd veterinary-clinic-api
cp .env-example .env
# Set JWT_SECRET (32+ chars) and other application settings.
# Docker Compose supplies local PostgreSQL and Redis URLs automatically.
```

### 2. Local full stack (Postgres + Redis + API)

From the repository root:

```bash
docker compose up --build
docker compose up --build -d
```

| Service | URL |
|---------|-----|
| API | http://localhost:8000 |
| Health | http://localhost:8000/health |
| Swagger | http://localhost:8000/swagger/index.html (dev) |
| Postgres | Available only to the API on the Compose network |
| Redis | Available only to the API on the Compose network |

Migrations run on API startup unless `SKIP_MIGRATIONS=true`.

### 3. Run on the host (without Docker)

```bash
# Start Postgres and Redis yourself, then:
export $(grep -v '^#' .env | xargs)
go run ./cmd/migrate    # optional: migrations only
go run ./cmd/api
```

## Configuration

Copy [.env-example](.env-example) to `.env`. Never commit `.env`.

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL host URL (`jdbc:postgresql://` prefix optional) |
| `DATABASE_USER` / `DATABASE_PASSWORD` | Credentials merged into connection |
| `REDIS_URL` | Redis URL (`redis://` or `rediss://` for TLS) |
| `JWT_SECRET` | Signing key (minimum 32 characters) |
| `JWT_EXPIRATION_TIME` / `REFRESH_TOKEN_EXPIRY` | Token lifetimes |
| `SERVER_PORT` / `SERVER_HOST` | HTTP bind (default `8000` / `0.0.0.0`) |
| `ENVIRONMENT` | `development` or `production` |
| `ENABLE_SWAGGER` | Swagger UI (disabled in production by default) |
| `CORS_ALLOW_ORIGINS` | Allowed frontend origins |
| `RATE_LIMIT_*` | Global rate limiter toggles |
| `SMTP_*` / `TWILIO_*` | Email and SMS notifications |

RDS hosts (`*.rds.amazonaws.com`) get `sslmode=require` automatically when parsing `DATABASE_URL`.

---

## API overview

Base URL (local): `http://localhost:8000`  
Base URL (production placeholder): `https://api.vet-clinic.example.com`

Authentication: `Authorization: Bearer <access_token>` on protected routes.

| Area | Base path | Doc |
|------|-----------|-----|
| Health | `GET /health` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Auth (public) | `/api/v2/auth/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Auth (session) | `/auth/` (logout, refresh, 2FA) | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Profile | `/api/v2/profile/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Users | `/users/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Customers | `/api/v2/customers/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Employees | `/api/v2/employees/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Appointments | `/api/v2/me/appointments/`, `/api/v2/appointments/`, … | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Notifications | `/api/v2/me/notifications`, `/api/v2/notifications` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Medical | `/api/v2/medical/`, `/api/v2/me/medical/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Pets | `/api/v2/pets/`, `/api/v2/me/pets/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Payments | `/api/v2/payments/`, `/api/v2/me/payments/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |
| Addresses | `/api/v2/addresses/`, `/api/v2/me/addresses/` | [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api) |

Full endpoint list with request/response examples: [API](https://alexis-trejo.com/projects/veterinary-clinic-api/api).

---

## Project structure

```
veterinary-clinic-api/
├── cmd/api/                     # API entrypoint
├── cmd/migrate/                 # Host-side migration runner
├── internal/
│   ├── core/                    # Domain (auth, pets, appointments, …)
│   ├── infrastructure/http/   # Handlers, router, DTOs
│   ├── infrastructure/persitence/
│   ├── config/                  # Settings, Redis, bootstrap
│   └── middleware/              # Auth, CORS, audit, rate limit
├── database/migrations/         # SQL migrations
├── database/sqlc/               # Generated query code
├── docs/                        # Swagger bundle (docs.go, swagger.*)
├── Dockerfile                   # API image build
├── docker-compose.yml           # Self-contained local Compose stack
├── scripts/                     # entrypoint, URL parsers
├── .env-example
└── go.mod
```

---

## Deployment

**Live today:** RDS PostgreSQL and Upstash Redis via `DATABASE_URL` and `REDIS_URL`.

**Next step (documented, not automated in repo):** Build `Dockerfile` → push to ECR → run on **ECS Fargate** behind an **ALB** with health check `GET /health`.

Checklist and cost placeholders: [Infrastructure](https://alexis-trejo.com/projects/veterinary-clinic-api/infrastructure).

```bash
# Local API, PostgreSQL, and Redis
docker compose up --build -d
```

---

## Testing

```bash
go test ./...
```

Add focused tests under the same package as the code under test (`internal/...`).

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-change`)
3. Commit with clear messages
4. Open a pull request against `main`

Keep changes focused; match existing patterns in `internal/core` and handlers.

---

## Security

- Store secrets in `.env` or a secrets manager—never in git.
- Use a strong `JWT_SECRET` (32+ characters) and restrict `CORS_ALLOW_ORIGINS` in production.
- Do not run demo seed migration `000007_insert_demo_data` on production databases.

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/alexisTrejo11/veterinary-clinic-api/security/advisories/new) for this repository.

---

## License

No license file is included yet. Add a `LICENSE` file before distributing or publishing the package.

---

## Links

| Resource | URL |
|----------|-----|
| Repository | https://github.com/alexisTrejo11/veterinary-clinic-api |
| **Documentation hub** | https://alexis-trejo.com/projects/veterinary-clinic-api |
| Overview | https://alexis-trejo.com/projects/veterinary-clinic-api/overview |
| API | https://alexis-trejo.com/projects/veterinary-clinic-api/api |
| Health (placeholder) | https://api.vet-clinic.example.com/health |


# Run Migration manually

```bash
migrate -path database/migrations -database "postgresql://user:password@localhost/vet_database?sslmode=disable" up
```
