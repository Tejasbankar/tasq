# Tasq

A lightweight task queue system being built in Go to explore backend systems engineering concepts including task orchestration, worker processing, retries, API design, and PostgreSQL-backed persistence.

## Current Scope

- REST API service using Chi
- PostgreSQL integration using pgxpool
- Configuration management
- Health checks with database dependency verification
- Docker-based local development environment

## Tech Stack

- Go
- Chi
- PostgreSQL
- pgxpool
- Docker

## Prerequisites
- Go 1.24+
- Docker
- Goose

## Install Goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Local Setup
1. Start PostgreSQL
```bash
docker compose up -d
```

2. Run Database Migrations
```bash
goose -dir migrations postgres "postgres://tasquser:tasqpassword@localhost:5432/tasq?sslmode=disable" up
```

Check migration status:
```bash
goose -dir migrations postgres "postgres://tasquser:tasqpassword@localhost:5432/tasq?sslmode=disable" status
```

3. Run API Service
```bash
go run ./cmd/api
```

4. Verify Health Endpoint
```bash
curl http://localhost:8080/health
```

Expected response:

{
  "status": "ok"
}
