# API Documentation Guide

This guide defines the structure to follow when writing `api.md` for a project that exposes a programmatic interface (REST, GraphQL, gRPC, etc.).

`api.md` documents the **contract** — how to talk to this project from the outside. It is not implementation detail, and it is not a live testing tool.

---

## Scope: how this differs from other files

It's easy to let this overlap with other files. Keep the boundary clear:

- **`api.md`** — the external contract. Auth, request/response shape, endpoint list. Written for someone who has never seen the code and wants to make a request.
- **`features.md`** — internal technical capabilities and how they're implemented. If a feature has an externally visible effect on the contract (e.g. a specific error code from rate limiting), that effect goes in `api.md`; the implementation (Redis, sliding window, etc.) stays in `features.md`.
- **`architecture.md`** — how the pieces are organized internally. Never referenced from `api.md`.
- **The API tester (`apiExplorer`)** — a separate, decoupled tool for actually running requests. `api.md` is documentation to *read*, not a live console. Don't assume the reader has access to it.

---

## Applicability

`api.md` isn't backend-exclusive. Write one for any project — backend, fullstack, or even an `infra`/`frontend` project that exposes something like a gateway or BFF — as long as there's a real programmatic interface behind it. If a project doesn't expose one, skip the file and leave `apiDocs` unset.

---

## Suggested structure

**1. Overview**
Protocol (REST / GraphQL / gRPC / mixed), base URL, versioning strategy if one exists (`/v1/` prefix, header-based, none).

```
## Overview

REST over HTTPS, JSON payloads. Base URL: `https://api.example.com/v1`.
Versioning via URL prefix; no breaking changes are made within a major version.
```

**2. Documentation Tooling**
If there's Swagger UI, Redoc, or a Postman collection, link to it — real, reachable URL only. If none of these exist, say so and skip to the endpoint table below; don't imply tooling that isn't there.

**3. Authentication**
The single most important section — this is what someone needs before they can try anything.

- Mechanism: JWT, OAuth2, API key, session cookie, etc.
- How to send it, with a literal example:

```
Authorization: Bearer <token>
```

- Where the token comes from (login endpoint, OAuth flow, issued manually, etc.)
- Any per-endpoint exceptions (public endpoints that skip auth).

**4. Request / Response Conventions**
- **Envelope/wrapper**, if the API wraps responses in a standard shape (`{ data, error, meta }`), shown with a real example.
- **Pagination format** — cursor or offset-based, with a real example of both the request query params and the response shape.

```
GET /v1/orders?cursor=eyJpZCI6MTIzfQ&limit=20

{
  "data": [ ... ],
  "meta": { "nextCursor": "eyJpZCI6MTQzfQ", "hasMore": true }
}
```

**5. Endpoints**
A flat table — no payload detail (that's what Swagger/Redoc is for, if it exists; if it doesn't, this table is the only reference, so keep it accurate and current).

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/v1/orders` | List orders for the authenticated user | required |
| GET | `/v1/orders/{id}` | Get a single order | required |
| POST | `/v1/orders` | Create an order | required |
| GET | `/v1/products` | List public product catalog | none |

**6. GraphQL Schema** *(only if applicable)*
Same high-level treatment — queries/mutations with one line each, not the full schema.

```
## GraphQL

| Operation | Type | Description |
|-----------|------|-------------|
| `order(id)` | Query | Fetch a single order |
| `orders(cursor, limit)` | Query | Paginated list of orders |
| `createOrder(input)` | Mutation | Create a new order |
```

**7. Error Handling** *(optional)*
Standard error shape if one exists, and any status codes that aren't self-explanatory (e.g. a `409` used for a specific business conflict, not just generic REST semantics).

---

## Checklist

- [ ] Auth section includes a literal request example (header/token format), not just the mechanism name.
- [ ] Pagination format shown with a real example, not described abstractly.
- [ ] Endpoint table has method + path + one-line description — no payload bodies.
- [ ] Documentation tooling links (Swagger/Redoc/Postman) are real and reachable, or the section says there is none.
- [ ] Nothing here duplicates implementation detail already covered in `features.md`.
- [ ] If GraphQL is used, the operation list is high-level only — no full schema dump.
