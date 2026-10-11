# Backend Project Documentation Guide

This guide defines the structure to follow when writing `architecture.md`, `features.md`, and `infrastructure.md` for a backend project. It is intentionally plain Markdown, with no custom components or app-specific syntax, so it can be produced independently and plugged in later.

> `product.md` and `api.md` are **not** covered here — they have their own guides (Product Documentation Guide, API Documentation Guide) since both apply beyond backend projects. This guide only covers the three files that are backend-specific in nature.

Only include a subsection if there is real, honest content for it. Skip subsections that don't apply — do not fill them with generic or aspirational content.

---

## 1. architecture.md

Answers: **how are the pieces organized, and how do they connect?**

### Suggested structure (in this order)

**1. Overview**
1-2 sentences, specific to the architecture (not a repeat of the project's general overview). Example: "Hexagonal architecture with 4 layers, event-driven communication between services."

**2. Architecture Type**
Name the pattern (hexagonal, clean, layered, onion, event-driven, etc.). Follow it with a short "Why" note explaining when/why this pattern was chosen for this specific project — not a generic definition.

```
## Architecture Type

Hexagonal Architecture (Ports & Adapters).

**Why:** Chosen because the domain logic needs to stay independent from delivery
mechanisms (REST today, possibly message queues later). Keeps core business
rules testable without spinning up infrastructure.
```

**3. Layers**
Short list, 1-2 lines per layer: its responsibility, and what it must NOT know about.

```
## Layers

- **Presentation** — handles HTTP input/output, validation, serialization. Has no business logic.
- **Application / Service** — orchestrates use cases, calls domain and infrastructure through interfaces.
- **Domain** — core business rules. No framework or infrastructure dependencies.
- **Infrastructure** — persistence, external services, messaging. Implements the interfaces defined by the domain.
```

**4. Communication / Request Flow**
A flow diagram, written as a fenced code block (plain text/ASCII, or mermaid syntax if the target supports it). Example:

```
## Request Flow

```
Request → Presentation → Service → Domain → Infrastructure → Response
```
```

If the project has multiple services, add a second diagram showing how they communicate with each other (sync/HTTP, async/events).

**5. Domain Modules** *(optional — only if the project has clear bounded contexts)*
Short list of business modules (auth, payment, inventory...) with one line each on what it covers. Skip entirely for small projects without a clear domain split.

**6. Key Design Decisions**
2-4 decisions max. Prioritize quality over quantity — two well-explained decisions beat five shallow ones. Each decision must include a rationale AND a tradeoff (not just the upside). Use this format:

```
## Key Design Decisions

### Hexagonal over classic MVC

**Rationale:** Keeps the domain isolated from framework concerns, makes unit
testing the business logic straightforward without mocking the framework.

**Tradeoff:** More boilerplate for simple CRUD operations — interfaces and
adapters add ceremony even where a direct implementation would be shorter.

**Reference:** https://github.com/{user}/{repo}/tree/main/src/domain
```

**7. Project Tree**
A fenced code block with the real folder structure. This can go last — it's a quick-reference, not narrative.

### Checklist
- [ ] Does it explain *why* the pattern was chosen, not just its name?
- [ ] Is the flow diagram understandable without reading the code?
- [ ] Does every decision include both a rationale and a tradeoff?

---

## 2. features.md

Answers: **what technical capabilities were implemented, as named, self-contained pieces?**

Litmus test: if it can be explained in 3-4 lines without needing the architecture diagram, it belongs here. If it needs the diagram to make sense, it belongs in architecture.md instead.

If a feature has an externally visible effect on the API contract (a specific header, error code, rate-limit behavior), document that effect in `api.md` — this file stays focused on the internal implementation.

### Suggested structure

Group features into categories if there are several (optional — only useful once the list grows):

- **Core Patterns** — CQRS, Event Sourcing, Query Bus, Command Bus, Repository Pattern, Factory Pattern, etc.
- **Cross-cutting Concerns** — rate limiting, authentication (JWT/OAuth), caching, structured logging, centralized error handling.
- **Integrations** — Kafka/event messaging, PostgreSQL views/materialized views, third-party services, gRPC.

Each feature entry:

```
### Rate Limiting with Redis

Per-user request throttling on write endpoints using a sliding window
counter stored in Redis, to protect the service from abuse.

**Reference:** https://github.com/{user}/{repo}/blob/main/src/middleware/rate-limit.ts
```

- Title must be specific and recognizable to another engineer — not a vague buzzword.
- 2-3 line description: what problem it solves, how it's implemented at a high level.
- Reference link is optional but preferred when there's a concrete file/folder to point to.

### Checklist
- [ ] Is the title specific (not a generic buzzword)?
- [ ] Does each feature link to the real implementation when possible?
- [ ] Does it avoid duplicating something better explained in architecture.md or api.md?

---

## 3. infrastructure.md

Answers: **how does this run and get deployed in the real world?**

### Suggested structure

**1. Overview**
1-2 lines on how the project is deployed overall. Example: "Containerized with Docker, CI/CD via GitHub Actions, deployed to a VPS through a Cloudflare Tunnel."

**2. Deployment Pipeline**
Sequential steps, each with what it does (not just the tool name) and a reference when relevant:

```
## Deployment Pipeline

### 1. Containerization — Docker

Multi-stage build to keep the final image small. Base image: node:20-alpine.

**Reference:** https://github.com/{user}/{repo}/blob/main/Dockerfile

### 2. CI/CD — GitHub Actions

On push to main: runs tests, builds the image, pushes to the registry.

**Reference:** https://github.com/{user}/{repo}/blob/main/.github/workflows/deploy.yml

### 3. Hosting — VPS via Cloudflare Tunnel

Deployed to a self-managed VPS, exposed through a Cloudflare Tunnel
(no public inbound ports opened directly).
```

**3. Infra Stack**
Plain list of tools: Docker, (Kubernetes if applicable), tunnel/hosting provider, other infra tooling.

**4. Monitoring & Observability** *(optional — only if something real is actually running)*
Short description of what is monitored and with what (Grafana, Prometheus, Loki, ELK). Include a link to a public/accessible dashboard if there is one. Do not include this
section aspirationally — only if it's real and running today.

### Checklist
- [ ] Does each pipeline step explain what it does, not just name the tool?
- [ ] Are code references short/relevant snippets or file links, not full files pasted in?
- [ ] Is Monitoring only included if something is actually running and observable?

---

## General rule for all backend files

If a project has no honest, real content for a subsection (e.g. no real monitoring yet, or no interesting architecture decisions to share), **omit that subsection** rather
than filling it with generic content. A short, honest `architecture.md` is better than a long one padded with filler — the quality of the Decisions and Features is what differentiates the project, not the word count.

This applies equally to `api.md` and `product.md` — see their respective guides.
