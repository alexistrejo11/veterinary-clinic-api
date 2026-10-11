# Project Documentation Guide

This guide explains the **schema and conventions** used to document a repository for the project explorer (the `docs/` folder with `projects.model.ts` and `projects.ts`).

Use this guide every time you wire up a new repo. It is intentionally generic so it can be copied across all of your projects.

This is the entry point of the documentation set. The other guides — Backend Project Documentation Guide, API Documentation Guide, and Product Documentation Guide — go into detail on each individual file referenced below.

---

## File structure

In the repository root, create the following:

```
docs/
  projects.model.ts   # TypeScript interfaces — copy from docs/projects.model.ts in drugstore-platform
  projects.ts         # The PROJECTS array — your real data
  {slug}/
    product.md         # optional — see Product Documentation Guide
    architecture.md     # see Backend Project Documentation Guide
    features.md
    infrastructure.md
    api.md              # optional — see API Documentation Guide
```

- `projects.model.ts` is identical across all repos. Copy it, do not edit.
- `projects.ts` is the only file you should actually author.
- Only create the Markdown files for sections you actually flip on via the booleans below. A missing file for a `true` flag is a bug; a missing flag for a file you don't have is correct.

---

## The schema (reference)

```ts
export type ProjectType = 'backend' | 'frontend' | 'infra' | 'fullstack';
export type ProjectStatus = 'production' | 'development' | 'archived';

export interface ProjectModel {
  slug: string;           // url-safe id, e.g. 'drugstore-platform'
  name: string;           // display name, e.g. 'Drugstore Platform'
  logo?: string;          // '/logos/<name>.svg'
  type: ProjectType;
  description: string;    // one or two sentences, English, no marketing fluff
  status: ProjectStatus;
  docs?: ProjectDocsModel;
  metadata: ProjectMetadataModel;
  services?: ProjectModel[];  // nested microservices/submodules
}
```

`ProjectDocsModel` controls which sections are rendered. `overview` is typed data; the rest map to Markdown files (or, in the case of `apiExplorer`, to an external tool — see below):

```ts
export interface ProjectDocsModel {
  overview?: ProjectOverviewData;   // typed payload (below)
  product?: boolean;                // /docs/{slug}/product.md — see Product Documentation Guide. Optional in every type.
  architecture?: boolean;           // /docs/{slug}/architecture.md
  features?: boolean;               // /docs/{slug}/features.md
  infrastructure?: boolean;         // /docs/{slug}/infrastructure.md
  apiDocs?: boolean;                // /docs/{slug}/api.md — see API Documentation Guide. Optional; relevant for any project that exposes a programmatic interface.
  apiExplorer?: boolean;            // registered in the external API tester. Fully decoupled from apiDocs — one can be true without the other.
  services?: boolean;               // renders nested services list
}
```

> **`apiDocs` vs `apiExplorer`** — these are independent booleans, not a pair. `apiExplorer` means "this project exists in the external Postman-style tester" (can I *run* requests against it?). `apiDocs` means "`api.md` exists" (can I *read* what the contract is?). A project can have either, both, or neither.

```ts
export interface ProjectOverviewData {
  techStack: TechStackModel[];            // max ~6 items, see naming rules below
  highlightedFeatures: HighlightedFeatureModel[]; // 3–4 items
  highlightedCommand?: HighlightCommandModel;      // one curl example
}
```

```ts
export interface TechStackModel { icon: string; label: string; }
export interface HighlightedFeatureModel { icon: string; label: string; description: string; }
export interface HighlightCommandModel { title: string; command: string; }

export interface ProjectMetadataModel {
  repositoryLink: string;
  deploymentLink?: string;
  license: string;
  version: string;
  recentCommits: CommitSummaryModel[];
  metrics: ProjectMetricModel[];
}
export interface ProjectMetricModel { label: string; value: number; }
export interface CommitSummaryModel { hash: string; message: string; date: Date; }
```

---

## Field-by-field rules

### `description`
One or two sentences. State what it **is** and the **core domains** it covers. English. No fluff like "revolutionary" or "blazing fast".

```
Bad:  "A super awesome platform that does everything!!"
Good: "A multi-service e-commerce platform for pharmacy and drugstore domains — identity,
       catalog, carts, orders, payments, inventory — built as a Spring Boot monorepo."
```

### `techStack`
- Use **simple, raw technology names** — no versions, no sub-components.
- Only the **crucial** technologies (max ~6). The consumer already knows Spring Actuator exists if it says Spring Boot.
- Follow this exact style:

```
Good:  Java, Spring Boot, PostgreSQL, Redis, Kafka, Docker
Bad:   Java JVM 25, Spring Boot Actuator 2.2.5, PostgreSQL 15.4.1-3, springdoc-openapi 2.7.0
```

### `highlightedFeatures`
3–4 items max. Pick the things that make this repo worth reading:

| Field        | Value                                                             |
|--------------|-------------------------------------------------------------------|
| `icon`       | Material Symbol name (e.g. `bolt`, `verified_user`, `architecture`) |
| `label`      | Short headline (2–4 words)                                        |
| `description`| One sentence, concrete — name the actual tech that backs it        |

```
{ icon: 'verified_user', label: 'Secure by Default',
  description: 'JWT-based auth with OAuth2, role-scoped endpoints, 2FA, and strict input validation.' }
```

### `highlightedCommand`
One real, copy-pasteable example. Use the real deployment URL and a real endpoint path — **grep the codebase** for the actual route, don't invent it.

### `metadata`

| Field            | How to fill it                                                                  |
|------------------|---------------------------------------------------------------------------------|
| `repositoryLink` | Real GitHub URL. Check `git remote -v`, don't guess the username.               |
| `deploymentLink` | Only if the service is actually deployed. Omit otherwise.                       |
| `license`        | Real license (check for a LICENSE file or a mention in docs).                    |
| `version`        | Highest version tag/commit message, e.g. `2.0.1`.                               |
| `recentCommits`  | Pull from `git log` — real hashes and messages. `date` must be a `Date` object, **not** a string. |
| `metrics`        | 3 numbers that you **counted**, never estimated. See below.                     |

### `metrics`
Always **compute**, never invent:

```bash
# Lines of code
find . -name "*.java" -path "*/src/*" | xargs wc -l | tail -1

# REST endpoints (count mapping annotations)
find . -name "*.java" -path "*/src/*" | xargs grep -o "@\(GetMapping\|PostMapping\|PutMapping\|DeleteMapping\|PatchMapping\)" | wc -l

# Microservices / modules
ls -d *-service | wc -l
```

### `recentCommits`
Always take from real `git log` output:

```bash
git log -5 --pretty=format:'%h|%ad|%s' --date=short
```

Format them as:
```ts
{ hash: '2d69cd79', message: 'version 2.0.1: simplify logging', date: new Date('2026-07-27') },
```

---

## Which flags apply to which `type`

| Flag             | backend | frontend | infra | fullstack |
|------------------|:-------:|:--------:|:-----:|:---------:|
| `product`        | usually present | often | rare/optional | often |
| `architecture`   | yes | yes | yes | yes |
| `features`       | yes | yes | yes | yes |
| `infrastructure` | yes | sometimes | yes | yes |
| `apiDocs`        | yes | if it exposes an API (e.g. BFF) | if it exposes an API (e.g. gateway) | yes |
| `apiExplorer`    | if registered in the tester | rare | rare | if registered in the tester |

None of these are hard requirements — they follow the same general rule as everything else in this documentation set: **only flip a boolean on, and only write the file, if there is real content to put in it.** No filler sections just to complete the table.

---

## Checklist before finishing

- [ ] `description` is English, 1–2 sentences, no fluff.
- [ ] `techStack` has max 6 **raw** tech names, no versions.
- [ ] `metrics` are real numbers from `wc -l` / grep / `ls`.
- [ ] `recentCommits` come from `git log`, dates are `Date` objects.
- [ ] `highlightedCommand` uses a real endpoint from the code.
- [ ] Every `true` flag in `docs` has a matching Markdown file in `/docs/{slug}/`.
- [ ] Every Markdown file that exists has its flag set to `true`.
- [ ] `apiDocs` and `apiExplorer` were set independently, not copy-pasted from each other.
- [ ] `projects.ts` compiles: `npx -p typescript tsc --noEmit --strict --skipLibCheck docs/projects.ts`
- [ ] `services` is `[]` until you actually document the nested services.
