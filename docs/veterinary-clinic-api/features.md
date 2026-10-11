# Features

## Identity and Access

### JWT Access Tokens with Redis-Backed Sessions

Login issues JWT access and refresh tokens. Redis stores refresh-session state and short-lived verification, password-reset, and two-factor challenge tokens so sessions can be refreshed or revoked.

**Reference:** [token service](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/token/token_service.go), [session repository](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/session/redis_repository.go)

### Role-Based Route Policies

Authentication middleware validates bearer tokens and places user identity and role claims in the request context. Route groups then restrict operations for customers, veterinarians, receptionists, and administrators.

**Reference:** [auth middleware](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/middleware/auth_middleware.go), [router](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/http/router.go)

## Clinic Operations

### Appointment Scheduling Integrity

Customers can request and cancel appointments, veterinarians manage assigned appointments, and staff administer the schedule. Scheduling constraints prevent overlapping veterinarian bookings and limit active appointments within an hour.

**Reference:** [appointment routes](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/http/router.go), [schema migration](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/database/migrations/000001_initial_schema.up.sql)

### Medical Records and Clinical Catalogs

Medical staff manage sessions and their vaccinations, surgeries, prescriptions, attachments, services, and catalog entries. Customers receive read-only access to their own pets' clinical history.

**Reference:** [medical core module](https://github.com/alexisTrejo11/veterinary-clinic-api/tree/main/internal/core/medical)

### Payment Lifecycle Tracking

Staff can create, update, process, cancel, and refund payment records, while customers can read their own payment history. Database rules validate amounts and completed-payment fields.

**Reference:** [payment handler](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/infrastructure/http/handlers/payment.go)

## Cross-Cutting Concerns

### Standard Response Envelope and Page-Based Pagination

Handlers share a JSON response structure containing success state, data or error details, timestamps, request metadata, and page metadata where applicable. Pagination accepts page and page-size parameters with a maximum page size of 100.

**Reference:** [HTTP response type](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/shared/http/response.go), [pagination type](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/shared/page/page.go)

### Audit Logging and Configurable Rate Limiting

Requests are audit logged, CORS is configured from application settings, and rate limiting can be enabled through configuration. The current limiter uses an in-memory sliding-window store.

**Reference:** [API bootstrap](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/cmd/api/main.go), [rate limiter](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/internal/middleware/rate_limiter.go)

### Generated OpenAPI Contract

Handler annotations generate the checked-in Swagger JSON and YAML contract. CI regenerates both artifacts and fails if generated output differs from the repository.

**Reference:** [Swagger files](https://github.com/alexisTrejo11/veterinary-clinic-api/tree/main/docs), [CI workflow](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/.github/workflows/ci-cd.yml)
