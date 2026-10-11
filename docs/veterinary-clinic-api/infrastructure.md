# Infrastructure

## Overview

The repository provides a Docker-based development stack with the API, PostgreSQL, and Redis. GitHub Actions builds and tests changes, publishes the container image to GitHub Container Registry on `main`, and deploys the published image to a remote Docker host through Cloudflare Access SSH.

## Deployment Pipeline

### 1. Containerization - Docker

A multi-stage build compiles the API, packages migrations and the migration command, and runs the service in a small Alpine runtime image. The entrypoint waits for PostgreSQL and applies migrations before starting the API unless migration execution is disabled.

**Reference:** [Dockerfile](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/Dockerfile), [entrypoint](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/scripts/entrypoint.sh)

### 2. Local Runtime - Docker Compose

`docker compose up --build` starts the API with PostgreSQL 16 and Redis 7. Health checks gate API startup, and named volumes persist local database and Redis data. The API is exposed on the required `SERVER_PORT`; PostgreSQL and Redis are mapped to host ports 5233 and 6380.

**Reference:** [docker-compose.yml](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/docker-compose.yml)

### 3. Continuous Integration - GitHub Actions

Pull requests and pushes to `main` build the API, run migrations, unit and integration tests, race tests, static analysis, and generated sqlc and Swagger checks.

**Reference:** [CI/CD workflow](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/.github/workflows/ci-cd.yml)

### 4. Image Publishing - GitHub Container Registry

After CI succeeds for a push to `main`, the workflow builds the Docker image and publishes both a `latest` tag and a commit-SHA tag to GHCR.

**Reference:** [publish job](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/.github/workflows/ci-cd.yml)

### 5. Remote Deployment - Cloudflare Access SSH

The deployment job connects to the configured remote host through Cloudflare Access SSH, then pulls the published image and refreshes the remote service using the configured deployment path.

**Reference:** [deploy job](https://github.com/alexisTrejo11/veterinary-clinic-api/blob/main/.github/workflows/ci-cd.yml)

## Infra Stack

- Docker and Docker Compose
- PostgreSQL 16
- Redis 7
- GitHub Actions
- GitHub Container Registry
- Cloudflare Access SSH
- Remote Docker host

## Data Services

PostgreSQL stores clinic data and database-enforced business constraints. Redis stores authentication session and temporary-token state. Connection details are supplied through environment variables; production deployments must provide non-development credentials and secrets.
