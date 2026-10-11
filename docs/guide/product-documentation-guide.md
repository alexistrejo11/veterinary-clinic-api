# Product Documentation Guide

This guide defines the structure for `product.md`. Unlike the other guides, this one is **transversal** — it applies to every `ProjectType` (`backend`, `frontend`, `infra`, `fullstack`), not just one.

---

## Purpose

`product.md` is meant to be the **source of truth** for the project: the definition that everything else — architecture, features, code — is a *consequence* of, not the other way around. It's also the entry point for someone non-technical who wants to understand what a project is actually for, without reading `architecture.md` or any code.

Write it in plain language. No stack names, no framework references, no implementation detail — that's what the other files are for.

`product` is optional on every type. In `backend` projects it will usually be present, since most backend services have a real domain worth defining. In `infra` projects it will often be skipped — omit it rather than forcing a shallow version, per the general rule shared by this whole documentation set.

---

## Universal skeleton (same in every type)

**1. Problem / Vision**
What this solves, and for whom. 2-3 sentences, no stack mentions.

```
## Problem / Vision

Pharmacies currently manage inventory and orders through disconnected spreadsheets.
This platform gives them a single system to track stock, take orders, and fulfill them,
without needing separate tools for each step.
```

**2. Users / Actors**
Who interacts with this and what they need from it. Content varies by type (see below), but the question is always the same: who is this for, and what do they come here to do.

**3. Scope**
What's explicitly in and out. This is what prevents half-documented feature creep.

```
## Scope

**In scope:** inventory tracking, order intake, basic fulfillment status.
**Out of scope:** payment processing (handled by a separate service), delivery logistics.
```

---

## Central block — varies by `type`

This is the part that changes shape depending on what kind of project this is. Pick the one matching the `type`, and only include it if there's real, honest content.

### `backend` → Domain Model

Entities, relationships, and invariants — described conceptually, not as an ERD or table schema.

```
## Domain Model

An **Order** belongs to a **Customer** and contains one or more **Items**.
An Order cannot receive new Items once it has been dispatched.
A **Customer** can have multiple Orders, but only one active Cart at a time.
```

Include a **Business Rules / Invariants** subsection if there are enough of them to warrant separating from the model description above:

```
## Business Rules

- An Order total must always equal the sum of its Item subtotals.
- A Customer cannot check out with an empty Cart.
- Inventory cannot go negative — an Order that would cause this is rejected at checkout.
```

### `frontend` → UX & Information Architecture

Not business logic — the product surface. What screens/pages exist, what each does in one line, and the key user journeys. No wireframes, no visual detail, no component names.

```
## UX & Information Architecture

**Pages:**
- **Catalog** — browse and filter available products.
- **Cart** — review selected items before checkout.
- **Checkout** — 3-step flow: cart review → address → payment.

**Key journey — Checkout:**
The user can navigate back to a previous step without losing already-entered data.
Payment step is only reachable once a valid address is set.
```

### `infra` → Capabilities / Service Model

The rarest and most optional of the four — only write this if the infra project genuinely has "users" (other teams/engineers) and offers something with real guarantees, rather than being fully self-explanatory from `infrastructure.md` alone.

What it provides, to whom, and under what conditions:

```
## Capabilities

This pipeline provides deployment automation for backend services.

**Provides:** build, test, and deploy to production on merge to `main`.
**Guarantees:** deploy completes in under 10 minutes; automatic rollback if the
health check fails 3 consecutive times.
**Consumers:** any repository in the `backend-*` family that opts in via the shared workflow.
```

### `fullstack` → Domain Model + UX & Information Architecture

Not a new section — use both of the above together, since a fullstack project genuinely has both layers.

---

## Optional section (any type)

**Product Rules** *(frontend/fullstack equivalent of Business Rules)*
User-facing validation rules or feature flags worth documenting conceptually — only if there's something real, not a restatement of form validation already obvious from the UI.

---

## Checklist

- [ ] No stack names, framework references, or code paths anywhere in this file.
- [ ] Problem/Vision explains the "why," not the "how."
- [ ] The central block matches the project's `type` (Domain Model / UX & IA / Capabilities, or both for fullstack).
- [ ] Scope explicitly lists what's out, not just what's in.
- [ ] This file was skipped entirely (and `product` left unset) if there was nothing honest to put in it — not padded to exist.
- [ ] A non-technical reader could understand this file without opening `architecture.md`.
