# API Completion Plan

## Purpose

This document defines the scope and implementation order required to make the
veterinary clinic API functional, secure, and testable. It is based on a
read-only audit of the current codebase performed on 2026-08-17.

The goal is to complete the existing v2 API surface, not reduce it. Versioning
or removal of features can be considered separately after this work is stable.

## Scope

All existing feature areas are in scope:

- Authentication, account activation, sessions, refresh tokens, password
  recovery, and two-factor authentication.
- Users and profiles.
- Customers, employees, addresses, and pets.
- Appointments and lifecycle notifications.
- Payments.
- Medical sessions, extensions, and catalogs.
- Notifications by email and SMS.
- Health checking, API documentation, and route registration.

## Current Delivery Status

The functional implementation work in Phases 1 through 8 is complete. The
following work was completed after the original audit:

- Canonical roles, route policy, auth/session infrastructure, public customer
  registration, activation, payments, notifications, medical routes, Swagger,
  SQLC generation, and CI verification were implemented or repaired.
- Production bootstrap now constructs the health handler; public registration
  creates an inactive customer profile and activation enables it.
- Veterinarian medical reads and mutations are constrained to assigned sessions,
  including child records and scoped aggregate medical queries.
- Payment searches bind documented query parameters; authorized ledger entries
  can be updated and payment appointment/customer relationships are validated.
- CI runs build, unit tests, race tests, vet, migrations, PostgreSQL/Redis smoke
  verification, SQLC generation, and Swagger generation checks.

Phase 9 remains in progress. A Testcontainers PostgreSQL/Redis E2E harness now
constructs the real router with per-test PostgreSQL schemas and deterministic
notification fakes. The registration, activation, login, authenticated-profile,
health, validation, unauthenticated-profile, and customer pet create/read
workflows run through it. Customer appointment request, ownership read,
cancellation, and persisted status transitions are also covered, along with
authenticated customer notification reads. Admin payment creation and
cancellation, including customer payment visibility and ledger persistence, are
covered through role middleware.
Administrative manual email notifications are covered with a persisted staff
fixture and deterministic sender-delivery assertion.
Medical session creation is covered for an administrative clinician fixture,
with veterinarian assignment access and customer read access asserted through
their respective route policies.
Every registered protected operation is also swept without credentials and must
return `401`; the lifecycle suite asserts representative customer/staff role
denials return `403`.
Complete HTTP-to-database coverage for every advertised operation and error case
is still outstanding.

## Canonical Authorization Model

The application must persist and authorize only these roles:

| Role           | Responsibilities                                                                |
| -------------- | ------------------------------------------------------------------------------- |
| `admin`        | User and staff administration; unrestricted operational access.                 |
| `veterinarian` | Assigned appointments and medical records.                                      |
| `receptionist` | Customers, pets, addresses, payments, and appointment administration.           |
| `customer`     | Only their own profile, pets, appointments, medical records, and notifications. |

The legacy route roles `manager` and `employee` must be removed from policy
checks. Their intended capabilities map to `receptionist` and `veterinarian`,
respectively. The database role constraint, existing records, seed data, DTO
validation, and middleware must be migrated consistently.

Public registration must create customer accounts only. Staff and administrator
accounts must be created through privileged administration flows.

## Current Baseline Findings

### Production blockers

- Any anonymous caller can register an active administrator by supplying an
  administrative role during registration.
- Authentication is not functional: bootstrap injects a nil session repository
  and successful login dereferences it.
- Two-factor authentication is not functional: bootstrap injects a nil provider
  and enable/verify flows do not persist state correctly.
- Route middleware uses `manager` and `employee`, neither of which is a valid
  persisted role.
- Health, address, pet, and payment handlers are created but never registered.
- Medical routes are skipped because the medical handler is intentionally nil.
- Customer handlers contain a nil mapper and panic when mapping a customer.

### Incomplete endpoint behavior

- Customer and employee search endpoints bind input but do not query or return
  a response.
- Customer restore hard-deletes data instead of restoring it.
- Employee restore is explicitly unsupported.
- Password-reset requests issue no token and send no message.
- The documented payment endpoint is not registered.
- User and customer deleted-state lookups are no-op implementations.

### Authorization and data defects

- A customer can request an appointment for another customer's pet.
- Employee appointment mutations do not consistently verify assignment.
- Employee and customer update routes trust a body ID instead of the path ID.
- Employee schedules always save Monday, ignoring the requested day.
- Appointment pagination metadata is not based on the requested page and size.
- Token/session revocation failures are ignored in parts of the auth service.

### Testing gap

`go test ./...`, `go test -race ./...`, and `go vet ./...` currently pass, but
there is only one HTTP test. It manually mounts a health endpoint and does not
exercise the production bootstrap, route registration, auth, roles,
repositories, or endpoint behavior.

## Work Phases

### Phase 1: Roles, migrations, and route policy

Status: Completed on 2026-08-17.

Implementation notes:

- Migration `000010_canonical_user_roles` maps `manager` to `receptionist` and
  `employee` to `veterinarian`.
- The migration aborts if `superadmin` or any other unsupported role exists;
  those records require an explicit data decision before migration.
- User and employee lifecycle routes are admin-only. Receptionist/admin routes
  manage appointments and customers; veterinarian/admin routes manage assigned
  appointments and medical records.

1. Add a database migration that reconciles legacy role data and enforces the
   four canonical roles.
2. Replace every `manager` authorization requirement with the correct
   receptionist/admin policy.
3. Replace every `employee` authorization requirement with the correct
   veterinarian/admin policy.
4. Define a route-by-route access matrix, including resource ownership rules.
5. Add middleware and integration tests for the role matrix.

Acceptance criteria:

- No application route checks `manager` or `employee`.
- Database constraints and Go enums agree.
- Every registered route has an explicit allowed-role policy.

### Phase 2: Authentication and external infrastructure

Status: Completed on 2026-08-17.

Implementation notes:

- Redis-backed refresh-token sessions are implemented without storing raw
  refresh tokens in Redis keys.
- TOTP secrets are encrypted with AES-256-GCM using the required base64
  `TWO_FACTOR_ENCRYPTION_KEY`; backup-code hashing is available for the
  enrollment workflow in Phase 3.
- SMTP delivery requires TLS: STARTTLS for standard SMTP ports and implicit TLS
  for port 465. Twilio delivery is wired through the configured sender number.
- Email/SMS two-factor challenges remain part of Phase 3 because the current
  authentication contract has no challenge issuance or delivery lifecycle.

1. Implement a Redis-backed `auth.SessionRepository`.
2. Implement a secure two-factor provider:
   TOTP verification, encrypted or otherwise protected secret storage, hashed
   one-time backup codes, and SMS/email delivery where supported.
3. Implement SMTP and Twilio adapters behind notification interfaces.
4. Make bootstrap inject all required providers and fail clearly if a required
   provider is absent.
5. Ensure Redis lifecycle handling and configuration failures are testable.

Acceptance criteria:

- Valid login produces working access and refresh tokens without panic.
- Sessions can be created, queried, rotated, revoked, and deleted.
- Selecting a configured notification channel never dereferences a nil sender.

### Phase 3: Complete authentication workflows

Status: Completed on 2026-08-17.

Implementation notes:

- Public registration accepts only `customer`, creates a pending account, and
  emails an opaque activation token.
- Activation and password reset consume their Redis-backed tokens; reset also
  verifies the requested email and revokes all active sessions.
- Refresh-token, password-reset, and two-factor verification routes are public
  token-driven flows rather than access-token-protected routes.
- TOTP login uses a five-minute, single-use opaque pre-authentication challenge
  that preserves device metadata without exposing an access token.
- TOTP enrollment stores encrypted secrets and hashed backup codes, returns
  recovery codes only at enrollment, and consumes backup codes after use.
- Email/SMS two-factor challenges remain deferred: their challenge-delivery
  contract is not yet implemented, so enrollment accepts TOTP only.

1. Restrict public registration to customer accounts.
2. Create customer profiles as part of the appropriate registration workflow.
3. Generate, persist, deliver, consume, and revoke activation tokens.
4. Implement email verification or remove the unused separate workflow after
   activation behavior is consolidated.
5. Make refresh token rotation authenticate from the refresh token, not an
   access-token middleware requirement.
6. Implement password-reset token issuance, delivery, validation, consumption,
   and revocation. Reset routes must not require an existing access token.
7. Implement a pre-authentication two-factor challenge. The verification route
   must use that challenge rather than requiring an access token that has not
   yet been issued.
8. Persist two-factor enable/disable state, method, secret, and backup-code
   changes atomically.
9. Propagate revocation and session-deletion failures rather than reporting a
   successful logout or deletion prematurely.

Acceptance criteria:

- Customer registration, activation, login, refresh, logout, logout-all, reset
  password, and account deletion have end-to-end tests.
- Two-factor enrollment, challenge verification, backup-code consumption, and
  disable flows have end-to-end tests.
- An anonymous caller cannot create or activate privileged accounts.

### Phase 4: Register and repair core resource endpoints

Status: Completed on 2026-08-17.

Implementation notes:

- Health, address, pet, and payment route groups are registered under the v2
  API and covered by route-registration tests.
- Customer and employee searches bind query parameters, preserve pagination,
  map responses, and return standard API payloads.
- User, customer, and employee restore semantics are implemented through SQL
  restore/deleted-state queries.
- Customer, employee, and address updates require matching path and body IDs.
- Employee schedule days, pet activation persistence, pet totals, appointment
  pagination metadata, country parsing, and missing handler responses were
  corrected.
- Address default reassignment now happens in the create/update SQL statement,
  avoiding the previous follow-up bulk update state gap.

1. Register health, address, pet, and payment route groups.
2. Correct customer mapper construction.
3. Complete customer and employee search handlers with query binding,
   specifications, pagination, mapping, and API responses.
4. Implement correct user, customer, and employee restore repository/service
   operations.
5. Make URL path IDs authoritative for update routes; reject conflicting body
   IDs where the body retains an ID for compatibility.
6. Ensure all successful handlers return the standard API response.
7. Fix employee schedule day mapping.
8. Fix pagination metadata to match requested pagination and total records.
9. Persist pet activation/deactivation changes.
10. Make default-address updates transactional and complete country validation.

Acceptance criteria:

- Every address, pet, customer, employee, payment, and health operation in the
  chosen API contract is registered and returns a documented response.
- Soft delete and restore preserve data correctly.
- Update requests cannot target a record different from their URL.

### Phase 5: Ownership and appointment enforcement

Status: Completed on 2026-08-17.

Implementation notes:

- Customer appointment requests verify that the pet belongs to the authenticated
  customer in the command service.
- Customers can cancel only their own cancellable appointments through
  `POST /api/v2/me/appointments/:id/cancel`.
- Veterinarian routes expose only assigned-work operations and enforce the
  assigned employee in handlers and command services. Veterinarians cannot
  create, confirm, or self-assign appointments.
- Receptionist/admin routes retain appointment creation, assignment, and
  confirmation. Appointment search excludes soft-deleted records and preserves
  requested pagination.
- The appointment SQLC output was updated with its source query; run `make
sqlc` on a machine with the SQLC binary installed before release.

1. Verify that a customer owns a pet before creating a customer appointment.
2. Verify veterinarian assignment before every veterinarian appointment read or
   mutation.
3. Enforce the same ownership rules in command services, not only Gin handlers.
4. Ensure receptionist/admin appointment operations follow the role matrix.
5. Correct appointment search and pagination behavior.
6. Add appointment lifecycle event publishing for notifications.

Acceptance criteria:

- Cross-customer pet appointment creation is rejected.
- A veterinarian cannot read or mutate another veterinarian's appointment.
- Appointment lifecycle transitions and invalid transitions are covered by
  integration tests.

### Phase 6: Notifications

Status: Completed on 2026-08-18.

Implementation notes:

- Delivery records are created as `pending` before provider delivery. Every
  attempt is then persisted as `sent` or `failed`, with its provider reference,
  timestamp, attempt count, and failure message where applicable.
- Manual sends record the authenticated sender. Customers remain restricted to
  their own notification records; staff monitoring remains role-gated by routes.
- Account activation/reset and appointment lifecycle messages use the durable
  notification service. New customer requests are delivered to active
  receptionists only.

1. Persist notification requests, delivery attempts, status, errors, and
   provider references.
2. Implement notification summaries from persisted data.
3. Wire email and SMS delivery adapters.
4. Send notifications for account activation, password reset, and appointment
   creation, confirmation, cancellation, completion, and no-show events.
5. Enforce that customers can read only their own notifications.

Acceptance criteria:

- Manual notification sending persists a visible delivery record.
- Channel-specific delivery failures are represented consistently.
- Appointment and auth events create the expected notifications.

### Phase 7: Payments

Status: Completed on 2026-08-18.

Implementation notes:

- Payments are explicitly modeled as an internal clinic ledger; no external
  gateway, webhooks, or provider refund behavior is claimed.
- References are idempotency keys, local receipt numbers are created when an
  authorized payment is settled, and cancellation/refund operations require a
  reason.
- Customer payment reads are limited to `/api/v2/me/payments`; staff payment
  management remains restricted to admin and receptionist roles.
- Migration `000012_local_payment_ledger` converts legacy status values and
  must be applied before deployment.

1. Register all existing payment routes and reconcile Swagger documentation.
2. Decide and document the payment boundary before implementation:
   a real external gateway integration, or an internal payment ledger.
3. If using a gateway, add provider transaction identifiers, idempotency,
   provider status handling, refund references, and secure webhook validation.
4. If using an internal ledger, rename/document operations so they do not claim
   to process external payments.
5. Validate refund/cancel amount, reason, reference, state transitions, and
   customer/appointment authorization.

Acceptance criteria:

- Payment documentation accurately describes the implemented boundary.
- Payment state transitions are durable, authorized, idempotent where needed,
  and fully tested.

### Phase 8: Medical records and catalogs

Status: Completed on 2026-08-18.

Implementation notes:

- Concrete application services now support medical sessions, vaccinations,
  surgeries, prescriptions, attachments, session services, and all catalogs.
- Bootstrap constructs the medical repositories, services, and handler, making
  the documented medical route groups reachable.
- Medical operations validate parent sessions and active catalog entries; session
  creation validates related pet/customer and employee records.
- Customer read routes remain constrained to the authenticated customer's pets,
  while veterinarian/admin routes retain medical write access.

1. Implement concrete application services for medical sessions, vaccinations,
   surgeries, prescriptions, attachments, session services, and all catalogs.
2. Use the existing SQLC repositories where they are valid; complete repository
   behavior and transactions where needed.
3. Construct and inject `MedicalHandler` in bootstrap.
4. Register every supported medical route.
5. Enforce veterinarian write permissions and customer ownership for read-only
   pet/session views.
6. Validate cross-entity relationships: session, pet, customer, veterinarian,
   catalog item, and attachment associations.

Acceptance criteria:

- No medical handler/service dependency is nil in production bootstrap.
- All documented medical endpoints are registered, authorized, and tested.
- Customers cannot access another customer's medical information.

### Phase 9: Contract tests and dead-code cleanup

Status: In progress. Route contract, generated-artifact, race, vet, migration,
and PostgreSQL/Redis smoke checks are complete. Isolated auth/profile/health
E2E coverage is in CI; remaining resource lifecycle E2E coverage is outstanding.

Implementation notes:

- The route contract test verifies every Swagger operation is registered by the
  production router; Swagger now documents 151 registered operations.
- Removed obsolete commented bootstrap code and unreachable veterinarian
  appointment creation/confirmation handlers.
- Regenerated Swagger and SQLC artifacts and added reusable `make test-race`
  and `make vet` verification targets.
- Unit, route-contract, race, vet, SQLC, and Swagger checks pass. A disposable
  PostgreSQL/Redis integration environment is still required to validate
  migrations and real provider/database interactions before deployment.

1. Add production bootstrap tests that construct the real router and verify the
   complete expected route list.
2. Add integration tests using a test PostgreSQL database and Redis instance.
3. Add endpoint tests for success, validation, authorization, ownership, and
   failure behavior for every route.
4. Add a Swagger-to-router contract test to prevent documentation drift.
5. Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and static
   analysis in CI.
6. After all supported features are covered, remove unreachable code:
   the commented legacy bootstrap implementation, obsolete helpers, unused
   DTOs/interfaces, stale documentation, and redundant dependencies.

Acceptance criteria:

- Every Swagger operation maps to one working registered route.
- CI executes unit, repository, integration, race, vet, and static-analysis
  checks.
- Remaining code is reachable from a supported feature or deliberately shared.

## End-to-End Test Environment

The remaining Phase 9 endpoint coverage must use an isolated environment that
executes the real router, middleware, services, PostgreSQL schema, SQLC queries,
and Redis session storage. Do not use SQLite or an in-memory SQL substitute:
the production schema relies on PostgreSQL-specific migrations, enums,
constraints, and queries.

### Test Infrastructure

1. Add `testcontainers-go` to start disposable PostgreSQL 16 and Redis 7
   containers for integration test suites.
2. Apply all migrations once when the suite starts.
3. Create a unique PostgreSQL schema for every test, set that test connection's
   `search_path`, and clean up by dropping the schema after the test.
4. Construct the real Gin router and authentication middleware from the test
   dependencies; requests must travel through registered routes rather than
   calling handlers directly.
5. Refactor bootstrap dependencies only as needed to inject deterministic fake
   SMTP and Twilio senders. Fakes record delivery attempts in memory; real
   provider smoke tests remain optional and secret-gated.
6. Keep unit tests Docker-free. E2E tests run with
   `go test -tags=integration ./internal/integration/...` on Docker-enabled
   developer machines and CI runners.

### Shared Test Kit

Create `internal/testkit/` with:

- Container lifecycle, migration, schema creation, and cleanup helpers.
- Real-router construction with fake notification providers.
- Fixtures for every canonical role, customers, employees, pets, appointments,
  payments, medical sessions, and notifications.
- HTTP request, authenticated-client, response-schema, and database-state
  assertion helpers.

### Required Endpoint Coverage

Organize E2E suites by lifecycle instead of one unstructured file per route:

1. Identity: register, activate, login, refresh, logout, password reset, and
   two-factor flows; include privileged-registration rejection.
2. Profiles and administration: profile, users, customers, employees,
   addresses, pets, soft deletion, and restoration.
3. Appointments: customer request/cancel, receptionist/admin lifecycle, and
   veterinarian assigned-work transitions.
4. Payments: search, create, update, process, cancel, refund, idempotency, and
   customer/appointment integrity.
5. Medical: catalogs, sessions, every child resource, customer read-only views,
   veterinarian assignment isolation, and deletion/restoration.
6. Notifications and health: customer visibility, staff monitoring/manual send,
   and bootstrapped health response.

For every advertised endpoint, include a success assertion and the applicable
validation, unauthenticated, wrong-role, ownership, missing/deleted-resource,
and persistence-state assertions.

## Dependency Order

The phases must generally be performed in this order:

1. Roles and database migration.
2. Auth/session/provider infrastructure.
3. Auth workflows.
4. Core routes and resource correctness.
5. Appointment ownership.
6. Notifications.
7. Payments.
8. Medical services and routes.
9. Comprehensive tests and dead-code removal.

Medical work is the largest implementation area because its HTTP contracts and
repositories exist, but most concrete application services are missing. Auth
infrastructure is the first operational blocker because successful login cannot
currently complete safely.

## Non-Goals for This Plan

- Changing the public API version.
- Removing advertised feature groups solely because they are incomplete.
- Introducing backward-compatibility behavior beyond what is necessary for
  existing database role records and deployed consumers.
