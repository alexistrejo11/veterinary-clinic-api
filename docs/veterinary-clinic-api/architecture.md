# Architecture

## Overview

This is a modular monolith with a hexagonal layout. One API process contains the domain and application modules, while adapters connect them to HTTP, PostgreSQL, Redis, and notification providers.

## Architecture Type

Hexagonal Architecture (Ports and Adapters).

**Why:** Clinic rules need to remain independent from HTTP handlers and storage details. Repository and service interfaces allow the core workflows to be tested without starting the web server or database, while adapters can change independently.

## Layers

- **HTTP delivery** — parses and validates requests, applies authentication and role policies, and serializes responses. It must not contain business rules.
- **Core** — contains domain entities, value objects, commands, queries, application services, and ports. It must not depend on the HTTP framework or database implementation.
- **Infrastructure** — implements persistence, token and session storage, two-factor support, notification sending, and HTTP adapters. It must implement core contracts rather than redefine domain policy.
- **Configuration and bootstrap** — loads settings, initializes external clients, wires modules, and manages process startup and shutdown.

## Request Flow

```text
Client
  -> HTTP router and global middleware
  -> JWT authentication and role middleware
  -> handler and request validation
  -> core application service
  -> repository or integration port
  -> PostgreSQL, Redis, SMTP, or Twilio
  -> standardized JSON response
```

## Domain Modules

- **Auth and users** — account lifecycle, credentials, sessions, two-factor authentication, roles, and user administration.
- **Customers, employees, and addresses** — clinic identities and contact information.
- **Pets and appointments** — pet records and customer, veterinarian, and administrative scheduling workflows.
- **Medical** — clinical sessions, treatments, prescriptions, attachments, services, and catalogs.
- **Payments and notifications** — payment lifecycle records and user messaging.

## Key Design Decisions

### Hexagonal architecture over handler-driven CRUD

**Rationale:** Keeps scheduling, authorization, and clinical rules independent from delivery and persistence concerns.

**Tradeoff:** Adds interfaces, mapping, and application-service layers to simple CRUD operations.

**Reference:** [internal/core](https://github.com/alexisTrejo11/veterinary-clinic-api/tree/main/internal/core)

### PostgreSQL constraints for scheduling integrity

**Rationale:** Database-level constraints protect veterinarian time ranges and hourly capacity even when concurrent requests reach the service.

**Tradeoff:** Some business behavior lives in migrations and requires database-aware tests and deployments.

**Reference:** [initial schema migration](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/database/migrations/000001_initial_schema.up.sql)

### Redis-backed refresh sessions

**Rationale:** Access-token validation remains stateless while refresh sessions can be revoked for one user or all of a user's devices.

**Tradeoff:** Logout and refresh operations depend on Redis availability and correct expiration configuration.

**Reference:** [Redis session repository](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/session/redis_repository.go)

### Generated SQL access with sqlc

**Rationale:** SQL remains explicit while generated types catch mismatches between queries and Go code during builds and CI.

**Tradeoff:** Schema or query changes require regeneration and review of checked-in generated code.

**Reference:** [sqlc configuration](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/database/sqlc.yaml)

## Project Tree

```text
cmd/
  api/                 API process entry point
  migrate/             migration runner
database/
  migrations/          PostgreSQL schema migrations
  queries/             sqlc query definitions
  sqlc/                generated Go query code
internal/
  config/              configuration and dependency bootstrap
  core/                domain and application modules
  infrastructure/      HTTP, persistence, token, session, and notification adapters
  middleware/          CORS, authentication, audit, and rate limiting
  shared/              common types, errors, logging, HTTP responses, and pagination
docs/
  swagger.json         generated API contract
  swagger.yaml         generated API contract
```
