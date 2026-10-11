# API

## Overview

REST over HTTP with JSON payloads. The local base URL is `http://localhost:${SERVER_PORT}`; application routes use `/api/v2`. The unversioned public endpoints are `GET /`, `GET /home`, and `GET /health`.

The URL prefix is the versioning strategy. The generated Swagger metadata identifies the contract as version `2.0`.

## Documentation Tooling

The repository contains generated OpenAPI 2.0 documents at [`docs/swagger.json`](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/docs/swagger.json) and [`docs/swagger.yaml`](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/docs/swagger.yaml). In development, Swagger UI is served at `http://localhost:${SERVER_PORT}/swagger/index.html` when `ENABLE_SWAGGER=true`; it is disabled in production.

## Authentication

Protected routes require a JWT access token:

```http
Authorization: Bearer <access_token>
```

The access token is returned by `POST /api/v2/auth/login`, or after completing the two-factor flow when required. Refresh tokens are exchanged at `POST /api/v2/auth/refresh`. Registration, activation, email verification, two-factor verification, and password-reset endpoints are public. The health and home endpoints are also public.

Route authorization is role-based. The supported route roles are `customer`, `veterinarian`, `receptionist`, and `admin`.

## Request / Response Conventions

Successful application responses use this envelope. The unversioned home and health endpoints return their own small status payloads.

```json
{
  "success": true,
  "data": {},
  "message": "Request completed",
  "timestamp": "2026-10-09T12:00:00Z"
}
```

List endpoints that support pagination use page-based query parameters. `page` defaults to `1`, `page_size` defaults to `10`, and `page_size` is capped at `100` in the shared pagination type.

```http
GET /api/v2/appointments?page=2&page_size=20
```

Paginated responses place the page metadata in `meta`:

```json
{
  "success": true,
  "data": [],
  "meta": {
    "total_count": 42,
    "total_pages": 3,
    "current_page": 2,
    "page_size": 20,
    "sort_direction": "ASC",
    "has_next_page": true,
    "has_previous_page": true
  },
  "timestamp": "2026-10-09T12:00:00Z"
}
```

## Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/`, `/home`, `/health` | Home and health information | none |
| POST | `/api/v2/auth/register`, `/login`, `/activate`, `/verify-email` | Register and activate an account or sign in | none |
| POST | `/api/v2/auth/refresh`, `/2fa/verify`, `/reset-password/request`, `/reset-password/reset` | Refresh credentials or complete account recovery | none |
| POST | `/api/v2/auth/logout`, `/logout-all`, `/2fa/enable`, `/2fa/disable` | Manage authenticated sessions and two-factor authentication | required |
| GET, PUT | `/api/v2/profile/` | Read or update the authenticated profile | required |
| GET, POST, PUT, DELETE | `/api/v2/users/`, `/api/v2/customers/`, `/api/v2/employees/` | Manage users and clinic identities | role-scoped |
| GET, POST, PUT, DELETE | `/api/v2/me/appointments/`, `/api/v2/employees/appointments/`, `/api/v2/appointments/` | Manage customer, veterinarian, and administrative appointments | role-scoped |
| GET, POST, PUT, DELETE | `/api/v2/pets/`, `/api/v2/me/pets/` | Manage clinic and owner pet records | role-scoped |
| GET, POST, PUT, DELETE | `/api/v2/addresses/`, `/api/v2/me/addresses/` | Manage user addresses | required or role-scoped |
| GET, POST, PUT, DELETE | `/api/v2/payments/`, `/api/v2/me/payments/` | Track and manage payment records | role-scoped |
| GET, POST, PUT, DELETE | `/api/v2/medical/`, `/api/v2/me/medical/` | Manage or view clinical sessions and related records | role-scoped |
| GET, POST | `/api/v2/notifications`, `/api/v2/me/notifications` | Read notifications or send one as staff | role-scoped |

The generated OpenAPI documents contain the complete operation list, parameters, request bodies, and response schemas.

## Error Handling

Application errors use the same envelope with `success: false`:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "type": "validation",
    "message": "Invalid request",
    "details": {
      "email": "must be a valid email address"
    }
  },
  "timestamp": "2026-10-09T12:00:00Z"
}
```

Protected routes return `401` for missing or invalid bearer credentials and `403` when the authenticated role is not allowed to use the route.
